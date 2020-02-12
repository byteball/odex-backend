package testutils

import (
	"time"

	"github.com/byteball/odex-backend/types"
	"github.com/globalsign/mgo/bson"
)

func GetTestOrder1() types.Order {
	return types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0000"),
		UserAddress:    "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		MatcherAddress: "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0xe41d2489571d322189246dafa5ebde1f4699f498",
		QuoteToken:     "0x12459c951127e0c374ff9105dda097662a027093",
		Price:          1000,
		Amount:         1000,
		FilledAmount:   100,
		Status:         "OPEN",
		Side:           "BUY",
		PairName:       "ZRX/WETH",
		Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
		CreatedAt:      time.Unix(1405544146, 0),
		UpdatedAt:      time.Unix(1405544146, 0),
	}
}

func GetTestOrder2() types.Order {
	return types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0000"),
		UserAddress:    "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		MatcherAddress: "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0x4bc89ac6f1c55ea645294f3fed949813a768ac6d",
		QuoteToken:     "0xd27a76b12bc4a870c1045c86844161337393d9fa",
		Price:          1200,
		Amount:         1000,
		FilledAmount:   100,
		Status:         "OPEN",
		Side:           "SELL",
		PairName:       "ZRX/WETH",
		Hash:           "0xecf27444c5ce65a88f73db628687fb9b4ac2686b5577df405958d47bee8eaa53",
		CreatedAt:      time.Unix(1405544146, 0),
		UpdatedAt:      time.Unix(1405544146, 0),
	}
}

func GetTestOrder3() types.Order {
	return types.Order{
		ID:             bson.ObjectIdHex("537f700b537461b70c5f0000"),
		UserAddress:    "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		MatcherAddress: "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0x4bc89ac6f1c55ea645294f3fed949813a768ac6d",
		QuoteToken:     "0xd27a76b12bc4a870c1045c86844161337393d9fa",
		Price:          1200,
		Amount:         1000,
		FilledAmount:   100,
		Status:         "OPEN",
		Side:           "SELL",
		PairName:       "ZRX/WETH",
		Hash:           "0x400558b2f5a7b20dd06241c2313c08f652b297e819926b5a51a5abbc60f451e6",
		CreatedAt:      time.Unix(1405544146, 0),
		UpdatedAt:      time.Unix(1405544146, 0),
	}
}
