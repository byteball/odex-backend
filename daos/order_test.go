package daos

import (
	"testing"
	"time"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
	"github.com/byteball/odex-backend/utils/testutils"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func init() {
	// temp, _ := ioutil.TempDir("", "test")
	// server.SetPath(temp)

	// session := server.Session()
	// db = &Database{session}
}

func TestUpdateOrderByHash(t *testing.T) {
	exchange := "0x2"

	o := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    "0x1",
		MatcherAddress: exchange,
		BaseToken:      "0x3",
		QuoteToken:     "0x4",
		Price:          1000,
		Amount:         1000,
		FilledAmount:   100,
		Status:         "OPEN",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           "0x8",
		CreatedAt:      time.Unix(1405544146, 0),
		UpdatedAt:      time.Unix(1405544146, 0),
	}

	dao := NewOrderDao()

	err := dao.Create(o)
	if err != nil {
		t.Errorf("Could not create order object")
	}

	updated := &types.Order{
		ID:             o.ID,
		UserAddress:    o.UserAddress,
		MatcherAddress: exchange,
		BaseToken:      o.BaseToken,
		QuoteToken:     o.QuoteToken,
		Price:          4000,
		Amount:         4000,
		FilledAmount:   200,
		Status:         "FILLED",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           o.Hash,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}

	err = dao.UpdateByHash(
		o.Hash,
		updated,
	)

	if err != nil {
		t.Errorf("Could not updated order from hash %v", err)
	}

	queried, err := dao.GetByHash(o.Hash)
	if err != nil {
		t.Errorf("Could not get order by hash")
	}

	testutils.CompareOrder(t, updated, queried)
}

func TestOrderUpdate(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Errorf("Could not drop previous order state")
	}

	o := &types.Order{
		ID:           bson.ObjectIdHex("537f700b537461b70c5f0000"),
		UserAddress:  "0x1",
		BaseToken:    "0x3",
		QuoteToken:   "0x4",
		Price:        1000,
		Amount:       1000,
		FilledAmount: 100,
		Status:       "OPEN",
		Side:         "BUY",
		PairName:     "ZRX/WETH",
		Hash:         "0x8",
		CreatedAt:    time.Unix(1405544146, 0),
		UpdatedAt:    time.Unix(1405544146, 0),
	}

	err = dao.Create(o)
	if err != nil {
		t.Errorf("Could not create order object")
	}

	updated := &types.Order{
		ID:           o.ID,
		UserAddress:  o.UserAddress,
		BaseToken:    o.BaseToken,
		QuoteToken:   o.QuoteToken,
		Price:        4000,
		Amount:       4000,
		FilledAmount: 200,
		Status:       "FILLED",
		Side:         "BUY",
		PairName:     "ZRX/WETH",
		Hash:         o.Hash,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}

	err = dao.Update(
		o.ID,
		updated,
	)

	if err != nil {
		t.Errorf("Could not updated order from hash %v", err)
	}

	queried, err := dao.GetByHash(o.Hash)
	if err != nil {
		t.Errorf("Could not get order by hash")
	}

	testutils.CompareOrder(t, queried, updated)
}

func TestOrderDao1(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Errorf("Could not drop previous order state")
	}

	o := &types.Order{
		ID:           bson.ObjectIdHex("537f700b537461b70c5f0000"),
		UserAddress:  "0x1",
		BaseToken:    "0x3",
		QuoteToken:   "0x4",
		Amount:       1000,
		FilledAmount: 100,
		Status:       "OPEN",
		Side:         "BUY",
		PairName:     "ZRX/WETH",
		Price:        1000,
		Hash:         "0x8",
		CreatedAt:    time.Unix(1405544146, 0),
		UpdatedAt:    time.Unix(1405544146, 0),
	}

	err = dao.Create(o)
	if err != nil {
		t.Errorf("Could not create order object")
	}

	o1, err := dao.GetByHash("0x8")
	if err != nil {
		t.Errorf("Could not get order by hash")
	}

	testutils.CompareOrder(t, o, o1)

	o2, err := dao.GetByUserAddress("0x1")
	if err != nil {
		t.Errorf("Could not get order by user address")
	}

	testutils.CompareOrder(t, o, o2[0])
}

func TestOrderDaoGetByHashes(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order state")
	}

	o1 := testutils.GetTestOrder1()
	o2 := testutils.GetTestOrder2()
	o3 := testutils.GetTestOrder3()

	dao.Create(&o1)
	dao.Create(&o2)
	dao.Create(&o3)

	orders, err := dao.GetByHashes([]string{o1.Hash, o2.Hash})
	if err != nil {
		t.Error("Could not get order by hashes")
	}

	assert.Equal(t, len(orders), 2)
	testutils.CompareOrder(t, orders[0], &o1)
	testutils.CompareOrder(t, orders[1], &o2)
}

func TestGetUserLockedBalance(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"

	p := &types.Pair{
		BaseTokenSymbol:    "ZRX",
		QuoteTokenSymbol:   "WETH",
		BaseAsset:          baseToken,
		QuoteAsset:         quoteToken,
		BaseTokenDecimals:  6,
		QuoteTokenDecimals: 9,
	}

	o1 := &types.Order{
		ID:                  bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:         user,
		MatcherAddress:      exchange,
		FilledAmount:        0,
		Amount:              10,
		RemainingSellAmount: 1000,
		Price:               100,
		BaseToken:           p.BaseAsset,
		QuoteToken:          p.QuoteAsset,
		Status:              "OPEN",
		Side:                "BUY",
		PairName:            "ZRX/WETH",
		Hash:                "0x12",
	}

	o2 := &types.Order{
		ID:                  bson.ObjectIdHex("537f700b537461b70c5f0002"),
		UserAddress:         user,
		MatcherAddress:      exchange,
		FilledAmount:        0,
		RemainingSellAmount: 1000,
		Amount:              10,
		Price:               100,
		BaseToken:           p.BaseAsset,
		QuoteToken:          p.QuoteAsset,
		Status:              "OPEN",
		Side:                "BUY",
		PairName:            "ZRX/WETH",
		Hash:                "0x12",
	}

	o3 := &types.Order{
		ID:                  bson.ObjectIdHex("537f700b537461b70c5f0003"),
		UserAddress:         user,
		MatcherAddress:      exchange,
		FilledAmount:        5,
		Amount:              10,
		Price:               100,
		RemainingSellAmount: 500,
		BaseToken:           p.BaseAsset,
		QuoteToken:          p.QuoteAsset,
		Status:              "PARTIAL_FILLED",
		Side:                "BUY",
		PairName:            "ZRX/WETH",
		Hash:                "0x12",
	}

	o4 := &types.Order{
		ID:                  bson.ObjectIdHex("537f700b537461b70c5f0004"),
		UserAddress:         user,
		MatcherAddress:      exchange,
		Amount:              10,
		FilledAmount:        10,
		Price:               100,
		RemainingSellAmount: 0,
		BaseToken:           p.BaseAsset,
		QuoteToken:          p.QuoteAsset,
		Status:              "FILLED",
		Side:                "BUY",
		PairName:            "ZRX/WETH",
		Hash:                "0x12",
	}

	dao.Create(o1)
	dao.Create(o2)
	dao.Create(o3)
	dao.Create(o4)

	lockedBalance, _, err := dao.GetUserLockedBalance(user, quoteToken)
	if err != nil {
		t.Error("Could not get locked balance", err)
	}

	assert.Equal(t, int64(2500), lockedBalance)
}

func TestGetUserOrderHistory(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"

	o1 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    user,
		MatcherAddress: exchange,
		FilledAmount:   0,
		Amount:         5,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Price:          1e9,
		Status:         "FILLED",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           "0x12",
	}

	o2 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0002"),
		UserAddress:    user,
		MatcherAddress: exchange,
		FilledAmount:   0,
		Amount:         5,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Price:          1e9,
		Status:         "OPEN",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           "0x12",
	}

	o3 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0003"),
		UserAddress:    user,
		MatcherAddress: exchange,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Amount:         5,
		FilledAmount:   5,
		Price:          1e9,
		Status:         "INVALID",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           "0x12",
	}

	o4 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0004"),
		UserAddress:    user,
		MatcherAddress: exchange,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Amount:         5,
		FilledAmount:   10,
		Price:          1e9,
		Status:         "PARTIAL_FILLED",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           "0x12",
	}

	dao.Create(o1)
	dao.Create(o2)
	dao.Create(o3)
	dao.Create(o4)

	orders, err := dao.GetHistoryByUserAddress(user)
	if err != nil {
		t.Error("Could not get order history", err)
	}

	assert.Equal(t, 2, len(orders))
	testutils.CompareOrder(t, orders[0], o1)
	testutils.CompareOrder(t, orders[1], o3)
	assert.NotContains(t, orders, o2)
	assert.NotContains(t, orders, o4)
}

func TestUpdateOrderFilledAmount1(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"
	hash := "0x5"

	o1 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		MatcherAddress: exchange,
		UserAddress:    user,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Amount:         10,
		FilledAmount:   0,
		Status:         "OPEN",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           hash,
	}

	err = dao.Create(o1)
	if err != nil {
		t.Error("Could not create order")
	}

	err = dao.UpdateOrderFilledAmount(hash, 5)
	if err != nil {
		t.Error("Could not get order history", err)
	}

	stored, err := dao.GetByHash(hash)
	if err != nil {
		t.Error("Could not retrieve order", err)
	}

	assert.Equal(t, "PARTIAL_FILLED", stored.Status)
	assert.Equal(t, int64(5), stored.FilledAmount)
}

func TestUpdateOrderFilledAmount2(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"
	hash := "0x5"

	o1 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    user,
		MatcherAddress: exchange,
		Amount:         10,
		FilledAmount:   5,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Status:         "OPEN",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           hash,
	}

	err = dao.Create(o1)
	if err != nil {
		t.Error("Could not create order")
	}

	err = dao.UpdateOrderFilledAmount(hash, 6)
	if err != nil {
		t.Error("Could not get order history", err)
	}

	stored, err := dao.GetByHash(hash)
	if err != nil {
		t.Error("Could not retrieve order", err)
	}

	assert.Equal(t, "FILLED", stored.Status)
	assert.Equal(t, int64(10), stored.FilledAmount)
}

func TestUpdateOrderFilledAmount3(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"
	hash := "0x5"

	o1 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    user,
		MatcherAddress: exchange,
		Amount:         10,
		FilledAmount:   5,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Status:         "OPEN",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           hash,
	}

	err = dao.Create(o1)
	if err != nil {
		t.Error("Could not create order")
	}

	err = dao.UpdateOrderFilledAmount(hash, -6)
	if err != nil {
		t.Error("Could not get order history", err)
	}

	stored, err := dao.GetByHash(hash)
	if err != nil {
		t.Error("Could not retrieve order", err)
	}

	assert.Equal(t, "OPEN", stored.Status)
	assert.Equal(t, int64(0), stored.FilledAmount)
}

func TestUpdateOrderFilledAmounts(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"
	hash1 := "0x5"
	hash2 := "0x6"

	o1 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    user,
		MatcherAddress: exchange,
		Amount:         2,
		FilledAmount:   0,
		Status:         "FILLED",
		Side:           "BUY",
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		PairName:       "ZRX/WETH",
		Hash:           hash1,
	}

	o2 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    user,
		MatcherAddress: exchange,
		Amount:         10,
		FilledAmount:   0,
		Status:         "FILLED",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           hash2,
	}

	err = dao.Create(o1)
	if err != nil {
		t.Error("Could not create order")
	}

	err = dao.Create(o2)
	if err != nil {
		t.Error("Could not create order")
	}

	hashes := []string{hash1, hash2}
	amounts := []int64{-1, -2}
	orders, err := dao.UpdateOrderFilledAmounts(hashes, amounts)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 2, len(orders))
	assert.Equal(t, int64(1), orders[0].FilledAmount)
	assert.Equal(t, int64(2), orders[1].FilledAmount)
}

func TestOrderStatusesByHashes(t *testing.T) {
	dao := NewOrderDao()
	err := dao.Drop()
	if err != nil {
		t.Error("Could not drop previous order collection")
	}

	user := "0x1"
	exchange := "0x2"
	baseToken := "0x3"
	quoteToken := "0x4"
	hash1 := "0x5"
	hash2 := "0x6"

	o1 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		UserAddress:    user,
		MatcherAddress: exchange,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Amount:         1,
		FilledAmount:   1,
		Status:         "FILLED",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           hash1,
	}

	o2 := &types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0002"),
		UserAddress:    user,
		MatcherAddress: exchange,
		BaseToken:      baseToken,
		QuoteToken:     quoteToken,
		Amount:         1,
		FilledAmount:   1,
		Status:         "FILLED",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           hash2,
	}

	err = dao.Create(o1)
	if err != nil {
		t.Error("Could not create order")
	}

	err = dao.Create(o2)
	if err != nil {
		t.Error("Could not create order")
	}

	orders, err := dao.UpdateOrderStatusesByHashes("INVALIDATED", hash1, hash2)
	if err != nil {
		t.Error("Error in updateOrderStatusHashes", err)
	}

	assert.Equal(t, 2, len(orders))
	assert.Equal(t, "INVALIDATED", orders[0].Status)
	assert.Equal(t, "INVALIDATED", orders[1].Status)
}

func ExampleGetOrderBook() {
	session, err := mgo.Dial(app.Config.MongoURL)
	if err != nil {
		panic(err)
	}

	db = &Database{session}
	pairDao := NewPairDao(PairDaoDBOption("odex"))
	orderDao := NewOrderDao(OrderDaoDBOption("odex"))
	pair, err := pairDao.GetByTokenSymbols("BAT", "WETH")
	if err != nil {
		panic(err)
	}

	bids, asks, err := orderDao.GetOrderBook(pair)
	if err != nil {
		panic(err)
	}

	utils.PrintJSON(bids)
	utils.PrintJSON(asks)
}

func ExampleGetOrderBookPrice() {
	session, err := mgo.Dial(app.Config.MongoURL)
	if err != nil {
		panic(err)
	}

	db = &Database{session}

	pairDao := NewPairDao(PairDaoDBOption("odex"))
	orderDao := NewOrderDao(OrderDaoDBOption("odex"))
	pair, err := pairDao.GetByTokenSymbols("AE", "WETH")
	if err != nil {
		panic(err)
	}

	orderPricePoint, _, __, err := orderDao.GetOrderBookPrice(pair, float64(59303), "BUY")
	if err != nil {
		panic(err)
	}

	utils.PrintJSON(orderPricePoint)
}

func ExampleGetRawOrderBook() {
	session, err := mgo.Dial(app.Config.MongoURL)
	if err != nil {
		panic(err)
	}

	db = &Database{session}

	pairDao := NewPairDao(PairDaoDBOption("odex"))
	orderDao := NewOrderDao(OrderDaoDBOption("odex"))
	pair, err := pairDao.GetByTokenSymbols("AE", "WETH")
	if err != nil {
		panic(err)
	}

	orders, err := orderDao.GetRawOrderBook(pair)
	if err != nil {
		panic(err)
	}

	utils.PrintJSON(orders)
}
