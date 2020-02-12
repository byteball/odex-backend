package daos

import (
	"io/ioutil"
	"testing"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/globalsign/mgo/bson"
)

func init() {
	server := testutils.NewDBTestServer()
	temp, _ := ioutil.TempDir("", "test")
	server.SetPath(temp)

	session := server.Session()
	db = &Database{session}
}

func TestPairDao(t *testing.T) {
	dao := NewPairDao()

	pair := &types.Pair{
		ID:               bson.NewObjectId(),
		BaseTokenSymbol:  "REQ",
		BaseAsset:        "0xcf7389dc6c63637598402907d5431160ec8972a5",
		QuoteTokenSymbol: "WETH",
		QuoteAsset:       "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		Active:           true,
	}

	err := dao.Create(pair)
	if err != nil {
		t.Errorf("Could not create pair object: %+v", err)
	}

	all, err := dao.GetAll()
	if err != nil {
		t.Errorf("Could not get pairs: %+v", err)
	}

	testutils.ComparePair(t, pair, &all[0])

	byID, err := dao.GetByID(pair.ID)
	if err != nil {
		t.Errorf("Could not get pair by ID: %v", err)
	}

	testutils.ComparePair(t, pair, byID)

	pairByAsset, err := dao.GetByAsset(pair.BaseAsset, pair.QuoteAsset)
	if err != nil {
		t.Errorf("Could not get pair by asset: %v", err)
	}

	testutils.ComparePair(t, pair, pairByAsset)
}
