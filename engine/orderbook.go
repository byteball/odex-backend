package engine

// The orderbook currently uses the four following data structures to store engine
// state in redis
// 1. Pricepoints set
// 2. Pricepoints volume set
// 3. Pricepoints hashes set
// 4. Orders map

// 1. The pricepoints set is an ordered set that store all pricepoints.
// Keys: ~ pair addresses + side (BUY or SELL)
// Values: pricepoints set (sorted set but all ranks are actually 0)

// 2. The pricepoints volume set is an order set that store the volume for a given pricepoint
// Keys: pair addresses + side + pricepoint
// Values: volume for corresponding (pair, pricepoint)

// 3. The pricepoints hashes set is an ordered set that stores a set of hashes ranked by creation time for a given pricepoint
// Keys: pair addresses + side + pricepoint
// Values: hashes of orders with corresponding pricepoint

// 4. The orders hashmap is a mapping that stores serialized orders
// Keys: hash
// Values: serialized order

import (
	"fmt"
	"math"
	"strconv"

	sync "github.com/sasha-s/go-deadlock"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/types"
)

type OrderBook struct {
	rabbitMQConn  *rabbitmq.Connection
	orderDao      interfaces.OrderDao
	tradeDao      interfaces.TradeDao
	pair          *types.Pair
	mutex         *sync.Mutex
	obyteProvider interfaces.ObyteProvider
	orderService  interfaces.OrderService
}

// newOrder calls buyOrder/sellOrder based on type of order recieved and
// publishes the response back to rabbitmq
func (ob *OrderBook) newOrder(o *types.Order) (err error) {
	// Attain lock on engineResource, so that recovery or cancel order function doesn't interfere
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	res := &types.EngineResponse{}
	if o.Side == "SELL" {
		res, err = ob.sellOrder(o)
		if err != nil {
			logger.Error(err)
			return err
		}

	} else if o.Side == "BUY" {
		res, err = ob.buyOrder(o)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	// Note: Plug the option for orders like FOC, Limit here (if needed)
	err = ob.rabbitMQConn.PublishEngineResponse(res)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (ob *OrderBook) addOrder(o *types.Order) error {
	if o.FilledAmount == 0 {
		o.Status = "OPEN"
	}

	ob.orderService.FixOrderStatus(o)

	_, err := ob.orderDao.FindAndModify(o.Hash, o)
	if err != nil {
		// we add this condition in the case an order is re-run through the orderbook (in case of invalid counterpart order for example)
		logger.Error(err)
		return err
	}

	return nil
}

// buyOrder is triggered when a buy order comes in, it fetches the ask list
// from orderbook. First it checks ths price point list to check whether the order can be matched
// or not, if there are pricepoints that can satisfy the order then corresponding list of orders
// are fetched and trade is executed
func (ob *OrderBook) buyOrder(o *types.Order) (*types.EngineResponse, error) {
	res := &types.EngineResponse{}

	matchingOrders, err := ob.orderDao.GetMatchingSellOrders(o)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// case where no order is matched
	if len(matchingOrders) == 0 || o.MatcherAddress != ob.obyteProvider.GetOperatorAddress() {
		ob.addOrder(o)
		res.Status = "ORDER_ADDED"
		res.Order = o
		return res, nil
	}

	matches := types.Matches{TakerOrder: o}
	for _, mo := range matchingOrders {
		trade, err := ob.execute(o, mo)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		matches.AppendMatch(mo, trade)

		if o.Status == "FILLED" {
			_, err := ob.orderDao.FindAndModify(o.Hash, o)
			if err != nil {
				logger.Error(err)
				return nil, err
			}

			res.Status = "ORDER_FILLED"
			res.Order = o
			res.Matches = &matches
			return res, nil
		}
	}

	// the order can be partial filled and then immediately cancelled
	ob.orderService.FixOrderStatus(o)
	_, err = ob.orderDao.FindAndModify(o.Hash, o)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res.Status = "ORDER_PARTIALLY_FILLED"
	res.Order = o
	res.Matches = &matches
	return res, nil
}

// sellOrder is triggered when a sell order comes in, it fetches the bid list
// from orderbook. First it checks ths price point list to check whether the order can be matched
// or not, if there are pricepoints that can satisfy the order then corresponding list of orders
// are fetched and trade is executed
func (ob *OrderBook) sellOrder(o *types.Order) (*types.EngineResponse, error) {
	res := &types.EngineResponse{}

	matchingOrders, err := ob.orderDao.GetMatchingBuyOrders(o)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(matchingOrders) == 0 || o.MatcherAddress != ob.obyteProvider.GetOperatorAddress() {
		o.Status = "OPEN"
		ob.addOrder(o)

		res.Status = "ORDER_ADDED"
		res.Order = o
		return res, nil
	}

	matches := types.Matches{TakerOrder: o}
	for _, mo := range matchingOrders {
		trade, err := ob.execute(o, mo)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		matches.AppendMatch(mo, trade)

		if o.Status == "FILLED" {
			_, err := ob.orderDao.FindAndModify(o.Hash, o)
			if err != nil {
				logger.Error(err)
				return nil, err
			}

			res.Status = "ORDER_FILLED"
			res.Order = o
			res.Matches = &matches
			return res, nil
		}
	}

	// the order can be partial filled and then immediately cancelled
	ob.orderService.FixOrderStatus(o)
	_, err = ob.orderDao.FindAndModify(o.Hash, o)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res.Status = "ORDER_PARTIALLY_FILLED"
	res.Order = o
	res.Matches = &matches
	return res, nil
}

// execute function is responsible for executing of matched orders
// i.e it deletes/updates orders in case of order matching and responds
// with trade instance and fillOrder
func (ob *OrderBook) execute(takerOrder *types.Order, makerOrder *types.Order) (*types.Trade, error) {
	trade := &types.Trade{}
	tradeAmount := int64(0)      // always in base currency
	tradeQuoteAmount := int64(0) // always in quote currency

	//TODO changes 'strictly greater than' condition. The orders that are almost completely filled
	//TODO should be removed/skipped
	if takerOrder.Side == "BUY" {
		// sell the remaining taker's quote amount at maker's price
		// takerOutput in base currency
		//takerOutput := round(float64(takerOrder.RemainingSellAmount) / toOscriptPrecision(makerOrder.Price))
		makerQuoteOutput := round(float64(makerOrder.RemainingSellAmount) * toOscriptPrecision(makerOrder.OriginalPrice()))
		//if makerOrder.RemainingSellAmount > takerOutput {
		if makerQuoteOutput > takerOrder.RemainingSellAmount {
			tradeAmount = round(float64(takerOrder.RemainingSellAmount) / toOscriptPrecision(makerOrder.OriginalPrice()))
			tradeQuoteAmount = takerOrder.RemainingSellAmount

			makerOrder.FilledAmount += tradeAmount
			makerOrder.RemainingSellAmount -= tradeAmount

			takerOrder.FilledAmount += tradeAmount
			takerOrder.RemainingSellAmount = 0

			makerOrder.Status = "PARTIAL_FILLED"
			if makerOrder.RemainingSellAmount == 0 {
				makerOrder.Status = "FILLED"
			}
			takerOrder.Status = "FILLED"
		} else { // maker <= taker
			tradeAmount = makerOrder.RemainingAmount()
			tradeQuoteAmount = round(float64(tradeAmount) * toOscriptPrecision(makerOrder.OriginalPrice()))

			makerOrder.FilledAmount += tradeAmount
			makerOrder.RemainingSellAmount -= tradeAmount
			if makerOrder.RemainingSellAmount != 0 {
				panic(fmt.Sprintf("smaller maker seller: remaining sell amount = %d", makerOrder.RemainingSellAmount))
			}

			takerOrder.FilledAmount += tradeAmount
			takerOrder.RemainingSellAmount -= tradeQuoteAmount

			makerOrder.Status = "FILLED"
			if takerOrder.RemainingSellAmount > 0 {
				takerOrder.Status = "PARTIAL_FILLED"
			} else if takerOrder.RemainingSellAmount == 0 {
				takerOrder.Status = "FILLED"
			} else {
				panic(fmt.Sprintf("takerOrder.RemainingSellAmount = %d", takerOrder.RemainingSellAmount))
			}
		}
	} else { // taker is seller
		makerOutput := round(float64(makerOrder.RemainingSellAmount) * toOscriptPrecision(makerOrder.OriginalPrice()))
		if makerOutput > takerOrder.RemainingAmount() {
			tradeAmount = takerOrder.RemainingAmount()
			tradeQuoteAmount = round(float64(tradeAmount) / toOscriptPrecision(makerOrder.OriginalPrice()))

			makerOrder.FilledAmount += tradeAmount
			makerOrder.RemainingSellAmount -= tradeQuoteAmount

			takerOrder.FilledAmount += tradeAmount
			takerOrder.RemainingSellAmount -= tradeAmount
			if takerOrder.RemainingSellAmount != 0 {
				panic(fmt.Sprintf("smaller taker seller: remaining sell amount = %d", takerOrder.RemainingSellAmount))
			}

			makerOrder.Status = "PARTIAL_FILLED"
			if makerOrder.RemainingSellAmount == 0 {
				makerOrder.Status = "FILLED"
			}
			takerOrder.Status = "FILLED"
		} else { // maker <= taker
			tradeAmount = makerOutput
			tradeQuoteAmount = makerOrder.RemainingSellAmount

			makerOrder.FilledAmount += tradeAmount
			makerOrder.RemainingSellAmount = 0

			takerOrder.FilledAmount += tradeAmount
			takerOrder.RemainingSellAmount -= tradeAmount

			makerOrder.Status = "FILLED"
			if takerOrder.RemainingSellAmount > 0 {
				takerOrder.Status = "PARTIAL_FILLED"
			} else if takerOrder.RemainingSellAmount == 0 {
				takerOrder.Status = "FILLED"
			} else {
				panic(fmt.Sprintf("takerOrder.RemainingSellAmount = %d", takerOrder.RemainingSellAmount))
			}
		}
	}

	_, err := ob.orderDao.FindAndModify(makerOrder.Hash, makerOrder)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	trade = &types.Trade{
		Amount:                   tradeAmount,
		QuoteAmount:              tradeQuoteAmount,
		Price:                    makerOrder.Price, // maker price!
		BaseToken:                takerOrder.BaseToken,
		QuoteToken:               takerOrder.QuoteToken,
		MakerOrderHash:           makerOrder.Hash,
		TakerOrderHash:           takerOrder.Hash,
		Taker:                    takerOrder.UserAddress,
		PairName:                 takerOrder.PairName,
		Maker:                    makerOrder.UserAddress,
		RemainingTakerSellAmount: takerOrder.RemainingSellAmount,
		RemainingMakerSellAmount: makerOrder.RemainingSellAmount,
		Status:                   "PENDING",
		MakerSide:                makerOrder.Side,
	}

	trade.Hash = trade.ComputeHash()
	return trade, nil
}

func toOscriptPrecision(x float64) float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%.15g", x), 64)
	if err != nil {
		panic(err)
	}
	return f
}

func round(x float64) int64 {
	return int64(math.RoundToEven(toOscriptPrecision(x)))
}

// CancelOrder is used to cancel the order from orderbook
func (ob *OrderBook) cancelOrder(o *types.Order) error {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	if o.Status != "AUTO_CANCELLED" && o.Status != "FILLED" {
		o.Status = "CANCELLED"
	}
	if o.Status == "AUTO_CANCELLED" || o.Status == "CANCELLED" {
		err := ob.orderDao.UpdateOrderStatus(o.Hash, o.Status)
		if err != nil {
			logger.Error(err, "when cancelling order", o.Hash)
			return err
		}
	}

	// todo: another engine response when the order was already cancelled or filled
	res := &types.EngineResponse{
		Status:  "ORDER_CANCELLED",
		Order:   o,
		Matches: nil,
	}

	err := ob.rabbitMQConn.PublishEngineResponse(res)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// cancelTrades revertTrades and reintroduces the taker orders in the orderbook
func (ob *OrderBook) invalidateMakerOrders(matches types.Matches) error {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	orders := matches.MakerOrders
	trades := matches.Trades
	//tradeAmounts := matches.TradeAmounts()
	makerOrderHashes := []string{}
	//takerOrderHashes := []string{}
	tradeAmount := int64(0)

	for i, _ := range orders {
		makerOrderHashes = append(makerOrderHashes, trades[i].MakerOrderHash)
		//takerOrderHashes = append(takerOrderHashes, trades[i].TakerOrderHash)
		tradeAmount += trades[i].Amount
	}

	takerOrders, err := ob.orderDao.UpdateOrderFilledAmounts([]string{matches.TakerOrder.Hash}, []int64{tradeAmount})
	if err != nil {
		logger.Error(err)
		return err
	}

	makerOrders, err := ob.orderDao.UpdateOrderStatusesByHashes("INVALIDATED", makerOrderHashes...)
	if err != nil {
		logger.Error(err)
		return err
	}

	//TODO in the case the trades are not in the database they should not be created.
	cancelledTrades, err := ob.tradeDao.UpdateTradeStatusesByOrderHashes("CANCELLED", makerOrderHashes...)
	if err != nil {
		logger.Error(err)
		return err
	}

	res := &types.EngineResponse{
		Status:            "TRADES_CANCELLED",
		InvalidatedOrders: &makerOrders,
		CancelledTrades:   &cancelledTrades,
	}

	err = ob.rabbitMQConn.PublishEngineResponse(res)
	if err != nil {
		logger.Error(err)
	}

	for _, o := range takerOrders {
		err := ob.rabbitMQConn.PublishNewOrderMessage(o)
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}

/*func (ob *OrderBook) invalidateTakerOrders(matches types.Matches) error {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	makerOrders := matches.MakerOrders
	takerOrder := matches.TakerOrder
	trades := matches.Trades
	tradeAmounts := matches.TradeAmounts()

	makerOrderHashes := []string{}
	for i, _ := range trades {
		makerOrderHashes = append(makerOrderHashes, trades[i].MakerOrderHash)
	}

	makerOrders, err := ob.orderDao.UpdateOrderFilledAmounts(makerOrderHashes, tradeAmounts)
	if err != nil {
		logger.Error(err)
		return err
	}

	invalidatedOrders, err := ob.orderDao.UpdateOrderStatusesByHashes("INVALIDATED", takerOrder.Hash)
	if err != nil {
		logger.Error(err)
		return err
	}

	cancelledTrades, err := ob.tradeDao.UpdateTradeStatusesByOrderHashes("CANCELLED", makerOrderHashes...)
	if err != nil {
		logger.Error(err)
		return err
	}

	res := &types.EngineResponse{
		Status:            "TRADES_CANCELLED",
		InvalidatedOrders: &invalidatedOrders,
		CancelledTrades:   &cancelledTrades,
	}

	err = ob.rabbitMQConn.PublishEngineResponse(res)
	if err != nil {
		logger.Error(err)
		return err
	}

	for _, o := range makerOrders {
		err := ob.rabbitMQConn.PublishNewOrderMessage(o)
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}*/

/*func (ob *OrderBook) InvalidateOrder(o *types.Order) (*types.EngineResponse, error) {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	o.Status = "ERROR"
	err := ob.orderDao.UpdateOrderStatus(o.Hash, "ERROR")
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res := &types.EngineResponse{
		Status:  "INVALIDATED",
		Order:   o,
		Matches: nil,
	}

	return res, nil
}*/
