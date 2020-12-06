package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	sync "github.com/sasha-s/go-deadlock"

	"github.com/spf13/cast"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/utils"
	"github.com/byteball/odex-backend/ws"

	"github.com/globalsign/mgo/bson"

	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/types"
)

// OrderService
type OrderService struct {
	orderDao            interfaces.OrderDao
	pairDao             interfaces.PairDao
	accountDao          interfaces.AccountDao
	tradeDao            interfaces.TradeDao
	validator           interfaces.ValidatorService
	broker              *rabbitmq.Connection
	orderChannels       map[string]chan *types.WebsocketEvent
	ordersInThePipeline map[string]*types.Order
	mu                  sync.Mutex
}

// NewOrderService returns a new instance of orderservice
func NewOrderService(
	orderDao interfaces.OrderDao,
	pairDao interfaces.PairDao,
	accountDao interfaces.AccountDao,
	tradeDao interfaces.TradeDao,
	validator interfaces.ValidatorService,
	broker *rabbitmq.Connection,
) *OrderService {

	orderChannels := make(map[string]chan *types.WebsocketEvent)
	ordersInThePipeline := make(map[string]*types.Order)

	s := &OrderService{
		orderDao,
		pairDao,
		accountDao,
		tradeDao,
		validator,
		broker,
		orderChannels,
		ordersInThePipeline,
		sync.Mutex{},
	}
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.CancelExpiredOrders()
		}
	}()
	return s
}

// GetByID fetches the details of an order using order's mongo ID
func (s *OrderService) GetByID(id bson.ObjectId) (*types.Order, error) {
	return s.orderDao.GetByID(id)
}

// GetByUserAddress fetches all the orders placed by passed user address
func (s *OrderService) GetByUserAddress(addr string, limit ...int) ([]*types.Order, error) {
	return s.orderDao.GetByUserAddress(addr, limit...)
}

// GetByHash fetches all trades corresponding to a trade hash
func (s *OrderService) GetByHash(hash string) (*types.Order, error) {
	return s.orderDao.GetByHash(hash)
}

func (s *OrderService) GetByHashes(hashes []string) ([]*types.Order, error) {
	return s.orderDao.GetByHashes(hashes)
}

// GetCurrentByUserAddress function fetches list of open/partial orders from order collection based on user address.
// Returns array of Order type struct
func (s *OrderService) GetCurrentByUserAddress(addr string, limit ...int) ([]*types.Order, error) {
	return s.orderDao.GetCurrentByUserAddress(addr, limit...)
}

// GetHistoryByUserAddress function fetches list of orders which are not in open/partial order status
// from order collection based on user address.
// Returns array of Order type struct
func (s *OrderService) GetHistoryByUserAddress(addr string, limit ...int) ([]*types.Order, error) {
	return s.orderDao.GetHistoryByUserAddress(addr, limit...)
}

// NewOrder validates if the passed order is valid or not based on user's available
// funds and order data.
// If valid: Order is inserted in DB with order status as new and order is publiched
// on rabbitmq queue for matching engine to process the order
func (s *OrderService) NewOrder(o *types.Order) (e error) {
	s.mu.Lock()
	existingOrder := s.ordersInThePipeline[o.Hash]
	if existingOrder == nil {
		s.ordersInThePipeline[o.Hash] = o
	}
	s.mu.Unlock()
	if existingOrder != nil {
		logger.Info("duplicate order found in memory", o.Hash)
		return nil // not an error
	}
	defer func() {
		if e != nil {
			s.mu.Lock()
			delete(s.ordersInThePipeline, o.Hash)
			s.mu.Unlock()
		}
	}()

	existingOrder, daoErr := s.orderDao.GetByHash(o.Hash)
	if daoErr != nil {
		logger.Error(daoErr)
		return daoErr
	}
	if existingOrder != nil {
		logger.Info("duplicate order found in db", o.Hash)
		return nil // not an error
	}

	if err := o.Validate(); err != nil {
		logger.Error(err)
		return err
	}

	if err := s.validator.ValidateOperatorAddress(o); err != nil {
		logger.Error(err)
		return err
	}

	/*id, err := s.validator.VerifySignature(o)
	if err != nil {
		logger.Error(err)
		return err
	}
	o.Hash = id*/

	logger.Info("GetByAsset", o.BaseToken, o.QuoteToken)
	p, err := s.pairDao.GetByAsset(o.BaseToken, o.QuoteToken)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("pair", p)

	if p == nil {
		return errors.New("Pair not found")
	}

	//if o.QuoteAmount(p) < p.MinQuoteAmount() {
	//	return errors.New("Order amount too low")
	//}
	//logger.Info("126")

	// Fill token and pair data
	err = o.Process(p)
	if err != nil {
		logger.Error(err)
		return err
	}
	//logger.Info("filled pair", o.Pair)

	balanceLockedInMemoryOrders := int64(0)
	s.mu.Lock()
	for _, po := range s.ordersInThePipeline {
		if po.UserAddress == o.UserAddress && po.Hash != o.Hash && po.SellToken() == o.SellToken() && po.Status == "OPEN" {
			balanceLockedInMemoryOrders += po.RemainingSellAmount
		}
	}
	s.mu.Unlock()

	deltas := s.AdjustBalancesForUncommittedTrades(o.UserAddress, map[string]int64{})
	err = s.validator.ValidateAvailableBalance(o, deltas, balanceLockedInMemoryOrders)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = s.broker.PublishNewOrderMessage(o)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// CancelOrder handles the cancellation order requests.
// Only Orders which are OPEN or NEW i.e. Not yet filled/partially filled
// can be cancelled
func (s *OrderService) CancelOrder(oc *types.OrderCancel) error {
	s.mu.Lock()
	memoryOrder := s.ordersInThePipeline[oc.OrderHash]
	if memoryOrder != nil {
		memoryOrder.Status = "CANCELLED"
		logger.Info("in-memory order status set to CANCELLED")
	}
	s.mu.Unlock()

	o, err := s.orderDao.GetByHash(oc.OrderHash)
	if err != nil {
		logger.Error(err)
		return err
	}

	/*_, err = s.validator.VerifyCancelSignature(oc)
	if err != nil {
		logger.Error(err)
		return err
	}*/

	foundInDb := o != nil
	if o == nil {
		if memoryOrder == nil {
			return errors.New("No order with corresponding hash: " + oc.OrderHash)
		} else {
			o = memoryOrder
			logger.Info("to-be-cancelled order " + oc.OrderHash + " found in memory")
		}
	}

	if o.Status == "FILLED" || o.Status == "ERROR" || foundInDb && o.Status == "CANCELLED" {
		return fmt.Errorf("Cannot cancel order %v. Status is %v", o.Hash, o.Status)
	}

	// update order status early to make sure new orders see the freed-up balance.
	// The status will be updated again after going through rabbitmq
	if foundInDb && o.Status != "CANCELLED" && o.Status != "AUTO_CANCELLED" && o.Status != "FILLED" {
		err := s.orderDao.UpdateOrderStatus(o.Hash, "CANCELLED")
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	err = s.broker.PublishCancelOrderMessage(o)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (s *OrderService) GetSenderAddresses(oc *types.OrderCancel) (string, string, error) {
	/*addr, err := s.validator.VerifyCancelSignature(oc)
	if err != nil {
		logger.Error(err)
		return "", err
	}*/
	s.mu.Lock()
	o := s.ordersInThePipeline[oc.OrderHash]
	s.mu.Unlock()

	var err error
	if o == nil {
		o, err = s.orderDao.GetByHash(oc.OrderHash)
	}

	if err == nil && o == nil {
		err = errors.New("failed to find the order to be cancelled: " + oc.OrderHash)
	}

	if err != nil {
		logger.Error(err)
		return "", "", err
	}

	authors := cast.ToSlice(o.OriginalOrder["authors"])
	author := cast.ToStringMapString(authors[0])
	signer := author["address"]

	return o.UserAddress, signer, nil
}

func (s *OrderService) CheckIfBalancesAreSufficientAndCancel(address string, balances map[string]int64) {
	for token, balance := range balances {
		s.checkIfBalanceIsSufficientAndCancel(address, token, balance)
	}
}

func (s *OrderService) checkIfBalanceIsSufficientAndCancel(address string, token string, balance int64) {
	lockedBalance, _, err := s.orderDao.GetUserLockedBalance(address, token)
	if err != nil {
		panic(err)
	}
	if lockedBalance <= balance {
		return
	}
	orders, err := s.orderDao.GetCurrentByUserAddress(address)
	if err != nil {
		panic(err)
	}
	bCancelAll := (token == "base" && balance < int64(len(orders)*1000)) // not enough for fees
	for _, order := range orders {
		if !bCancelAll {
			var soldToken string
			if order.Side == "SELL" {
				soldToken = order.BaseToken
			} else {
				soldToken = order.QuoteToken
			}
			if soldToken != token {
				continue
			}
		}
		order.Status = "AUTO_CANCELLED"
		err = s.broker.PublishCancelOrderMessage(order)
		if err != nil {
			panic(err)
		}
		if !bCancelAll {
			lockedBalance -= order.RemainingSellAmount
			if lockedBalance <= balance { // enough orders cancelled
				return
			}
		}
	}
}

func (s *OrderService) CancelOrdersSignedByRevokedSigner(address string, signer string) {
	orders, err := s.orderDao.GetCurrentByUserAddressAndSignerAddress(address, signer)
	if err != nil {
		panic(err)
	}
	logger.Info("will cancel", len(orders), "orders due to revocation of authorization on", address, "from signer", signer)
	for _, order := range orders {
		order.Status = "AUTO_CANCELLED"
		err = s.broker.PublishCancelOrderMessage(order)
		if err != nil {
			panic(err)
		}
	}
}

func (s *OrderService) CancelExpiredOrders() {
	orders, err := s.orderDao.GetExpiredOrders()
	if err != nil {
		panic(err)
	}
	logger.Info("will cancel", len(orders), "expired orders")
	for _, order := range orders {
		order.Status = "AUTO_CANCELLED"
		err = s.broker.PublishCancelOrderMessage(order)
		if err != nil {
			panic(err)
		}
	}
}

func (s *OrderService) handleOrderCancelled(res *types.EngineResponse) {
	go ws.SendOrderMessage("ORDER_CANCELLED", res.Order.UserAddress, res.Order)
	s.broadcastOrderBookUpdate([]*types.Order{res.Order})
	s.broadcastRawOrderBookUpdate([]*types.Order{res.Order})
	return
}

// HandleEngineResponse listens to messages incoming from the engine and handles websocket
// responses and database updates accordingly
func (s *OrderService) HandleEngineResponse(res *types.EngineResponse) error {
	s.mu.Lock()
	memoryOrder := s.ordersInThePipeline[res.Order.Hash]
	if memoryOrder != nil && memoryOrder.Status == "CANCELLED" {
		logger.Info("EngineResponse: status of memory order " + res.Order.Hash + " is CANCELLED")
		if res.Order.Status != "CANCELLED" {
			res.Order.Status = "CANCELLED"
			logger.Info("EngineResponse: fixed status of order " + res.Order.Hash + " to CANCELLED")
		}
	}
	delete(s.ordersInThePipeline, res.Order.Hash)
	s.mu.Unlock()
	switch res.Status {
	case "ERROR":
		s.handleEngineError(res)
	case "ORDER_ADDED":
		s.handleEngineOrderAdded(res)
	case "ORDER_FILLED":
		s.handleEngineOrderMatched(res)
	case "ORDER_PARTIALLY_FILLED":
		s.handleEngineOrderMatched(res)
	case "ORDER_CANCELLED":
		s.handleOrderCancelled(res)
	case "TRADES_CANCELLED":
		s.handleOrdersInvalidated(res)
	default:
		s.handleEngineUnknownMessage(res)
	}

	return nil
}

func (s *OrderService) HandleOperatorMessages(msg *types.OperatorMessage) error {
	switch msg.MessageType {
	case "TRADE_ERROR":
		s.handleOperatorTradeError(msg)
	case "TRADE_TX_PENDING":
		s.handleOperatorTradeTxPending(msg)
	case "TRADE_TX_SUCCESS":
		s.handleOperatorTradeTxSuccess(msg)
	case "TRADE_TX_ERROR":
		s.handleOperatorTradeTxError(msg)
	//case "TRADE_INVALID":
	//	s.handleOperatorTradeInvalid(msg)
	default:
		s.handleOperatorUnknownMessage(msg)
	}

	return nil
}

func (s *OrderService) handleOrdersInvalidated(res *types.EngineResponse) error {
	orders := res.InvalidatedOrders
	trades := res.CancelledTrades

	for _, o := range *orders {
		go ws.SendOrderMessage("ORDER_INVALIDATED", o.UserAddress, o)
	}

	if orders != nil && len(*orders) != 0 {
		s.broadcastOrderBookUpdate(*orders)
	}

	if orders != nil && len(*orders) != 0 {
		s.broadcastRawOrderBookUpdate(*orders)
	}

	if trades != nil && len(*trades) != 0 {
		s.broadcastTradeUpdate(*trades)
	}

	return nil
}

// handleEngineError returns an websocket error message to the client
func (s *OrderService) handleEngineError(res *types.EngineResponse) {
	o := res.Order
	go ws.SendOrderMessage("ERROR", o.UserAddress, nil)
}

// handleEngineOrderAdded returns a websocket message informing the client that his order has been added
// to the orderbook (but currently not matched)
func (s *OrderService) handleEngineOrderAdded(res *types.EngineResponse) {
	o := res.Order
	go ws.SendOrderMessage("ORDER_ADDED", o.UserAddress, o)

	s.broadcastOrderBookUpdate([]*types.Order{o})
	s.broadcastRawOrderBookUpdate([]*types.Order{o})
}

// handleEngineOrderMatched returns a websocket message informing the client that his order has been added.
func (s *OrderService) handleEngineOrderMatched(res *types.EngineResponse) {
	o := res.Order //res.Order is the "taker" order
	taker := o.UserAddress
	matches := *res.Matches

	orders := []*types.Order{o}
	validMatches := types.Matches{TakerOrder: o}
	invalidMatches := types.Matches{TakerOrder: o}

	//res.Matches is an array of (order, trade) pairs where each order is an "maker" order that is being matched
	for i, _ := range matches.Trades {
		//err := s.validator.ValidateBalance(matches.MakerOrders[i])
		var err error
		if err != nil {
			logger.Error(err)
			invalidMatches.AppendMatch(matches.MakerOrders[i], matches.Trades[i])

		} else {
			validMatches.AppendMatch(matches.MakerOrders[i], matches.Trades[i])
			orders = append(orders, matches.MakerOrders[i])
		}
	}

	// if there are any invalid matches, the maker orders are at cause (since taker orders have been validated in the
	// newOrder() function. We remove the maker orders from the orderbook)
	/*if invalidMatches.Length() > 0 {
		err := s.broker.PublishInvalidateMakerOrdersMessage(invalidMatches)
		if err != nil {
			logger.Error(err)
		}
	}*/

	if validMatches.Length() > 0 {
		err := s.tradeDao.Create(validMatches.Trades...)
		if err != nil {
			logger.Error(err)
			go ws.SendOrderMessage("ERROR", taker, err)
			return
		}

		err = s.broker.PublishTrades(&validMatches)
		if err != nil {
			logger.Error(err)
			go ws.SendOrderMessage("ERROR", taker, err)
			return
		}

		for _, o2 := range orders {
			go ws.SendOrderMessage("ORDER_MATCHED", o2.UserAddress, types.OrderMatchedPayload{Matches: &matches})
		}
	}

	// we only update the orderbook with the current set of orders if there are no invalid matches.
	// If there are invalid matches, the corresponding maker orders will be removed and the taker order
	// amount filled will be updated as a result, and therefore does not represent the current state of the orderbook
	if invalidMatches.Length() == 0 {
		s.broadcastOrderBookUpdate(orders)
		s.broadcastRawOrderBookUpdate(orders)
	}
}

func (s *OrderService) handleEngineUnknownMessage(res *types.EngineResponse) {
	log.Print("Receiving unknown engine message")
	utils.PrintJSON(res)
}

func (s *OrderService) handleOperatorUnknownMessage(msg *types.OperatorMessage) {
	log.Print("Receiving unknown message")
	utils.PrintJSON(msg)
}

// handleOperatorTradeTxPending notifies clients a trade tx was successfully sent to the Obyte chain
// and is currently pending
func (s *OrderService) handleOperatorTradeTxPending(msg *types.OperatorMessage) {
	matches := msg.Matches
	trades := matches.Trades
	orders := matches.MakerOrders

	taker := trades[0].Taker
	go ws.SendOrderMessage("ORDER_PENDING", taker, types.OrderPendingPayload{Matches: matches})

	for _, o := range orders {
		maker := o.UserAddress
		go ws.SendOrderMessage("ORDER_PENDING", maker, types.OrderPendingPayload{Matches: matches})
	}

	s.broadcastTradeUpdate(trades)
}

// handleOperatorTradeTxSuccess handles successfull trade messages from the orderbook. It updates
// the trade status in the database and
func (s *OrderService) handleOperatorTradeTxSuccess(msg *types.OperatorMessage) {
	matches := msg.Matches
	hashes := []string{}
	trades := matches.Trades

	for _, t := range trades {
		hashes = append(hashes, t.Hash)
	}

	if len(hashes) == 0 {
		return
	}

	trades, err := s.tradeDao.UpdateTradeStatuses("SUCCESS", hashes...)
	if err != nil {
		logger.Error(err)
	}

	// Send ORDER_SUCCESS message to order takers
	taker := trades[0].Taker
	go ws.SendOrderMessage("ORDER_SUCCESS", taker, types.OrderSuccessPayload{Matches: matches})

	// Send ORDER_SUCCESS message to order makers
	for i, _ := range trades {
		match := matches.NthMatch(i)
		maker := match.MakerOrders[0].UserAddress
		go ws.SendOrderMessage("ORDER_SUCCESS", maker, types.OrderSuccessPayload{Matches: match})
	}

	s.broadcastTradeUpdate(trades)
}

// handleOperatorTradeTxError handles cases where a blockchain transaction is reverted
func (s *OrderService) handleOperatorTradeTxError(msg *types.OperatorMessage) {
	matches := msg.Matches
	trades := matches.Trades
	orders := matches.MakerOrders

	errType := msg.ErrorType
	if errType != "" {
		logger.Error("")
	}

	for _, t := range trades {
		err := s.tradeDao.UpdateTradeStatus(t.Hash, "REJECTED")
		if err != nil {
			logger.Error(err)
		}

		t.Status = "ERROR"
	}

	taker := trades[0].Taker
	go ws.SendOrderMessage("ORDER_ERROR", taker, matches.TakerOrder)
	// cancel the remaining order, if any
	if matches.TakerOrder.Status == "PARTIAL_FILLED" {
		s.orderDao.UpdateOrderStatus(taker, "AUTO_CANCELLED")
	}

	for _, o := range orders {
		maker := o.UserAddress
		go ws.SendOrderMessage("ORDER_ERROR", maker, o)
		// cancel the remaining order, if any
		if o.Status == "PARTIAL_FILLED" {
			s.orderDao.UpdateOrderStatus(maker, "AUTO_CANCELLED")
		}
	}

	s.broadcastTradeUpdate(trades)
}

// handleOperatorTradeError handles cases where the operator encountered a server error (not due to an invalid order
// or a blockchain error)
func (s *OrderService) handleOperatorTradeError(msg *types.OperatorMessage) {
	matches := msg.Matches
	trades := matches.Trades
	orders := matches.MakerOrders

	errType := msg.ErrorType
	if errType != "" {
		logger.Error("")
	}

	for _, t := range trades {
		err := s.tradeDao.UpdateTradeStatus(t.Hash, "ERROR")
		if err != nil {
			logger.Error(err)
		}

		t.Status = "ERROR"
	}

	taker := trades[0].Taker
	go ws.SendOrderMessage("ORDER_ERROR", taker, matches.TakerOrder)

	for _, o := range orders {
		maker := o.UserAddress
		go ws.SendOrderMessage("ORDER_ERROR", maker, o)
	}

	s.broadcastTradeUpdate(trades)
}

// handleOperatorTradeInvalid handles the case where one of the two orders is invalid
// which can be the case for example if one of the account addresses does suddendly
// not have enough tokens to satisfy the order. Ultimately, the goal would be to
// reinclude the non-invalid orders in the orderbook
/*func (s *OrderService) handleOperatorTradeInvalid(msg *types.OperatorMessage) {
	matches := msg.Matches
	trades := matches.Trades
	orders := matches.MakerOrders

	errType := msg.ErrorType
	if errType != "" {
		logger.Error("")
	}

	for _, t := range trades {
		err := s.tradeDao.UpdateTradeStatus(t.Hash, "ERROR")
		if err != nil {
			logger.Error(err)
		}

		t.Status = "ERROR"
	}

	taker := trades[0].Taker
	go ws.SendOrderMessage("ORDER_ERROR", taker, matches.TakerOrder)

	for _, o := range orders {
		maker := o.UserAddress
		go ws.SendOrderMessage("ORDER_ERROR", maker, o)
	}

	s.broadcastTradeUpdate(trades)
}*/

func (s *OrderService) broadcastOrderBookUpdate(orders []*types.Order) {
	bids := []map[string]interface{}{}
	asks := []map[string]interface{}{}

	p, err := orders[0].Pair()
	if err != nil {
		logger.Error()
		return
	}

	for _, o := range orders {
		pp := o.Price
		side := o.Side

		amount, matcherAddress, matcherFeeRate, err := s.orderDao.GetOrderBookPrice(p, pp, side)
		if err != nil {
			logger.Error(err)
		}

		// case where the amount at the pricepoint is equal to 0
		if amount == 0 {
			amount = 0
		}

		update := map[string]interface{}{
			"price":          pp,
			"amount":         amount,
			"matcherAddress": matcherAddress,
			"matcherFeeRate": matcherFeeRate,
		}

		if side == "BUY" {
			bids = append(bids, update)
		} else {
			asks = append(asks, update)
		}
	}

	logger.Info("broadcastOrderBookUpdate", bids, asks)
	id := utils.GetOrderBookChannelID(p.BaseAsset, p.QuoteAsset)
	go ws.GetOrderBookSocket().BroadcastMessage(id, map[string]interface{}{
		"pairName": orders[0].PairName,
		"bids":     bids,
		"asks":     asks,
	})
}

func (s *OrderService) broadcastRawOrderBookUpdate(orders []*types.Order) {
	p, err := orders[0].Pair()
	if err != nil {
		logger.Error(err)
		return
	}

	id := utils.GetOrderBookChannelID(p.BaseAsset, p.QuoteAsset)
	go ws.GetRawOrderBookSocket().BroadcastMessage(id, orders)
}

func (s *OrderService) broadcastTradeUpdate(trades []*types.Trade) {
	p, err := trades[0].Pair()
	if err != nil {
		logger.Error(err)
		return
	}

	id := utils.GetTradeChannelID(p.BaseAsset, p.QuoteAsset)
	go ws.GetTradeSocket().BroadcastMessage(id, trades)
}

func (s *OrderService) AdjustBalancesForUncommittedTrades(address string, balances map[string]int64) map[string]int64 {
	//logger.Info("balances in:", balances)
	trades := s.tradeDao.GetUncommittedTradesByUserAddress(address)

	for _, t := range trades {
		baseSymbol := t.PairName[:strings.IndexByte(t.PairName, '/')]
		quoteSymbol := t.PairName[strings.IndexByte(t.PairName, '/')+1:]
		var myOrderHash string
		if t.Taker == address {
			myOrderHash = t.TakerOrderHash
		} else {
			myOrderHash = t.MakerOrderHash
		}
		o, err := s.orderDao.GetByHash(myOrderHash)
		if err != nil {
			panic(err)
		}
		if o == nil {
			logger.Error("order not found for trade " + t.Hash + " address " + address)
			continue
		}
		baseTokenAmount := t.Amount
		quoteTokenAmount := int64(math.Round(float64(baseTokenAmount) * t.Price))
		var sellAmount, buyAmount int64
		var sellSymbol, buySymbol string
		if o.Side == "SELL" {
			sellAmount = baseTokenAmount
			buyAmount = quoteTokenAmount
			sellSymbol = baseSymbol
			buySymbol = quoteSymbol
		} else {
			sellAmount = quoteTokenAmount
			buyAmount = baseTokenAmount
			sellSymbol = quoteSymbol
			buySymbol = baseSymbol
		}
		logger.Info("adjusted balance of", address, "by -", sellAmount, sellSymbol, ", +", buyAmount, buySymbol)
		balances[sellSymbol] -= sellAmount
		balances[buySymbol] += buyAmount
	}

	return balances
}

func (s *OrderService) FixOrderStatus(o *types.Order) {
	s.mu.Lock()
	memoryOrder := s.ordersInThePipeline[o.Hash]
	if memoryOrder != nil && memoryOrder.Status == "CANCELLED" {
		o.Status = "CANCELLED"
		logger.Info("set status of order " + o.Hash + " to CANCELLED")
	}
	s.mu.Unlock()
}
