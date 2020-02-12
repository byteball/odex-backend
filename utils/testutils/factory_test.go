package testutils

import (
	"testing"

	"github.com/byteball/odex-backend/types"
)

func TestNewOrderFromFactory(t *testing.T) {
	pair := GetZRXWETHTestPair()
	wallet := GetTestWallet1()
	matcherAddress := GetTestAddress2()
	ZRX := pair.BaseAsset
	WETH := pair.QuoteAsset

	f, err := NewOrderFactory(pair, wallet, matcherAddress)
	if err != nil {
		t.Errorf("Error creating order factory client: %v", err)
	}

	order, err := f.NewOrder(ZRX, WETH, 1, 1)
	if err != nil {
		t.Errorf("Error creating new order: %v", err)
	}

	expected := &types.Order{
		UserAddress:    wallet.Address,
		MatcherAddress: matcherAddress,
		BaseToken:      ZRX,
		QuoteToken:     WETH,
		Amount:         1,
		Hash:           order.Hash,
		Status:         "OPEN",
		Price:          1,
	}

	Compare(t, expected, order)
}

func TestNewFactoryBuyOrder(t *testing.T) {
	matcherAddress := GetTestAddress3()
	pair := GetZRXWETHTestPair()
	wallet := GetTestWallet1()
	ZRX := pair.BaseAsset
	WETH := pair.QuoteAsset

	f, err := NewOrderFactory(pair, wallet, matcherAddress)
	if err != nil {
		t.Errorf("Error creating order factory client: %v", err)
	}

	order, err := f.NewBuyOrder(50, 2)
	if err != nil {
		t.Errorf("Error creating new order: %v", err)
	}

	expected := types.Order{
		UserAddress:         wallet.Address,
		MatcherAddress:      matcherAddress,
		BaseToken:           ZRX,
		QuoteToken:          WETH,
		FilledAmount:        0,
		Price:               50,
		Amount:              2,
		RemainingSellAmount: 100,
		Side:                "BUY",
		Status:              "OPEN",
		PairName:            "ZRX/WETH",
		Hash:                order.Hash,
	}

	CompareOrder(t, &expected, &order)
}

func TestNewFactorySellOrder1(t *testing.T) {
	matcherAddress := GetTestAddress3()
	pair := GetZRXWETHTestPair()
	wallet := GetTestWallet1()
	ZRX := pair.BaseAsset
	WETH := pair.QuoteAsset

	f, err := NewOrderFactory(pair, wallet, matcherAddress)
	if err != nil {
		t.Errorf("Error creating order factory client: %v", err)
	}

	order, err := f.NewSellOrder(100, 1)
	if err != nil {
		t.Errorf("Error creating new order: %v", err)
	}

	expected := types.Order{
		UserAddress:         wallet.Address,
		MatcherAddress:      matcherAddress,
		BaseToken:           ZRX,
		QuoteToken:          WETH,
		FilledAmount:        0,
		Side:                "SELL",
		Status:              "OPEN",
		PairName:            "ZRX/WETH",
		Hash:                order.Hash,
		Price:               100,
		Amount:              1,
		RemainingSellAmount: 1,
	}

	CompareOrder(t, &expected, &order)
}

func TestNewFactorySellOrder2(t *testing.T) {
	matcherAddress := GetTestAddress3()
	pair := GetZRXWETHTestPair()
	wallet := GetTestWallet1()
	ZRX := pair.BaseAsset
	WETH := pair.QuoteAsset

	f, err := NewOrderFactory(pair, wallet, matcherAddress)
	if err != nil {
		t.Errorf("Error creating factory: %v", err)
	}

	order, err := f.NewSellOrder(250, 10) //Selling 10 ZRX at the price of 1 ZRX = 250 WETH
	if err != nil {
		t.Errorf("Error creating new order: %v", err)
	}

	expected := types.Order{
		UserAddress:         wallet.Address,
		MatcherAddress:      matcherAddress,
		BaseToken:           ZRX,
		QuoteToken:          WETH,
		FilledAmount:        0,
		Side:                "SELL",
		Status:              "OPEN",
		PairName:            "ZRX/WETH",
		Hash:                order.Hash,
		Price:               250,
		Amount:              10,
		RemainingSellAmount: 10,
	}

	CompareOrder(t, &expected, &order)
}

func TestNewWebSocketMessage(t *testing.T) {
	matcherAddress := GetTestAddress3()
	pair := GetZRXWETHTestPair()
	wallet := GetTestWallet1()
	ZRX := pair.BaseAsset
	WETH := pair.QuoteAsset

	f, err := NewOrderFactory(pair, wallet, matcherAddress)
	if err != nil {
		t.Errorf("Error creating order factory client: %v", err)
	}

	msg, order, err := f.NewOrderMessage(ZRX, WETH, 1, 1)
	if err != nil {
		t.Errorf("Error creating order message: %v", err)
	}

	expectedOrder := &types.Order{
		UserAddress:    wallet.Address,
		MatcherAddress: matcherAddress,
		BaseToken:      ZRX,
		QuoteToken:     WETH,
		Amount:         1,
		Status:         "OPEN",
		Price:          1,
		Hash:           order.Hash,
	}

	expectedMessage := &types.WebsocketMessage{
		Channel: "orders",
		Event: types.WebsocketEvent{
			Type:    "NEW_ORDER",
			Hash:    order.Hash,
			Payload: expectedOrder,
		},
	}

	Compare(t, expectedMessage, msg)
	Compare(t, expectedOrder, order)
}
