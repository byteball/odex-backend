package operator

import (
	"encoding/json"
	"errors"
	sync "github.com/sasha-s/go-deadlock"

	"github.com/spf13/cast"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
	"github.com/byteball/odex-backend/ws"
)

var logger = utils.Logger

// Operator manages the transaction queue that will eventually be
// sent to the exchange AA. The Operator Wallet must be equal to the matcher of submitted orders
type Operator struct {
	AccountService    interfaces.AccountService
	TradeService      interfaces.TradeService
	OrderService      interfaces.OrderService
	ObyteProvider     interfaces.ObyteProvider
	TxQueues          []*TxQueue
	QueueAddressIndex map[string]*TxQueue
	Broker            *rabbitmq.Connection
	mutex             *sync.Mutex
}

type OperatorInterface interface {
	QueueTrade(o *types.Order, t *types.Trade) error
	GetShortestQueue() (*TxQueue, int, error)
}

// NewOperator creates a new operator struct.
// The error and trade events are received in the ErrorChannel and TradeChannel.
// Upon receiving errors and trades in their respective channels, event payloads are sent to the
// associated order maker and taker sockets through the through the event channel on the Order and Trade struct.
// In addition, an error event cancels the trade in the trading engine and makes the order available again.
func NewOperator(
	tradeService interfaces.TradeService,
	orderService interfaces.OrderService,
	accountService interfaces.AccountService,
	provider interfaces.ObyteProvider,
	conn *rabbitmq.Connection,
) (*Operator, error) {
	txqueues := []*TxQueue{}
	addressIndex := make(map[string]*TxQueue)

	name := "oper"
	ch := conn.GetChannel("TX_QUEUES:" + name)

	err := conn.DeclareThrottledQueue(ch, "TX_QUEUES:"+name)
	if err != nil {
		panic(err)
	}

	txq, err := NewTxQueue(
		name,
		tradeService,
		provider,
		orderService,
		conn,
	)

	if err != nil {
		panic(err)
	}

	txqueues = append(txqueues, txq)

	op := &Operator{
		TradeService:      tradeService,
		OrderService:      orderService,
		AccountService:    accountService,
		ObyteProvider:     provider,
		TxQueues:          txqueues,
		QueueAddressIndex: addressIndex,
		mutex:             &sync.Mutex{},
		Broker:            conn,
	}

	go op.HandleEvents()
	return op, nil
}

func (op *Operator) HandleError(m *types.Matches) {
	err := op.Broker.PublishErrorMessage(m, "Server error")
	if err != nil {
		logger.Error(err)
	}
}

func (op *Operator) HandleTxError(m *types.Matches, errType string) {
	err := op.Broker.PublishTxErrorMessage(m, errType)
	if err != nil {
		logger.Error(err)
	}
}

/*func (op *Operator) HandleTxSuccess(m *types.Matches, unot string) {
	err := op.Broker.PublishTradeSuccessMessage(m)
	if err != nil {
		logger.Error(err)
	}
}*/

// Bug: In certain cases, the trade channel seems to be receiving additional unexpected trades.
// In the case TestSocketExecuteOrder (in file socket_test.go) is run on its own, everything is working correctly.
// However, in the case TestSocketExecuteOrder is run among other tests, some tradeLogs do not correspond to an
// order hash in the ordertrade mapping. I suspect this is because the event listener catches events from previous
// tests. It might be helpful to see how to listen to events from up to a certain block.
func (op *Operator) HandleEvents() error {
	events, err := op.ObyteProvider.ListenToEvents()
	if err != nil {
		logger.Error(err)
		return err
	}

	for {
		select {
		case event := <-events:
			logger.Info("Receiving event from wallet", utils.JSON(event))
			arrData := event["data"].([]interface{})
			data := arrData[0].(map[string]interface{})

			switch event["event"] {

			case "loggedin":
				logger.Info("Loggedin event", utils.JSON(data))
				sessionId := data["sessionId"].(string)
				address := data["address"].(string)
				logger.Info("Logged in", sessionId, address)
				go ws.GetLoginSocket().SendMessageBySession(sessionId, address)
				ws.GetLoginSocket().LinkAddressToClient(sessionId, address)

			case "new_order":
				logger.Info("new order event from wallet", utils.JSON(data))
				o := &types.Order{}

				bytes, err := json.Marshal(arrData[0])
				if err != nil {
					logger.Error(err)
					break
				}
				logger.Info("order bytes", string(bytes))

				err = json.Unmarshal(bytes, &o)
				if err != nil {
					logger.Error(err)
					break
				}
				logger.Info("order struct", o)
				logger.Info("order fields", o.UserAddress, o.MatcherAddress, o.BaseToken)

				acc, err := op.AccountService.FindOrCreate(o.UserAddress)
				if err != nil {
					logger.Error(err)
					break
				}
				logger.Info("acc", acc)

				if acc.IsBlocked {
					go ws.SendOrderMessage("ERROR", o.UserAddress, "Account is blocked")
					break
				}

				err = op.OrderService.NewOrder(o)
				if err != nil {
					logger.Error(err)
					go ws.SendOrderMessage("ERROR", o.UserAddress, err.Error())
					break
				}

			case "cancel_order":
				logger.Info("cancel order event from wallet", utils.JSON(data))
				oc := &types.OrderCancel{}

				bytes, err := json.Marshal(arrData[0])
				if err != nil {
					logger.Error(err)
					break
				}
				logger.Info("cancel order bytes", string(bytes))

				err = json.Unmarshal(bytes, &oc)
				if err != nil {
					logger.Error(err)
					break
				}

				ownerAddress, signerAddress, err := op.OrderService.GetSenderAddresses(oc)
				if err != nil {
					logger.Error(err)
					go ws.SendOrderMessage("ERROR", oc.UserAddress, err.Error())
					break
				}

				if ownerAddress != oc.UserAddress && signerAddress != oc.UserAddress {
					authorizedAddresses, err := op.ObyteProvider.GetAuthorizedAddresses(ownerAddress)
					if err != nil {
						logger.Error(err)
						go ws.SendOrderMessage("ERROR", oc.UserAddress, err.Error())
						break
					}
					if !utils.Contains(authorizedAddresses, oc.UserAddress) {
						logger.Error("Not your order")
						go ws.SendOrderMessage("ERROR", oc.UserAddress, "Not your order")
						break
					}
				}

				err = op.OrderService.CancelOrder(oc)
				if err != nil {
					logger.Error(err)
					go ws.SendOrderMessage("ERROR", ownerAddress, err.Error())
					break
				}

			case "revoke":
				logger.Info("revoke event", utils.JSON(data))
				userAddress := data["userAddress"].(string)
				signerAddress := data["signerAddress"].(string)
				logger.Info("revoke authorization on owner", userAddress, "from signer", signerAddress)
				op.OrderService.CancelOrdersSignedByRevokedSigner(userAddress, signerAddress)

			case "balances_update":
				//logger.Info("balances update event", utils.JSON(data))
				address := data["address"].(string)
				ev := data["event"].(string)
				balances_by_symbol := cast.ToStringMapInt64(data["balances_by_symbol"])
				balances_by_asset := cast.ToStringMapInt64(data["balances_by_asset"])
				//logger.Info(address, balances_by_symbol)
				balances_by_symbol = op.OrderService.AdjustBalancesForUncommittedTrades(address, balances_by_symbol)
				//logger.Info("adjusted balances", address, balances_by_symbol)
				op.OrderService.CheckIfBalancesAreSufficientAndCancel(address, balances_by_asset)
				go ws.SendBalancesMessage("UPDATE", address, balances_by_symbol, ev)

			case "exchange_response":
				if data["trigger_unit"] == nil {
					logger.Error("no trigger unit")
					break
				}
				trigger_unit := data["trigger_unit"].(string)
				bounced := false
				switch data["bounced"].(type) {
				case float64:
					bounced = (int(data["bounced"].(float64)) == 1)
				case bool:
					bounced = data["bounced"].(bool)
				default:
					panic("unrecognized type of bounced")
				}
				response := data["response"].(map[string]interface{})

				trades, err := op.TradeService.GetByTriggerUnitHash(trigger_unit)
				if err != nil {
					logger.Error(err)
					break
				}

				if len(trades) == 0 {
					logger.Error("trade not found by trigger unit") // could be a trade by another matcher or bounced withdrawal
					break
				}
				trade := trades[0]

				takerOrder, err := op.OrderService.GetByHash(trade.TakerOrderHash)
				if err != nil {
					logger.Error(err)
					break
				}

				makerOrder, err := op.OrderService.GetByHash(trade.MakerOrderHash)
				if err != nil {
					logger.Error(err)
					break
				}

				matches := &types.Matches{
					MakerOrders: []*types.Order{makerOrder},
					TakerOrder:  takerOrder,
					Trades:      trades,
				}

				if bounced {
					bounceMessage := response["error"].(string)
					op.HandleTxError(matches, bounceMessage) // will also update status to REJECTED
					// the wallet sends balance updates only after successful trades
					go op.sendBalancesUpdateAfterTrade(makerOrder.UserAddress)
					go op.sendBalancesUpdateAfterTrade(takerOrder.UserAddress)
				} else {
					op.TradeService.UpdateTradeStatus(trade, "COMMITTED")
				}

			case "submitted_trades":
				if data["trade_hashes"] == nil {
					logger.Error("no trade hashes")
					break
				}
				trade_hashes := cast.ToStringSlice(data["trade_hashes"])

				trades, err := op.TradeService.GetByHashes(trade_hashes)
				if err != nil {
					logger.Error(err)
					break
				}

				if len(trades) == 0 {
					logger.Error("trades not found by trade hashes") // could be a trade by another matcher or bounced withdrawal
					break
				}

				takerOrder, err := op.OrderService.GetByHash(trades[0].TakerOrderHash)
				if err != nil {
					logger.Error(err)
					break
				}

				matches := &types.Matches{
					MakerOrders: []*types.Order{},
					TakerOrder:  takerOrder,
					Trades:      trades,
				}

				for _, trade := range trades {
					if trade.TakerOrderHash != trades[0].TakerOrderHash {
						err := errors.New("different takers")
						logger.Error(err)
						return err
					}
					makerOrder, err := op.OrderService.GetByHash(trade.MakerOrderHash)
					if err != nil {
						logger.Error(err)
						break
					}
					matches.MakerOrders = append(matches.MakerOrders, makerOrder)
				}

				txq, _, err := op.GetShortestQueue()
				if err != nil {
					logger.Error(err)
					return err
				}
				err = txq.HandleTxSuccess(matches)
				if err != nil {
					logger.Error(err)
					return err
				}
			}
		}
	}
}

func (op *Operator) sendBalancesUpdateAfterTrade(address string) {
	balances := op.ObyteProvider.GetBalances(address)
	balances = op.OrderService.AdjustBalancesForUncommittedTrades(address, balances)
	ws.SendBalancesMessage("UPDATE", address, balances, "trade")
}

func (op *Operator) HandleTrades(msg *types.OperatorMessage) error {
	err := op.QueueTrade(msg.Matches)
	if err != nil {
		logger.Error(err)
		op.HandleError(msg.Matches)
		return err
	}

	return nil
}

// QueueTrade
func (op *Operator) QueueTrade(m *types.Matches) error {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	txq, len, err := op.GetShortestQueue()
	if err != nil {
		logger.Error(err)
		return err
	}

	if len > 10 {
		logger.Warning("Transaction queue is overloaded")
		//return errors.New("Transaction queue is full")
	}

	logger.Infof("Queuing Trade on queue: %v (previous queue length = %v)", txq.Name, len)

	err = txq.PublishPendingTrades(m)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// GetShortestQueue
func (op *Operator) GetShortestQueue() (*TxQueue, int, error) {
	shortest := &TxQueue{}
	min := 1000

	for _, txq := range op.TxQueues {
		if shortest == nil {
			shortest = txq
			min = txq.Length()
		}

		ln := txq.Length()
		if ln < min {
			shortest = txq
			min = ln
		}
	}

	return shortest, min, nil
}

func (op *Operator) PurgeQueues() error {
	for _, txq := range op.TxQueues {
		err := txq.PurgePendingTrades()
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}
