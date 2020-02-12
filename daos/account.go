package daos

import (
	"time"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/types"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AccountDao contains:
// collectionName: MongoDB collection name
// dbName: name of mongodb to interact with
type AccountDao struct {
	collectionName string
	dbName         string
}

// NewAccountDao returns a new instance of AccountDao
func NewAccountDao() *AccountDao {
	dbName := app.Config.DBName
	collection := "accounts"
	index := mgo.Index{
		Key:    []string{"address"},
		Unique: true,
	}

	err := db.Session.DB(dbName).C(collection).EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	return &AccountDao{collection, dbName}
}

// Create function performs the DB insertion task for Balance collection
func (dao *AccountDao) Create(a *types.Account) error {
	a.ID = bson.NewObjectId()
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	err := db.Create(dao.dbName, dao.collectionName, a)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (dao *AccountDao) FindOrCreate(addr string) (*types.Account, error) {
	a := &types.Account{Address: addr}
	query := bson.M{"address": addr}
	updated := &types.Account{}

	change := mgo.Change{
		Update:    types.AccountBSONUpdate{Account: a},
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

func (dao *AccountDao) GetAll() (res []types.Account, err error) {
	err = db.Get(dao.dbName, dao.collectionName, bson.M{}, 0, 0, &res)
	return
}

func (dao *AccountDao) GetByID(id bson.ObjectId) (*types.Account, error) {
	res := []types.Account{}
	q := bson.M{"_id": id}

	err := db.Get(dao.dbName, dao.collectionName, q, 0, 0, &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &res[0], nil
}

func (dao *AccountDao) GetByAddress(owner string) (*types.Account, error) {
	res := []types.Account{}
	q := bson.M{"address": owner}
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

func (dao *AccountDao) GetTokenBalances(owner string) (map[string]*types.TokenBalance, error) {
	q := bson.M{"address": owner}
	res := []types.Account{}
	err := db.Get(dao.dbName, dao.collectionName, q, 0, 1, &res)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil
	}

	return res[0].TokenBalances, nil
}

func (dao *AccountDao) GetTokenBalance(owner string, token string) (*types.TokenBalance, error) {
	q := []bson.M{
		bson.M{
			"$match": bson.M{
				"address": owner,
			},
		},
		bson.M{
			"$project": bson.M{
				"tokenBalances": bson.M{
					"$objectToArray": "$tokenBalances",
				},
				"_id": 0,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"tokenBalances": bson.M{
					"$filter": bson.M{
						"input": "$tokenBalances",
						"as":    "kv",
						"cond": bson.M{
							"$eq": []interface{}{"$$kv.k", token},
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"tokenBalances": bson.M{
					"$arrayToObject": "$tokenBalances",
				},
			},
		},
	}

	var res []*types.Account
	err := db.Aggregate(dao.dbName, dao.collectionName, q, &res)
	if err != nil {
		return nil, err
	}
	//
	//a := &types.Account{}
	//bytes, _ := bson.Marshal(res[0])
	//bson.Unmarshal(bytes, &a)

	return res[0].TokenBalances[token], nil
}

func (dao *AccountDao) UpdateTokenBalance(owner, token string, tokenBalance *types.TokenBalance) error {
	q := bson.M{
		"address": owner,
	}

	updateQuery := bson.M{
		"$set": bson.M{
			"tokenBalances." + token + ".balance":        tokenBalance.Balance,
			"tokenBalances." + token + ".lockedBalance":  tokenBalance.LockedBalance,
			"tokenBalances." + token + ".pendingBalance": tokenBalance.PendingBalance,
		},
	}

	err := db.Update(dao.dbName, dao.collectionName, q, updateQuery)
	return err
}

func (dao *AccountDao) UpdateBalance(owner string, token string, balance int64) error {
	q := bson.M{
		"address": owner,
	}
	updateQuery := bson.M{
		"$set": bson.M{"tokenBalances." + token + ".balance": balance},
	}

	err := db.Update(dao.dbName, dao.collectionName, q, updateQuery)
	return err
}

// Drop drops all the order documents in the current database
func (dao *AccountDao) Drop() {
	db.DropCollection(dao.dbName, dao.collectionName)
}
