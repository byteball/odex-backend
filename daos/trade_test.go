package daos

import (
	"io/ioutil"
	"testing"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func init() {
	server := testutils.NewDBTestServer()
	temp, _ := ioutil.TempDir("", "test")
	server.SetPath(temp)

	session := server.Session()
	db = &Database{Session: session}
}

func TestTradeDao(t *testing.T) {
	dao := NewTradeDao()
	dao.Drop()

	ZRXAddress := "0xe41d2489571d322189246dafa5ebde1f4699f498"
	WETHAddress := "0x12459c951127e0c374ff9105dda097662a027093"
	DAIAddress := "0x4dc5790733b997f3db7fc49118ab013182d6ba9b"

	trs := []*types.Trade{
		&types.Trade{
			ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
			Maker:          "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
			Taker:          "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
			BaseToken:      ZRXAddress,
			QuoteToken:     WETHAddress,
			Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
			MakerOrderHash: "0x6d9ad89548c9e3ce4c97825d027291477f2c44a8caef792095f2cabc978493ff",
			TxHash:         "0x41787e3a418997174e2445b51849e79953e334d94a02119e25beff1f13e39aa8",
			PairName:       "ZRX/WETH",
			Price:          10000000,
			Amount:         100,
		},
		&types.Trade{
			ID:             bson.ObjectIdHex("537f700b537461b70c5f0004"),
			Maker:          "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
			Taker:          "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
			BaseToken:      ZRXAddress,
			QuoteToken:     WETHAddress,
			TxHash:         "0xb08514795a779381e0982606e7d33892615ede97dc67f567bf6e4b676db9c9c4",
			Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
			MakerOrderHash: "0x6d9ad89548c9e3ce4c97825d027291477f2c44a8caef792095f2cabc978493ff",
			PairName:       "ZRX/WETH",
			Price:          10000000,
			Amount:         100,
		},
		&types.Trade{
			ID:             bson.ObjectIdHex("537f700b537461b70c5f0007"),
			Maker:          "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
			Taker:          "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
			BaseToken:      ZRXAddress,
			QuoteToken:     DAIAddress,
			TxHash:         "0xf16e0b1ad8536bc43fba0ac009fc19098e19920e045273fa16fa0fc7c83ae1e8",
			Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
			MakerOrderHash: "0x6d9ad89548c9e3ce4c97825d027291477f2c44a8caef792095f2cabc978493ff",
			PairName:       "ZRX/DAI",
			Price:          10000000,
			Amount:         100,
		},
	}

	err := dao.Create(trs[0], trs[1], trs[2])
	if err != nil {
		t.Errorf("Could not create trade objects")
	}

	all, err := dao.GetAll()
	if err != nil {
		t.Errorf("Could not retrieve trade objects")
	}

	assert.Equal(t, len(all), 3)

	tr1, err := dao.GetByHash(trs[0].Hash)
	if err != nil {
		t.Errorf("Could not retrieve hash objects")
	}

	testutils.CompareTrade(t, tr1, trs[0])

	trs2, err := dao.GetByPairName("ZRX/WETH")
	if err != nil {
		t.Errorf("Could not fetch by pair name: %v", err)
	}

	assert.Equal(t, 2, len(trs2))

	testutils.CompareTrade(t, trs2[0], trs[0])
	testutils.CompareTrade(t, trs2[1], trs[1])

	trs3, err := dao.GetAllTradesByPairAssets(ZRXAddress, DAIAddress)
	if err != nil {
		t.Errorf("Could not retrieve objects")
	}

	assert.Equal(t, 1, len(trs3))
	testutils.CompareTrade(t, trs3[0], trs[2])
}

func TestUpdateTrade(t *testing.T) {
	dao := NewTradeDao()
	dao.Drop()

	tr := &types.Trade{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0000"),
		Maker:          "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		Taker:          "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0xe41d2489571d322189246dafa5ebde1f4699f498",
		QuoteToken:     "0x12459c951127e0c374ff9105dda097662a027093",
		Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
		MakerOrderHash: "0x6d9ad89548c9e3ce4c97825d027291477f2c44a8caef792095f2cabc978493ff",
		TxHash:         "Transaction  0xf16e0b1ad8536bc43fba0ac009fc19098e19920e045273fa16fa0fc7c83ae1e8",
		PairName:       "ZRX/WETH",
		Price:          10000000,
		Amount:         100,
	}

	err := dao.Create(tr)
	if err != nil {
		t.Errorf("Could not create trade object")
	}

	updated := &types.Trade{
		ID:             tr.ID,
		Taker:          tr.Taker,
		Maker:          tr.Maker,
		BaseToken:      tr.BaseToken,
		QuoteToken:     tr.QuoteToken,
		MakerOrderHash: tr.MakerOrderHash,
		Hash:           tr.Hash,
		TxHash:         tr.TxHash,
		PairName:       tr.PairName,
		CreatedAt:      tr.CreatedAt,
		UpdatedAt:      tr.UpdatedAt,
		Amount:         tr.Amount,
		Price:          tr.Price,
	}

	err = dao.Update(updated)

	if err != nil {
		t.Errorf("Could not updated order from hash %v", err)
	}

	queried, err := dao.GetByHash(tr.Hash)
	if err != nil {
		t.Errorf("Could not get order by hash")
	}

	testutils.CompareTrade(t, queried, updated)
}
