package daos

import (
	"io/ioutil"
	"testing"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
)

func init() {
	server := testutils.NewDBTestServer()
	temp, _ := ioutil.TempDir("", "test")
	server.SetPath(temp)

	session := server.Session()
	db = &Database{Session: session}
}

func TestTokenDao(t *testing.T) {
	dao := NewTokenDao()
	dao.Drop()

	token := &types.Token{
		Symbol:   "PRFT",
		Asset:  "0x6e9a406696617ec5105f9382d33ba3360fcfabcc",
		Decimals: 18,
		Active:   true,
		Quote:    true,
	}

	err := dao.Create(token)
	if err != nil {
		t.Errorf("Could not create token object: %+v", err)
	}

	all, err := dao.GetAll()
	if err != nil {
		t.Errorf("Could not get wallets: %+v", err)
	}

	testutils.CompareToken(t, token, &all[0])

	byId, err := dao.GetByID(token.ID)
	if err != nil {
		t.Errorf("Could not get token by ID: %+v", err)
	}

	testutils.CompareToken(t, token, byId)

	tokenByAsset, err := dao.GetByAsset("0x6e9a406696617ec5105f9382d33ba3360fcfabcc")
	if err != nil {
		t.Errorf("Could not get token by asset: %+v", err)
	}

	testutils.CompareToken(t, token, tokenByAsset)
}
