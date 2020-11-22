package services

import (
	"testing"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/rabbitmq"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/byteball/odex-backend/utils/testutils/mocks"
)

func TestCancelOrder(t *testing.T) {
	orderDao := new(mocks.OrderDao)
	pairDao := new(mocks.PairDao)
	accountDao := new(mocks.AccountDao)
	tradeDao := new(mocks.TradeDao)
	engine := new(mocks.Engine)
	validator := new(mocks.ValidatorService)

	//amqp := rabbitmq.InitConnection("amqp://guest:guest@localhost:5672/")
	amqp := rabbitmq.InitConnection(app.Config.RabbitMQURL)
	orderService := NewOrderService(
		orderDao,
		pairDao,
		accountDao,
		tradeDao,
		validator,
		amqp,
	)

	o := testutils.GetTestOrder1()

	oc := &types.OrderCancel{
		OrderHash:   o.Hash,
		UserAddress: o.UserAddress,
	}

	orderDao.On("GetByHash", o.Hash).Return(&o, nil)
	orderDao.On("UpdateOrderStatus", o.Hash, "CANCELLED").Return(nil)
	engine.On("CancelOrder", oc).Return(nil)

	err := orderService.CancelOrder(oc)
	if err != nil {
		t.Error("Could not cancel order", err)
	}

	orderDao.AssertCalled(t, "GetByHash", o.Hash)
	//engine.AssertNumberOfCalls(t, "CancelOrder", 1)
	//engine.AssertCalled(t, "handleCancelOrder")
}
