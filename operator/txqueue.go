package operator

import (
	"encoding/json"
	"errors"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/types"
	"github.com/streadway/amqp"
)

type TxQueue struct {
	Name          string
	TradeService  interfaces.TradeService
	OrderService  interfaces.OrderService
	ObyteProvider interfaces.ObyteProvider
	Broker        *rabbitmq.Connection
}

// NewTxQueue
func NewTxQueue(
	n string,
	tr interfaces.TradeService,
	p interfaces.ObyteProvider,
	o interfaces.OrderService,
	rabbitConn *rabbitmq.Connection,
) (*TxQueue, error) {
	txq := &TxQueue{
		Name:          n,
		TradeService:  tr,
		OrderService:  o,
		ObyteProvider: p,
		Broker:        rabbitConn,
	}

	err := txq.PurgePendingTrades()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	name := "TX_QUEUES:" + txq.Name
	ch := txq.Broker.GetChannel(name)

	q, err := ch.QueueInspect(name + "@" + app.Config.Env)
	if err != nil {
		logger.Error(err)
	}

	err = txq.Broker.ConsumeQueuedTrades(ch, &q, txq.ExecuteTrade)
	if err != nil {
		logger.Error(err)
	}

	return txq, nil
}

func (txq *TxQueue) GetChannel() *amqp.Channel {
	name := "TX_QUEUES" + txq.Name
	return txq.Broker.GetChannel(name)
}

// Length
func (txq *TxQueue) Length() int {
	name := "TX_QUEUES:" + txq.Name
	ch := txq.Broker.GetChannel(name)
	q, err := ch.QueueInspect(name + "@" + app.Config.Env)
	if err != nil {
		logger.Error(err)
	}

	return q.Messages
}

// ExecuteTrade send a trade execution order to the AA. After sending the
// trade message, the trade is updated on the database and is published to the operator subscribers
// (order service)
func (txq *TxQueue) ExecuteTrade(m *types.Matches, tag uint64) error {
	logger.Infof("Executing trades")

	arrTriggerUnits, ex_err := txq.ObyteProvider.ExecuteTrade(m)
	if ex_err != nil {
		//txq.HandleError(m)
		logger.Error(ex_err)
		//if len(arrTriggerUnits) == 0 {
		return ex_err
		//}
	}

	countSuccessful := len(arrTriggerUnits)
	if countSuccessful == 0 {
		panic("no error but units array is empty")
	}

	updatedTrades := []*types.Trade{}
	for i, _ := range arrTriggerUnits {
		/*updated, err := txq.TradeService.UpdatePendingTrade(m.Trades[i], trigger_unit)
		if err != nil {
			logger.Error(err)
		}

		updatedTrades = append(updatedTrades, updated)*/
		m.Trades[i].TxHash = arrTriggerUnits[i]
		m.Trades[i].Status = "SUCCESS"
		updatedTrades = append(updatedTrades, m.Trades[i])
	}

	successfulMatches := types.Matches{
		TakerOrder:  m.TakerOrder,
		MakerOrders: m.MakerOrders[:countSuccessful],
		Trades:      updatedTrades,
	}
	//m.Trades = updatedTrades
	err := txq.Broker.PublishTradeSentMessage(&successfulMatches)
	if err != nil {
		logger.Error(err)
		return errors.New("Could not update")
	}

	/*if countSuccessful != 0 {
		err = txq.HandleTxSuccess(&successfulMatches)
		if err != nil {
			logger.Error(err)
			return err
		}
	}*/

	// if ex_err != nil {
	// 	return ex_err
	// }

	/*if countSuccessful < len(m.MakerOrders) {
		failedMatches := types.Matches{
			TakerOrder:  m.TakerOrder,
			MakerOrders: m.MakerOrders[countSuccessful:],
			Trades:      m.Trades[countSuccessful:],
		}
		txq.HandleError(&failedMatches)
	}*/

	return nil
}

/*func (txq *TxQueue) HandleTradeInvalid(m *types.Matches) error {
	logger.Errorf("Trade invalid: %v", m)

	err := txq.Broker.PublishTradeInvalidMessage(m)
	if err != nil {
		logger.Error(err)
	}

	return nil
}*/

/*func (txq *TxQueue) HandleTxError(m *types.Matches) error {
	logger.Errorf("Transaction Error: %v", m)

	errType := "Transaction error"
	err := txq.Broker.PublishTxErrorMessage(m, errType)
	if err != nil {
		logger.Error(err)
	}

	return nil
}*/

func (txq *TxQueue) HandleTxSuccess(m *types.Matches) error {
	logger.Infof("Transaction success: %v", m)

	err := txq.Broker.PublishTradeSuccessMessage(m)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

/*func (txq *TxQueue) HandleError(m *types.Matches) error {
	logger.Errorf("Operator Error: %v", m)

	errType := "Server error"
	err := txq.Broker.PublishErrorMessage(m, errType)
	if err != nil {
		logger.Error(err)
	}

	return nil
}*/

func (txq *TxQueue) PublishPendingTrades(m *types.Matches) error {
	name := "TX_QUEUES:" + txq.Name
	ch := txq.Broker.GetChannel(name)
	q := txq.Broker.GetQueue(ch, name)

	b, err := json.Marshal(m)
	if err != nil {
		return errors.New("Failed to marshal trade object")
	}

	err = txq.Broker.Publish(ch, q, b)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (txq *TxQueue) PurgePendingTrades() error {
	name := "TX_QUEUES:" + txq.Name
	ch := txq.Broker.GetChannel(name)

	err := txq.Broker.Purge(ch, name)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
