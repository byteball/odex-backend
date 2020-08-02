package types

import (
	"encoding/json"

	"github.com/globalsign/mgo/bson"
)

// Tick is the format in which mongo aggregate pipeline returns data when queried for OHLCV data
type Tick struct {
	Pair        PairID  `json:"id,omitempty" bson:"_id"`
	Close       float64 `json:"close,omitempty" bson:"close"`
	Count       int64   `json:"count,omitempty" bson:"count"`
	High        float64 `json:"high,omitempty" bson:"high"`
	Low         float64 `json:"low,omitempty" bson:"low"`
	Open        float64 `json:"open,omitempty" bson:"open"`
	Volume      int64   `json:"volume,omitempty" bson:"volume"`
	QuoteVolume int64   `json:"quoteVolume,omitempty" bson:"quoteVolume"`
	Timestamp   int64   `json:"timestamp,omitempty" bson:"timestamp"`
}

// PairID is the subdocument for aggregate grouping for OHLCV data
type PairID struct {
	PairName   string `json:"pairName" bson:"pairName"`
	BaseToken  string `json:"baseToken" bson:"baseToken"`
	QuoteToken string `json:"quoteToken" bson:"quoteToken"`
}

type OHLCVParams struct {
	Pair     []PairAssets `json:"pair"`
	From     int64        `json:"from"`
	To       int64        `json:"to"`
	Duration int64        `json:"duration"`
	Units    string       `json:"units"`
}

func (t *Tick) AveragePrice() float64 {
	return (t.Open + t.Close) / 2
}

// RoundedVolume returns the value exchanged during this tick in the currency for which the 'exchangeRate' param
// was provided.
func (t *Tick) ConvertedVolume(p *Pair, exchangeRate float64) float64 {
	valueAsToken := float64(t.Volume) / float64(p.BaseTokenMultiplier())
	value := valueAsToken / exchangeRate

	return value
}

// MarshalJSON returns the json encoded byte array representing the trade struct
func (t *Tick) MarshalJSON() ([]byte, error) {
	tick := map[string]interface{}{
		"pair": map[string]interface{}{
			"pairName":   t.Pair.PairName,
			"baseToken":  t.Pair.BaseToken,
			"quoteToken": t.Pair.QuoteToken,
		},
		"timestamp": t.Timestamp,
	}

	if t.Open != 0 {
		tick["open"] = t.Open
	}

	if t.High != 0 {
		tick["high"] = t.High
	}

	if t.Low != 0 {
		tick["low"] = t.Low
	}

	if t.Volume != 0 {
		tick["volume"] = t.Volume
	}

	if t.QuoteVolume != 0 {
		tick["quoteVolume"] = t.QuoteVolume
	}

	if t.Close != 0 {
		tick["close"] = t.Close
	}

	if t.Count != 0 {
		tick["count"] = t.Count
	}

	bytes, err := json.Marshal(tick)
	return bytes, err
}

// UnmarshalJSON creates a trade object from a json byte string
func (t *Tick) UnmarshalJSON(b []byte) error {
	tick := map[string]interface{}{}
	err := json.Unmarshal(b, &tick)

	if err != nil {
		return err
	}

	if tick["pair"] != nil {
		pair := tick["pair"].(map[string]interface{})
		t.Pair = PairID{
			PairName:   pair["pairName"].(string),
			BaseToken:  pair["baseToken"].(string),
			QuoteToken: pair["quoteToken"].(string),
		}
	}

	if tick["timestamp"] != nil {
		t.Timestamp = int64(tick["timestamp"].(float64))
	}

	if tick["open"] != nil {
		t.Open = tick["open"].(float64)
	}

	if tick["high"] != nil {
		t.High = tick["high"].(float64)
	}

	if tick["low"] != nil {
		t.Low = tick["low"].(float64)
	}

	if tick["close"] != nil {
		t.Close = tick["close"].(float64)
	}

	if tick["volume"] != nil {
		t.Volume = int64(tick["volume"].(float64))
	}

	if tick["quoteVolume"] != nil {
		t.QuoteVolume = int64(tick["quoteVolume"].(float64))
	}

	if tick["count"] != nil {
		t.Count = int64(tick["count"].(float64))
	}

	return nil
}

func (t *Tick) GetBSON() (interface{}, error) {
	type PairID struct {
		PairName   string `json:"pairName" bson:"pairName"`
		BaseToken  string `json:"baseToken" bson:"baseToken"`
		QuoteToken string `json:"quoteToken" bson:"quoteToken"`
	}

	count := t.Count

	o := t.Open
	h := t.High
	l := t.Low
	c := t.Close

	v := t.Volume
	qv := t.QuoteVolume

	return struct {
		ID          PairID  `json:"id,omitempty" bson:"_id"`
		Count       int64   `json:"count" bson:"count"`
		Open        float64 `json:"open" bson:"open"`
		High        float64 `json:"high" bson:"high"`
		Low         float64 `json:"low" bson:"low"`
		Close       float64 `json:"close" bson:"close"`
		Volume      int64   `json:"volume" bson:"volume"`
		QuoteVolume int64   `json:"quoteVolume" bson:"quoteVolume"`
		Timestamp   int64   `json:"timestamp" bson:"timestamp"`
	}{
		ID: PairID{
			t.Pair.PairName,
			t.Pair.BaseToken,
			t.Pair.QuoteToken,
		},

		Open:        o,
		High:        h,
		Low:         l,
		Close:       c,
		Volume:      v,
		QuoteVolume: qv,
		Count:       count,
		Timestamp:   t.Timestamp,
	}, nil
}

func (t *Tick) SetBSON(raw bson.Raw) error {
	type PairIDRecord struct {
		PairName   string `json:"pairName" bson:"pairName"`
		BaseToken  string `json:"baseToken" bson:"baseToken"`
		QuoteToken string `json:"quoteToken" bson:"quoteToken"`
	}

	decoded := new(struct {
		Pair        PairIDRecord `json:"pair,omitempty" bson:"_id"`
		Count       int64        `json:"count" bson:"count"`
		Open        float64      `json:"open" bson:"open"`
		High        float64      `json:"high" bson:"high"`
		Low         float64      `json:"low" bson:"low"`
		Close       float64      `json:"close" bson:"close"`
		Volume      int64        `json:"volume" bson:"volume"`
		QuoteVolume int64        `json:"quoteVolume" bson:"quoteVolume"`
		Timestamp   int64        `json:"timestamp" bson:"timestamp"`
	})

	err := raw.Unmarshal(decoded)
	if err != nil {
		return err
	}

	t.Pair = PairID{
		PairName:   decoded.Pair.PairName,
		BaseToken:  decoded.Pair.BaseToken,
		QuoteToken: decoded.Pair.QuoteToken,
	}

	count := decoded.Count
	o := decoded.Open
	h := decoded.High
	l := decoded.Low
	c := decoded.Close
	v := decoded.Volume
	qv := decoded.QuoteVolume

	t.Count = count
	t.Close = c
	t.High = h
	t.Low = l
	t.Open = o
	t.Volume = v
	t.QuoteVolume = qv

	t.Timestamp = decoded.Timestamp
	return nil
}

func (t *Tick) AssetCode() string {
	code := t.Pair.BaseToken + "::" + t.Pair.QuoteToken
	return code
}
