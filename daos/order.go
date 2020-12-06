package daos

import (
	"time"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/types"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// OrderDao contains:
// collectionName: MongoDB collection name
// dbName: name of mongodb to interact with
type OrderDao struct {
	collectionName string
	dbName         string
}

type OrderDaoOption = func(*OrderDao) error

func OrderDaoDBOption(dbName string) func(dao *OrderDao) error {
	return func(dao *OrderDao) error {
		dao.dbName = dbName
		return nil
	}
}

// NewOrderDao returns a new instance of OrderDao
func NewOrderDao(opts ...OrderDaoOption) *OrderDao {
	dao := &OrderDao{}
	dao.collectionName = "orders"
	dao.dbName = app.Config.DBName

	for _, op := range opts {
		err := op(dao)
		if err != nil {
			panic(err)
		}
	}

	index := mgo.Index{
		Key:    []string{"hash"},
		Unique: true,
	}

	i1 := mgo.Index{
		Key: []string{"userAddress", "status"},
	}

	i2 := mgo.Index{
		Key: []string{"status"},
	}

	i3 := mgo.Index{
		Key: []string{"baseToken"},
	}

	i4 := mgo.Index{
		Key: []string{"quoteToken"},
	}

	/*i5 := mgo.Index{
		Key:       []string{"price"},
		Collation: &mgo.Collation{NumericOrdering: true, Locale: "en"},
	}*/

	i6 := mgo.Index{
		Key: []string{"baseToken", "quoteToken", "status", "side", "price"},
	}

	i7 := mgo.Index{
		Key: []string{"userAddress", "quoteToken", "side", "status"},
	}

	i8 := mgo.Index{
		Key: []string{"userAddress", "baseToken", "side", "status"},
	}

	/*i7 := mgo.Index{
		Key: []string{"side", "status"},
	}*/

	/*i8 := mgo.Index{
		Key: []string{"baseToken", "quoteToken", "side", "status"},
	}*/

	err := db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i1)
	if err != nil {
		panic(err)
	}

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i2)
	if err != nil {
		panic(err)
	}

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i3)
	if err != nil {
		panic(err)
	}

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i4)
	if err != nil {
		panic(err)
	}

	/*err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i5)
	if err != nil {
		panic(err)
	}*/

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i6)
	if err != nil {
		panic(err)
	}

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i7)
	if err != nil {
		panic(err)
	}

	err = db.Session.DB(dao.dbName).C(dao.collectionName).EnsureIndex(i8)
	if err != nil {
		panic(err)
	}

	return dao
}

// Create function performs the DB insertion task for Order collection
func (dao *OrderDao) Create(o *types.Order) error {
	o.ID = bson.NewObjectId()
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()

	if o.Status == "" {
		o.Status = "OPEN"
	}

	err := db.Create(dao.dbName, dao.collectionName, o)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) DeleteByHashes(hashes ...string) error {
	err := db.RemoveAll(dao.dbName, dao.collectionName, bson.M{"hash": bson.M{"$in": hashes}})
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) Delete(orders ...*types.Order) error {
	hashes := []string{}
	for _, o := range orders {
		hashes = append(hashes, o.Hash)
	}

	err := db.RemoveAll(dao.dbName, dao.collectionName, bson.M{"hash": bson.M{"$in": hashes}})
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// Update function performs the DB updations task for Order collection
// corresponding to a particular order ID
func (dao *OrderDao) Update(id bson.ObjectId, o *types.Order) error {
	o.UpdatedAt = time.Now()

	err := db.Update(dao.dbName, dao.collectionName, bson.M{"_id": id}, o)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) Upsert(id bson.ObjectId, o *types.Order) error {
	o.UpdatedAt = time.Now()

	err := db.Upsert(dao.dbName, dao.collectionName, bson.M{"_id": id}, o)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) UpsertByHash(h string, o *types.Order) error {
	err := db.Upsert(dao.dbName, dao.collectionName, bson.M{"hash": h}, types.OrderBSONUpdate{Order: o})
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) UpdateAllByHash(h string, o *types.Order) error {
	o.UpdatedAt = time.Now()

	err := db.Update(dao.dbName, dao.collectionName, bson.M{"hash": h}, o)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) FindAndModify(h string, o *types.Order) (*types.Order, error) {
	o.UpdatedAt = time.Now()
	query := bson.M{"hash": h}
	updated := &types.Order{}
	change := mgo.Change{
		Update:    types.OrderBSONUpdate{Order: o},
		Upsert:    true,
		Remove:    false,
		ReturnNew: true,
	}

	err := db.FindAndModify(dao.dbName, dao.collectionName, query, change, &updated)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return updated, nil
}

//UpdateByHash updates fields that are considered updateable for an order.
func (dao *OrderDao) UpdateByHash(h string, o *types.Order) error {
	o.UpdatedAt = time.Now()
	query := bson.M{"hash": h}
	update := bson.M{"$set": bson.M{
		"price":               o.Price,
		"amount":              o.Amount,
		"status":              o.Status,
		"filledAmount":        o.FilledAmount,
		"RemainingSellAmount": o.RemainingSellAmount,
		"updatedAt":           o.UpdatedAt,
	}}

	err := db.Update(dao.dbName, dao.collectionName, query, update)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) UpdateOrderStatus(h string, status string) error {
	query := bson.M{"hash": h}
	update := bson.M{"$set": bson.M{
		"status": status,
	}}

	err := db.Update(dao.dbName, dao.collectionName, query, update)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) UpdateOrderStatusesByHashes(status string, hashes ...string) ([]*types.Order, error) {
	hexes := []string{}
	for _, h := range hashes {
		hexes = append(hexes, h)
	}

	query := bson.M{"hash": bson.M{"$in": hexes}}
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
			"status":    status,
		},
	}

	err := db.UpdateAll(dao.dbName, dao.collectionName, query, update)
	if err != nil {
		logger.Error(err)
		return nil, nil
	}

	orders := []*types.Order{}
	err = db.Get(dao.dbName, dao.collectionName, query, 0, 0, &orders)
	if err != nil {
		logger.Error(err)
		return nil, nil
	}

	return orders, nil
}

func (dao *OrderDao) UpdateOrderFilledAmount(hash string, value int64) error {
	q := bson.M{"hash": hash}
	res := []types.Order{}
	err := db.Get(dao.dbName, dao.collectionName, q, 0, 1, &res)
	if err != nil {
		logger.Error(err)
		return err
	}

	o := res[0]
	status := ""
	filledAmount := o.FilledAmount + value

	if filledAmount <= 0 {
		filledAmount = 0
		status = "OPEN"
	} else if filledAmount >= o.Amount {
		filledAmount = o.Amount
		status = "FILLED"
	} else {
		status = "PARTIAL_FILLED"
	}

	update := bson.M{"$set": bson.M{
		"status":       status,
		"filledAmount": filledAmount,
	}}

	err = db.Update(dao.dbName, dao.collectionName, q, update)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *OrderDao) UpdateOrderFilledAmounts(hashes []string, amount []int64) ([]*types.Order, error) {
	hexes := []string{}
	orders := []*types.Order{}
	for i, _ := range hashes {
		hexes = append(hexes, hashes[i])
	}

	query := bson.M{"hash": bson.M{"$in": hexes}}
	err := db.Get(dao.dbName, dao.collectionName, query, 0, 0, &orders)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	updatedOrders := []*types.Order{}
	for i, o := range orders {
		status := ""
		filledAmount := o.FilledAmount - amount[i]

		if filledAmount <= 0 {
			filledAmount = 0
			status = "OPEN"
		} else if filledAmount >= o.Amount {
			filledAmount = o.Amount
			status = "FILLED"
		} else {
			status = "PARTIAL_FILLED"
		}

		query := bson.M{"hash": o.Hash}
		update := bson.M{"$set": bson.M{
			"status":       status,
			"filledAmount": filledAmount,
		}}
		change := mgo.Change{
			Update:    update,
			Upsert:    true,
			Remove:    false,
			ReturnNew: true,
		}

		updated := &types.Order{}
		err := db.FindAndModify(dao.dbName, dao.collectionName, query, change, updated)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		updatedOrders = append(updatedOrders, updated)
	}

	return updatedOrders, nil
}

// GetByID function fetches a single document from order collection based on mongoDB ID.
// Returns Order type struct
func (dao *OrderDao) GetByID(id bson.ObjectId) (*types.Order, error) {
	var response *types.Order
	err := db.GetByID(dao.dbName, dao.collectionName, id, &response)
	return response, err
}

// GetByHash function fetches a single document from order collection based on mongoDB ID.
// Returns Order type struct
func (dao *OrderDao) GetByHash(hash string) (*types.Order, error) {
	q := bson.M{"hash": hash}
	res := []types.Order{}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, 1, &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil
	}

	return &res[0], nil
}

// GetByHashes
func (dao *OrderDao) GetByHashes(hashes []string) ([]*types.Order, error) {
	hexes := []string{}
	for _, h := range hashes {
		hexes = append(hexes, h)
	}

	q := bson.M{"hash": bson.M{"$in": hexes}}
	res := []*types.Order{}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, 0, &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return res, nil
}

// GetByUserAddress function fetches list of orders from order collection based on user address.
// Returns array of Order type struct
func (dao *OrderDao) GetByUserAddress(addr string, limit ...int) ([]*types.Order, error) {
	if limit == nil {
		limit = []int{0}
	}

	var res []*types.Order
	q := bson.M{"userAddress": addr}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, limit[0], &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if res == nil {
		return []*types.Order{}, nil
	}

	return res, nil
}

// GetCurrentByUserAddress function fetches list of open/partial orders from order collection based on user address.
// Returns array of Order type struct
func (dao *OrderDao) GetCurrentByUserAddress(addr string, limit ...int) ([]*types.Order, error) {
	if limit == nil {
		limit = []int{0}
	}

	var res []*types.Order
	q := bson.M{
		"userAddress": addr,
		"status": bson.M{"$in": []string{
			"OPEN",
			"PARTIAL_FILLED",
		},
		},
	}

	err := db.GetAndSort(dao.dbName, dao.collectionName, q, []string{"createdAt"}, 0, limit[0], &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if res == nil {
		return []*types.Order{}, nil
	}

	return res, nil
}

func (dao *OrderDao) GetCurrentByUserAddressAndSignerAddress(address string, signer string) ([]*types.Order, error) {
	var res []*types.Order
	q := bson.M{
		"userAddress":                     address,
		"originalOrder.authors.0.address": signer,
		"status": bson.M{"$in": []string{
			"OPEN",
			"PARTIAL_FILLED",
		},
		},
	}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, 0, &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if res == nil {
		return []*types.Order{}, nil
	}

	return res, nil
}

// GetHistoryByUserAddress function fetches list of orders which are not in open/partial order status
// from order collection based on user address.
// Returns array of Order type struct
func (dao *OrderDao) GetHistoryByUserAddress(addr string, limit ...int) ([]*types.Order, error) {
	if limit == nil {
		limit = []int{0}
	}

	var res []*types.Order
	q := bson.M{
		"userAddress": addr,
		"status": bson.M{"$nin": []string{
			"OPEN",
			"PARTIAL_FILLED",
		},
		},
	}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, limit[0], &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return res, nil
}

func (dao *OrderDao) GetUserLockedBalance(account string, token string) (int64, []*types.Order, error) {
	var orders []*types.Order

	q := bson.M{
		"$or": []bson.M{
			bson.M{
				"userAddress": account,
				"status":      bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"quoteToken":  token,
				"side":        "BUY",
			},
			bson.M{
				"userAddress": account,
				"status":      bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":   token,
				"side":        "SELL",
			},
		},
	}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, 0, &orders)
	if err != nil {
		logger.Error(err)
		return 0, nil, err
	}

	totalLockedBalance := int64(0)
	for _, o := range orders {
		lockedBalance := o.RemainingSellAmount
		totalLockedBalance += lockedBalance
	}

	return totalLockedBalance, orders, nil
}

func (dao *OrderDao) GetRawOrderBook(p *types.Pair) ([]*types.Order, error) {
	var orders []*types.Order
	q := bson.M{
		"status":     bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
		"baseToken":  p.BaseAsset,
		"quoteToken": p.QuoteAsset,
	}
	/*q := []bson.M{
		bson.M{
			"$match": bson.M{
				"status":     bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":  p.BaseAsset,
				"quoteToken": p.QuoteAsset,
			},
		},
		bson.M{
			"$sort": bson.M{
				"price": 1,
			},
		},
	}*/

	err := db.GetAndSort(dao.dbName, dao.collectionName, q, []string{"price"}, 0, 0, &orders)
	if err != nil {
		panic(err)
	}
	/*err := db.Aggregate(dao.dbName, dao.collectionName, q, &orders)
	if err != nil {
		logger.Error(err)
		return nil, err
	}*/

	return orders, nil
}

func (dao *OrderDao) GetOrderBook(p *types.Pair) ([]map[string]interface{}, []map[string]interface{}, error) {
	/*bidsQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status":     bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":  p.BaseAsset,
				"quoteToken": p.QuoteAsset,
				"side":       "BUY",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":            "$price",
				"price":          bson.M{"$first": "$price"},
				"matcherAddress": bson.M{"$first": "$matcherAddress"},
				"amount": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"_id": 1,
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":            0,
				"price":          1,
				"matcherAddress": 1,
				"amount":         1,
			},
		},
	}*/

	/*asksQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status":     bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":  p.BaseAsset,
				"quoteToken": p.QuoteAsset,
				"side":       "SELL",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":            "$price",
				"price":          bson.M{"$first": "$price"},
				"matcherAddress": bson.M{"$first": "$matcherAddress"},
				"amount": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"_id": 1,
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":            0,
				"price":          1,
				"matcherAddress": 1,
				"amount":         1,
			},
		},
	}*/

	bids := []map[string]interface{}{}
	asks := []map[string]interface{}{}
	orders, _ := dao.GetRawOrderBook(p)
	sum := int64(0)
	for i, o := range orders {
		sum += o.Amount - o.FilledAmount
		last := (i == len(orders)-1 || o.Price != orders[i+1].Price || o.Side != orders[i+1].Side || o.MatcherAddress != orders[i+1].MatcherAddress)
		if last {
			entry := map[string]interface{}{
				"price":          o.Price,
				"matcherAddress": o.MatcherAddress,
				"matcherFeeRate": o.MatcherFeeRate(),
				"amount":         sum,
			}
			if o.Side == "SELL" {
				asks = append(asks, entry)
			} else {
				bids = append(bids, entry)
			}
			sum = int64(0)
		}
	}
	/*err := db.Aggregate(dao.dbName, dao.collectionName, bidsQuery, &bids)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	err = db.Aggregate(dao.dbName, dao.collectionName, asksQuery, &asks)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}*/
	logger.Info("bids", bids)

	return bids, asks, nil
}

func (dao *OrderDao) GetOrderBookPrice(p *types.Pair, pp float64, side string) (int64, string, float64, error) {
	q := bson.M{
		"status":     bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
		"baseToken":  p.BaseAsset,
		"quoteToken": p.QuoteAsset,
		"price":      pp,
		"side":       side,
	}
	/*q := []bson.M{
		bson.M{
			"$match": bson.M{
				"status":     bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":  p.BaseAsset,
				"quoteToken": p.QuoteAsset,
				"price":      pp,
				"side":       side,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":            "$price",
				"price":          bson.M{"$first": "$price"},
				"matcherAddress": bson.M{"$first": "$matcherAddress"},
				"amount": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":            0,
				"price":          1,
				"matcherAddress": 1,
				"amount":         1,
			},
		},
	}*/

	var orders []*types.Order
	err := db.Get(dao.dbName, dao.collectionName, q, 0, 0, &orders)
	if err != nil {
		panic(err)
	}
	amount := int64(0)
	matcherFeeRate := float64(0)
	matcherAddress := ""
	for _, o := range orders {
		amount += o.Amount - o.FilledAmount
		if matcherAddress == "" {
			matcherAddress = o.MatcherAddress
			matcherFeeRate = o.MatcherFeeRate()
		}
	}
	return amount, matcherAddress, matcherFeeRate, nil

	/*res := []map[string]interface{}{}
	err := db.Aggregate(dao.dbName, dao.collectionName, q, &res)
	if err != nil {
		logger.Error(err)
		return 0, "", err
	}

	if len(res) == 0 {
		return 0, "", nil
	}

	return cast.ToInt64(res[0]["amount"]), res[0]["matcherAddress"].(string), nil*/
}

func (dao *OrderDao) GetMatchingBuyOrders(o *types.Order) ([]*types.Order, error) {
	var orders []*types.Order

	q := bson.M{
		"status":         bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
		"baseToken":      o.BaseToken,
		"quoteToken":     o.QuoteToken,
		"matcherAddress": o.MatcherAddress,
		"side":           "BUY",
		"price":          bson.M{"$gte": o.Price},
		"$or": []bson.M{
			bson.M{"originalOrder.signed_message.expiry_ts": bson.M{"$exists": false}},
			bson.M{"originalOrder.signed_message.expiry_ts": bson.M{"$gte": time.Now().Unix() + 60}},
		},
	}
	/*q := []bson.M{
		bson.M{
			"$match": bson.M{
				"status":         bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":      o.BaseToken,
				"quoteToken":     o.QuoteToken,
				"matcherAddress": o.MatcherAddress,
				"side":           "BUY",
				"price":          bson.M{"$gte": o.Price},
			},
		},
		bson.M{
			"$sort": bson.M{"price": -1, "createdAt": 1},
		},
	}*/

	err := db.GetAndSort(dao.dbName, dao.collectionName, q, []string{"-price", "createdAt"}, 0, 0, &orders)

	//err := db.Aggregate(dao.dbName, dao.collectionName, q, &orders)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return orders, nil
}

func (dao *OrderDao) GetMatchingSellOrders(o *types.Order) ([]*types.Order, error) {
	var orders []*types.Order

	q := bson.M{
		"status":         bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
		"baseToken":      o.BaseToken,
		"quoteToken":     o.QuoteToken,
		"matcherAddress": o.MatcherAddress,
		"side":           "SELL",
		"price":          bson.M{"$lte": o.Price},
		"$or": []bson.M{
			bson.M{"originalOrder.signed_message.expiry_ts": bson.M{"$exists": false}},
			bson.M{"originalOrder.signed_message.expiry_ts": bson.M{"$gte": time.Now().Unix() + 60}},
		},
	}
	/*q := []bson.M{
		bson.M{
			"$match": bson.M{
				"status":         bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"baseToken":      o.BaseToken,
				"quoteToken":     o.QuoteToken,
				"matcherAddress": o.MatcherAddress,
				"side":           "SELL",
				"price":          bson.M{"$lte": o.Price},
			},
		},
		bson.M{
			"$sort": bson.M{"price": 1, "createdAt": 1},
		},
	}*/

	err := db.GetAndSort(dao.dbName, dao.collectionName, q, []string{"price", "createdAt"}, 0, 0, &orders)
	//err := db.Aggregate(dao.dbName, dao.collectionName, q, &orders)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return orders, nil
}

func (dao *OrderDao) GetExpiredOrders() ([]*types.Order, error) {
	var orders []*types.Order

	q := bson.M{
		"status":                                 bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
		"originalOrder.signed_message.expiry_ts": bson.M{"$lte": time.Now().Unix()},
	}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, 0, &orders)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return orders, nil
}

// Drop drops all the order documents in the current database
func (dao *OrderDao) Drop() error {
	err := db.DropCollection(dao.dbName, dao.collectionName)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// Aggregate function calls the aggregate pipeline of mongodb
func (dao *OrderDao) Aggregate(q []bson.M) ([]*types.OrderData, error) {
	orderData := []*types.OrderData{}
	err := db.Aggregate(dao.dbName, dao.collectionName, q, &orderData)
	if err != nil {
		logger.Error(err)
		return []*types.OrderData{}, err
	}

	return orderData, nil
}
