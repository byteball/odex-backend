package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

func TestNewOrderPayload(t *testing.T) {
	p := &NewOrderPayload{
		PairName:       "ZRX/WETH",
		UserAddress:    "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		MatcherAddress: "0xae55690d4b079460e6ac28aaa58c9ec7b73a7485",
		BaseToken:      "0xe41d2489571d322189246dafa5ebde1f4699f498",
		QuoteToken:     "0x12459c951127e0c374ff9105dda097662a027093",
		Amount:         1000,
		Price:          100,
		Hash:           "0xb9070a2d333403c255ce71ddf6e795053599b2e885321de40353832b96d8880a",
	}

	encoded, err := json.Marshal(p)
	if err != nil {
		t.Errorf("Error encoding order: %v", err)
	}

	decoded := &NewOrderPayload{}
	err = json.Unmarshal([]byte(encoded), decoded)
	if err != nil {
		t.Errorf("Could not unmarshal payload: %v", err)
	}

	if diff := deep.Equal(p, decoded); diff != nil {
		fmt.Printf("Expected: \n%+v\nGot: \n%+v\n\n", p, decoded)
	}
}
