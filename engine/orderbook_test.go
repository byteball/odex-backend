package engine

import (
	"io/ioutil"
	"log"
	"math"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/byteball/odex-backend/daos"
	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/byteball/odex-backend/utils/testutils/mocks"
)

var db *daos.Database

func setupTest() (
	*Engine,
	*OrderBook,
	string,
	*testutils.Wallet,
	*testutils.Wallet,
	*types.Pair,
	string,
	string,
	*testutils.OrderFactory,
	*testutils.OrderFactory) {

	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetPrefix("\nLOG: ")

	mongoServer := testutils.NewDBTestServer()
	temp, err := ioutil.TempDir("", "test")
	if err != nil {
		panic(err)
	}
	mongoServer.SetPath(temp)

	session := mongoServer.Session()
	daos.InitSession(session)
	rabbitConn := rabbitmq.InitConnection("amqp://guest:guest@localhost:5672/")

	opts := daos.OrderDaoDBOption("test")
	orderDao := daos.NewOrderDao(opts)
	orderDao.Drop()

	matcherAddress := testutils.GetTestAddress1()

	pair := testutils.GetZRXWETHTestPair()
	pairDao := new(mocks.PairDao)
	tradeDao := new(mocks.TradeDao)
	obyteProvider := new(mocks.ObyteProvider)
	orderService := new(mocks.OrderService)
	pairDao.On("GetAll").Return([]types.Pair{*pair}, nil)
	obyteProvider.On("GetOperatorAddress").Return(matcherAddress)
	orderService.On("FixOrderStatus", mock.Anything).Return()

	eng := NewEngine(rabbitConn, orderDao, tradeDao, pairDao, obyteProvider, orderService)
	maker := testutils.GetTestWallet1()
	taker := testutils.GetTestWallet2()
	zrx := pair.BaseAsset
	weth := pair.QuoteAsset

	factory1, err := testutils.NewOrderFactory(pair, maker, matcherAddress)
	if err != nil {
		panic(err)
	}

	factory2, err := testutils.NewOrderFactory(pair, taker, matcherAddress)
	if err != nil {
		panic(err)
	}

	ob := eng.orderbooks[pair.Code()]
	if ob == nil {
		panic("Could not get orderbook")
	}

	return eng, ob, matcherAddress, maker, taker, pair, zrx, weth, factory1, factory2
}

func TestSellOrder(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, _ := setupTest()

	o1, _ := factory1.NewSellOrder(1e3, 1e8, 0)

	exp1 := o1
	exp1.Status = "OPEN"
	expected := &types.EngineResponse{
		Status:  "ORDER_ADDED",
		Order:   &exp1,
		Matches: nil,
	}

	res, err := ob.sellOrder(&o1)
	if err != nil {
		t.Error("Error in sell order: ", err)
	}

	testutils.CompareEngineResponse(t, expected, res)
}

func TestBuyOrder(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, _ := setupTest()

	o1, _ := factory1.NewBuyOrder(1e3, 1e8, 0)

	exp1 := o1
	exp1.Status = "OPEN"
	expected := &types.EngineResponse{
		Status:  "ORDER_ADDED",
		Order:   &exp1,
		Matches: nil,
	}

	res, err := ob.buyOrder(&o1)
	if err != nil {
		t.Error("Error in buy order: ", err)
	}

	testutils.CompareEngineResponse(t, expected, res)
}

func TestFillOrder1(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, factory2 := setupTest()

	o1, _ := factory1.NewSellOrder(1e3, 1e8)
	o2, _ := factory2.NewBuyOrder(1e3, 1e8)
	expt1 := types.NewTrade(&o1, &o2, 1e8, 1e3)

	expo1 := o1
	expo1.Status = "OPEN"
	expectedSellOrderResponse := &types.EngineResponse{Status: "ORDER_ADDED", Order: &expo1}

	expo2 := o2
	expo2.Status = "FILLED"
	expo2.FilledAmount = 1e8
	expo2.RemainingSellAmount = 0

	expo3 := o1
	expo3.Status = "FILLED"
	expo3.FilledAmount = 1e8
	expo3.RemainingSellAmount = 0

	expectedMatches := types.NewMatches(
		[]*types.Order{&expo3},
		&expo2,
		[]*types.Trade{expt1},
	)

	expectedBuyOrderResponse := &types.EngineResponse{
		Status:  "ORDER_FILLED",
		Order:   &expo2,
		Matches: expectedMatches,
	}

	sellOrderResponse, err := ob.sellOrder(&o1)
	if err != nil {
		t.Errorf("Error when calling sell order")
	}

	buyOrderResponse, err := ob.buyOrder(&o2)
	if err != nil {
		t.Errorf("Error when calling buy order")
	}

	testutils.CompareEngineResponse(t, expectedBuyOrderResponse, buyOrderResponse)
	testutils.CompareEngineResponse(t, expectedSellOrderResponse, sellOrderResponse)
}

func TestFillOrder2(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, factory2 := setupTest()

	o1, _ := factory1.NewBuyOrder(1e3, 1e8)
	o2, _ := factory2.NewSellOrder(1e3, 1e8)
	expt1 := types.NewTrade(&o1, &o2, 1e8, 1e3)

	expo1 := o1
	expo1.Status = "OPEN"
	expectedBuyOrderResponse := &types.EngineResponse{
		Status: "ORDER_ADDED",
		Order:  &expo1,
	}

	expo2 := o2
	expo2.Status = "FILLED"
	expo2.FilledAmount = 1e8
	expo2.RemainingSellAmount = 0

	expo3 := o1
	expo3.Status = "FILLED"
	expo3.FilledAmount = 1e8
	expo3.RemainingSellAmount = 0

	expectedMatches := types.NewMatches(
		[]*types.Order{&expo3},
		&expo2,
		[]*types.Trade{expt1},
	)

	expectedSellOrderResponse := &types.EngineResponse{
		Status:  "ORDER_FILLED",
		Order:   &expo2,
		Matches: expectedMatches,
	}

	res1, err := ob.buyOrder(&o1)
	if err != nil {
		t.Error("Error when sending buy order")
	}

	res2, err := ob.sellOrder(&o2)
	if err != nil {
		t.Error("Error when sending sell order")
	}

	testutils.CompareEngineResponse(t, expectedBuyOrderResponse, res1)
	testutils.CompareEngineResponse(t, expectedSellOrderResponse, res2)
}

func TestMultiMatchOrder1(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, factory2 := setupTest()

	so1, _ := factory1.NewSellOrder(1e3+1, 1e8)
	so2, _ := factory1.NewSellOrder(1e3+2, 1e8)
	so3, _ := factory1.NewSellOrder(1e3+3, 1e8)
	bo1, _ := factory2.NewBuyOrder(1e3+4, 3e8)

	ob.sellOrder(&so1)
	ob.sellOrder(&so2)
	ob.sellOrder(&so3)

	expso1 := so1
	expso1.Status = "FILLED"
	expso1.FilledAmount = 1e8
	expso1.RemainingSellAmount = 0
	expso2 := so2
	expso2.Status = "FILLED"
	expso2.FilledAmount = 1e8
	expso2.RemainingSellAmount = 0
	expso3 := so3
	expso3.Status = "FILLED"
	expso3.FilledAmount = 1e8
	expso3.RemainingSellAmount = 0

	expbo1 := bo1
	expbo1.Status = "PARTIAL_FILLED"
	expbo1.FilledAmount = 3e8
	expbo1.RemainingSellAmount = (1e3+4)*3e8 - (1e3+1)*1e8 - (1e3+2)*1e8 - (1e3+3)*1e8

	expt1 := types.NewTrade(&so1, &bo1, 1e8, 1e3+1)
	expt2 := types.NewTrade(&so2, &bo1, 1e8, 1e3+2)
	expt3 := types.NewTrade(&so3, &bo1, 1e8, 1e3+3)

	expectedMatches := types.NewMatches(
		[]*types.Order{&expso1, &expso2, &expso3},
		&expbo1,
		[]*types.Trade{expt1, expt2, expt3},
	)

	expectedResponse := &types.EngineResponse{
		Status:  "ORDER_PARTIALLY_FILLED",
		Order:   &bo1,
		Matches: expectedMatches,
	}

	response, err := ob.buyOrder(&bo1)
	if err != nil {
		t.Errorf("Error in sellOrder: %s", err)
	}

	testutils.CompareEngineResponse(t, expectedResponse, response)
}

func TestMultiMatchOrder2(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, factory2 := setupTest()

	bo1, _ := factory1.NewBuyOrder(1e3+1, 1e8)
	bo2, _ := factory1.NewBuyOrder(1e3+2, 1e8)
	bo3, _ := factory1.NewBuyOrder(1e3+3, 1e8)
	so1, _ := factory2.NewSellOrder(1e3, 3e8)

	expbo1 := bo1
	expbo1.Status = "FILLED"
	expbo1.FilledAmount = 1e8
	expbo1.RemainingSellAmount = 0
	expbo2 := bo2
	expbo2.Status = "FILLED"
	expbo2.FilledAmount = 1e8
	expbo2.RemainingSellAmount = 0
	expbo3 := bo3
	expbo3.Status = "FILLED"
	expbo3.FilledAmount = 1e8
	expbo3.RemainingSellAmount = 0

	expso1 := so1
	expso1.Status = "FILLED"
	expso1.FilledAmount = 3e8
	expso1.RemainingSellAmount = 0

	ob.buyOrder(&bo1)
	ob.buyOrder(&bo2)
	ob.buyOrder(&bo3)

	expt1 := types.NewTrade(&bo1, &so1, 1e8, 1000+1)
	expt2 := types.NewTrade(&bo2, &so1, 1e8, 1000+2)
	expt3 := types.NewTrade(&bo3, &so1, 1e8, 1000+3)

	expectedMatches := types.NewMatches(
		[]*types.Order{&expbo3, &expbo2, &expbo1},
		&expso1,
		[]*types.Trade{expt3, expt2, expt1},
	)

	expectedResponse := &types.EngineResponse{
		Status:  "ORDER_FILLED",
		Order:   &so1,
		Matches: expectedMatches,
	}

	res, err := ob.sellOrder(&so1)
	if err != nil {
		t.Errorf("Error in sell order: %s", err)
	}

	testutils.CompareMatches(t, expectedResponse.Matches, res.Matches)
}

func TestPartialMatchOrder1(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, factory2 := setupTest()

	so1, _ := factory1.NewSellOrder(1e3+1, 1e8)
	so2, _ := factory1.NewSellOrder(1e3+2, 1e8)
	so3, _ := factory1.NewSellOrder(1e3+3, 1e8)
	so4, _ := factory1.NewSellOrder(1e3+4, 2e8)
	bo1, _ := factory2.NewBuyOrder(1e3+5, 4e8)

	expso1 := so1
	expso1.FilledAmount = 1e8
	expso1.RemainingSellAmount = 0
	expso1.Status = "FILLED"
	expso2 := so2
	expso2.FilledAmount = 1e8
	expso2.RemainingSellAmount = 0
	expso2.Status = "FILLED"
	expso3 := so3
	expso3.FilledAmount = 1e8
	expso3.RemainingSellAmount = 0
	expso3.Status = "FILLED"
	expso4 := so4
	expso4.FilledAmount = 100996016
	restAfter3Sells := (1e3+5)*4e8 - (1e3+1)*1e8 - (1e3+2)*1e8 - (1e3+3)*1e8
	filled := int64(math.Round(restAfter3Sells / (1e3 + 4)))
	expso4.FilledAmount = filled
	expso4.RemainingSellAmount = expso4.Amount - expso4.FilledAmount
	expso4.Status = "PARTIAL_FILLED"

	expbo1 := bo1
	expbo1.FilledAmount = 3*1e8 + filled
	expbo1.RemainingSellAmount = 0
	expbo1.Status = "FILLED"

	expt1 := types.NewTrade(&so1, &bo1, 1e8, 1e3+1)
	expt2 := types.NewTrade(&so2, &bo1, 1e8, 1e3+2)
	expt3 := types.NewTrade(&so3, &bo1, 1e8, 1e3+3)
	expt4 := types.NewTrade(&so4, &bo1, filled, 1e3+4)

	ob.sellOrder(&so1)
	ob.sellOrder(&so2)
	ob.sellOrder(&so3)
	ob.sellOrder(&so4)

	res, err := ob.buyOrder(&bo1)
	if err != nil {
		t.Errorf("Error when buying order")
	}

	expectedMatches := types.NewMatches(
		[]*types.Order{&expso1, &expso2, &expso3, &expso4},
		&expbo1,
		[]*types.Trade{expt1, expt2, expt3, expt4},
	)

	expectedResponse := &types.EngineResponse{
		Status:  "ORDER_FILLED",
		Order:   &expbo1,
		Matches: expectedMatches,
	}

	testutils.CompareEngineResponse(t, expectedResponse, res)
}

func TestPartialMatchOrder2(t *testing.T) {
	_, ob, _, _, _, _, _, _, factory1, factory2 := setupTest()

	bo1, _ := factory1.NewBuyOrder(1e3+5, 1e8)
	bo2, _ := factory1.NewBuyOrder(1e3+4, 1e8)
	bo3, _ := factory1.NewBuyOrder(1e3+3, 1e8)
	bo4, _ := factory1.NewBuyOrder(1e3+2, 2e8)
	so1, _ := factory2.NewSellOrder(1e3+1, 4e8)

	expbo1 := bo1
	expbo1.FilledAmount = 1e8
	expbo1.RemainingSellAmount = 0
	expbo1.Status = "FILLED"
	expbo2 := bo2
	expbo2.FilledAmount = 1e8
	expbo2.RemainingSellAmount = 0
	expbo2.Status = "FILLED"
	expbo3 := bo3
	expbo3.FilledAmount = 1e8
	expbo3.RemainingSellAmount = 0
	expbo3.Status = "FILLED"
	expbo4 := bo4
	expbo4.FilledAmount = 1e8
	expbo4.RemainingSellAmount = (1e3 + 2) * 1e8
	expbo4.Status = "PARTIAL_FILLED"

	expso1 := so1
	expso1.FilledAmount = 4e8
	expso1.RemainingSellAmount = 0
	expso1.Status = "FILLED"

	expt1 := types.NewTrade(&bo1, &so1, 1e8, 1e3+5)
	expt2 := types.NewTrade(&bo2, &so1, 1e8, 1e3+4)
	expt3 := types.NewTrade(&bo3, &so1, 1e8, 1e3+3)
	expt4 := types.NewTrade(&bo4, &so1, 1e8, 1e3+2)

	ob.buyOrder(&bo1)
	ob.buyOrder(&bo2)
	ob.buyOrder(&bo3)
	ob.buyOrder(&bo4)

	res, err := ob.sellOrder(&so1)
	if err != nil {
		t.Errorf("Error when buying order")
	}

	expectedMatches := types.NewMatches(
		[]*types.Order{&expbo1, &expbo2, &expbo3, &expbo4},
		&expso1,
		[]*types.Trade{expt1, expt2, expt3, expt4},
	)

	expectedResponse := &types.EngineResponse{
		Status:  "ORDER_FILLED",
		Order:   &expso1,
		Matches: expectedMatches,
	}

	testutils.CompareEngineResponse(t, expectedResponse, res)
}
