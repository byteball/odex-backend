package daos

import (
	"io/ioutil"
	"testing"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

// var db *Database

// func TestMain(m *testing.M) {
// 	db := &Database{}
// 	dropTestServer := testutils.InitDBTestServer(db)
// 	defer dropTestServer()
// 	m.Run()
// }

func init() {
	server := testutils.NewDBTestServer()
	temp, _ := ioutil.TempDir("", "test")
	server.SetPath(temp)

	session := server.Session()
	db = &Database{Session: session}
}

func TestAccountDao(t *testing.T) {
	dao := NewAccountDao()
	dao.Drop()

	address := "0xe8e84ee367bc63ddb38d3d01bccef106c194dc47"
	asset1 := "0xcf7389dc6c63637598402907d5431160ec8972a5"
	asset2 := "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa"

	tokenBalance1 := &types.TokenBalance{
		Asset:         asset1,
		Symbol:        "EOS",
		Balance:       10000,
		LockedBalance: 5000,
	}

	tokenBalance2 := &types.TokenBalance{
		Asset:         asset2,
		Symbol:        "ZRX",
		Balance:       10000,
		LockedBalance: 5000,
	}

	account := &types.Account{
		ID:      bson.NewObjectId(),
		Address: address,
		TokenBalances: map[string]*types.TokenBalance{
			asset1: tokenBalance1,
			asset2: tokenBalance2,
		},
		IsBlocked: false,
	}

	err := dao.Create(account)
	if err != nil {
		t.Errorf("Could not create order object")
	}

	a1, err := dao.GetByAddress(account.Address)
	if err != nil {
		t.Errorf("Could not get order by hash: %v", err)
	}

	testutils.CompareAccount(t, account, a1)
}

func TestAccountGetAllTokenBalances(t *testing.T) {
	dao := NewAccountDao()
	dao.Drop()

	address := "0xe8e84ee367bc63ddb38d3d01bccef106c194dc47"
	asset1 := "0xcf7389dc6c63637598402907d5431160ec8972a5"
	asset2 := "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa"

	tokenBalance1 := &types.TokenBalance{
		Asset:         asset1,
		Symbol:        "EOS",
		Balance:       10000,
		LockedBalance: 5000,
	}

	tokenBalance2 := &types.TokenBalance{
		Asset:         asset2,
		Symbol:        "ZRX",
		Balance:       10000,
		LockedBalance: 5000,
	}

	account := &types.Account{
		ID:      bson.NewObjectId(),
		Address: address,
		TokenBalances: map[string]*types.TokenBalance{
			asset1: tokenBalance1,
			asset2: tokenBalance2,
		},
		IsBlocked: false,
	}

	err := dao.Create(account)
	if err != nil {
		t.Errorf("Could not create account object")
	}

	balances, err := dao.GetTokenBalances(account.Address)
	if err != nil {
		t.Errorf("Could not retrieve token balances: %v", balances)
	}

	assert.Equal(t, balances[asset1], tokenBalance1)
	assert.Equal(t, balances[asset2], tokenBalance2)
}

func TestGetTokenBalance(t *testing.T) {
	dao := NewAccountDao()
	dao.Drop()

	address := "0xe8e84ee367bc63ddb38d3d01bccef106c194dc47"
	asset1 := "0xcf7389dc6c63637598402907d5431160ec8972a5"
	asset2 := "0xe41d2489571d322189246dafa5ebde1f4699f498"

	tokenBalance1 := &types.TokenBalance{
		Asset:         asset1,
		Symbol:        "EOS",
		Balance:       10000,
		LockedBalance: 5000,
	}

	tokenBalance2 := &types.TokenBalance{
		Asset:         asset2,
		Symbol:        "ZRX",
		Balance:       10000,
		LockedBalance: 5000,
	}

	account := &types.Account{
		ID:      bson.NewObjectId(),
		Address: address,
		TokenBalances: map[string]*types.TokenBalance{
			asset1: tokenBalance1,
			asset2: tokenBalance2,
		},
		IsBlocked: false,
	}

	err := dao.Create(account)
	if err != nil {
		t.Errorf("Could not create account: %v", err)
	}

	balance, err := dao.GetTokenBalance(address, asset2)
	if err != nil {
		t.Errorf("Could not get token balance: %v", err)
	}

	assert.Equal(t, balance, tokenBalance2)
}

func TestUpdateAccountBalance(t *testing.T) {
	dao := NewAccountDao()
	dao.Drop()

	address := "0xe8e84ee367bc63ddb38d3d01bccef106c194dc47"
	asset1 := "0xcf7389dc6c63637598402907d5431160ec8972a5"
	asset2 := "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa"

	tokenBalance1 := &types.TokenBalance{
		Asset:         asset1,
		Symbol:        "EOS",
		Balance:       10000,
		LockedBalance: 5000,
	}

	tokenBalance2 := &types.TokenBalance{
		Asset:         asset2,
		Symbol:        "ZRX",
		Balance:       10000,
		LockedBalance: 5000,
	}

	account := &types.Account{
		ID:      bson.NewObjectId(),
		Address: address,
		TokenBalances: map[string]*types.TokenBalance{
			asset1: tokenBalance1,
			asset2: tokenBalance2,
		},
		IsBlocked: false,
	}

	err := dao.Create(account)
	if err != nil {
		t.Errorf("Could not create account object")
	}

	err = dao.UpdateBalance(address, asset1, 20000)
	if err != nil {
		t.Errorf("Could not update balance")
	}

	balance, err := dao.GetTokenBalance(address, asset1)
	if err != nil {
		t.Errorf("Could not get token balance: %v", err)
	}

	assert.Equal(t, balance.Balance, int64(20000))
}
