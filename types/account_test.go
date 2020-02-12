package types

import (
	"testing"

	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestAccountBSON(t *testing.T) {
	assert := assert.New(t)

	address := "0xe8e84ee367bc63ddb38d3d01bccef106c194dc47"
	asset1 := "0xcf7389dc6c63637598402907d5431160ec8972a5"
	asset2 := "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa"

	tokenBalance1 := &TokenBalance{
		Asset:         asset1,
		Symbol:        "EOS",
		Balance:       10000,
		LockedBalance: 5000,
	}

	tokenBalance2 := &TokenBalance{
		Asset:         asset2,
		Symbol:        "ZRX",
		Balance:       10000,
		LockedBalance: 5000,
	}

	account := &Account{
		ID:      bson.NewObjectId(),
		Address: address,
		TokenBalances: map[string]*TokenBalance{
			asset1: tokenBalance1,
			asset2: tokenBalance2,
		},
		IsBlocked: false,
	}

	data, err := bson.Marshal(account)
	if err != nil {
		t.Error(err)
	}

	decoded := &Account{}
	if err := bson.Unmarshal(data, decoded); err != nil {
		t.Error(err)
	}

	assert.Equal(decoded, account)
}
