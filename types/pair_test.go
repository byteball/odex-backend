package types

import (
	"testing"

	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func ComparePair(t *testing.T, a, b *Pair) {
	assert.Equal(t, a.ID, b.ID)
	assert.Equal(t, a.Name(), b.Name())
	assert.Equal(t, a.BaseTokenSymbol, b.BaseTokenSymbol)
	assert.Equal(t, a.BaseAsset, b.BaseAsset)
	assert.Equal(t, a.QuoteTokenSymbol, b.QuoteTokenSymbol)
	assert.Equal(t, a.QuoteAsset, b.QuoteAsset)
	assert.Equal(t, a.Active, b.Active)
}

func TestPairBSON(t *testing.T) {
	pair := &Pair{
		ID:               bson.NewObjectId(),
		BaseTokenSymbol:  "REQ",
		BaseAsset:        "0xcf7389dc6c63637598402907d5431160ec8972a5",
		QuoteTokenSymbol: "WETH",
		QuoteAsset:       "0x7a9f3cd060ab180f36c17fe6bdf9974f577d77aa",
		Active:           true,
	}

	data, err := bson.Marshal(pair)
	if err != nil {
		t.Errorf("%+v", err)
	}

	decoded := &Pair{}
	if err := bson.Unmarshal(data, decoded); err != nil {
		t.Error(err)
	}

	ComparePair(t, pair, decoded)
}
