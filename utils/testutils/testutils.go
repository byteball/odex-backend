package testutils

import (
	"github.com/byteball/odex-backend/utils/testutils/mocks"
)

type MockServices struct {
	AccountService   *mocks.AccountService
	OrderService     *mocks.OrderService
	OrderBookService *mocks.OrderBookService
	TokenService     *mocks.TokenService
	PairService      *mocks.PairService
	TradeService     *mocks.TradeService
}

type MockDaos struct {
	AccountDao *mocks.AccountDao
	OrderDao   *mocks.OrderDao
	TokenDao   *mocks.TokenDao
	TradeDao   *mocks.TradeDao
	PairDao    *mocks.PairDao
}

func NewMockServices() *MockServices {
	return &MockServices{
		AccountService:   new(mocks.AccountService),
		OrderService:     new(mocks.OrderService),
		OrderBookService: new(mocks.OrderBookService),
		TokenService:     new(mocks.TokenService),
		PairService:      new(mocks.PairService),
	}
}

func NewMockDaos() *MockDaos {
	return &MockDaos{
		AccountDao: new(mocks.AccountDao),
		OrderDao:   new(mocks.OrderDao),
		TokenDao:   new(mocks.TokenDao),
		TradeDao:   new(mocks.TradeDao),
		PairDao:    new(mocks.PairDao),
	}
}
