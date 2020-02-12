package testutils

import (
	"github.com/byteball/odex-backend/types"
	"github.com/globalsign/mgo/bson"
)

func GetTestTrade1() types.Trade {
	return types.Trade{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0001"),
		Maker:          "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		Taker:          "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0xa114dd77c888aa2edb699de4faa2afbe4575ffd3",
		QuoteToken:     "0x4bc89ac6f1c55ea645294f3fed949813a768ac6d",
		Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
		MakerOrderHash: "0x6d9ad89548c9e3ce4c97825d027291477f2c44a8caef792095f2cabc978493ff",
		TakerOrderHash: "0x6d9ad89548c9e3ce4c97825d027291477f2c44a8caef792095f2cabc978493ff",
		PairName:       "ZRX/WETH",
		Price:          10000000,
		Amount:         100,
	}
}

func GetTestTrade2() types.Trade {
	return types.Trade{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0007"),
		Maker:          "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		Taker:          "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0xa114dd77c888aa2edb699de4faa2afbe4575ffd3",
		QuoteToken:     "0x9aef1ccfe2171300465bb5f752477eb52cb0c59d",
		Hash:           "0xecf27444c5ce65a88f73db628687fb9b4ac2686b5577df405958d47bee8eaa53",
		MakerOrderHash: "0x400558b2f5a7b20dd06241c2313c08f652b297e819926b5a51a5abbc60f451e6",
		TakerOrderHash: "0x400558b2f5a7b20dd06241c2313c08f652b297e819926b5a51a5abbc60f451e6",
		PairName:       "ZRX/DAI",
		Price:          10000000,
		Amount:         100,
	}
}
