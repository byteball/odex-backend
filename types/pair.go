package types

import (
	"encoding/json"
	"math"
	"time"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Pair struct is used to model the pair data in the system and DB
type Pair struct {
	ID                 bson.ObjectId `json:"-" bson:"_id"`
	BaseTokenSymbol    string        `json:"baseTokenSymbol,omitempty" bson:"baseTokenSymbol"`
	BaseAsset          string        `json:"baseAsset,omitempty" bson:"baseAsset"`
	BaseTokenDecimals  int           `json:"baseTokenDecimals,omitempty" bson:"baseTokenDecimals"`
	QuoteTokenSymbol   string        `json:"quoteTokenSymbol,omitempty" bson:"quoteTokenSymbol"`
	QuoteAsset         string        `json:"quoteAsset,omitempty" bson:"quoteAsset"`
	QuoteTokenDecimals int           `json:"quoteTokenDecimals,omitempty" bson:"quoteTokenDecimals"`
	Listed             bool          `json:"listed,omitempty" bson:"listed"`
	Active             bool          `json:"active,omitempty" bson:"active"`
	Rank               int           `json:"rank,omitempty" bson:"rank"`
	CreatedAt          time.Time     `json:"-" bson:"createdAt"`
	UpdatedAt          time.Time     `json:"-" bson:"updatedAt"`
}

func (p *Pair) UnmarshalJSON(b []byte) error {
	pair := map[string]interface{}{}

	err := json.Unmarshal(b, &pair)
	if err != nil {
		return err
	}

	if pair["baseAsset"] != nil {
		p.BaseAsset = pair["baseAsset"].(string)
	}

	if pair["quoteAsset"] != nil {
		p.QuoteAsset = pair["quoteAsset"].(string)
	}

	if pair["baseTokenSymbol"] != nil {
		p.BaseTokenSymbol = pair["baseTokenSymbol"].(string)
	}

	if pair["quoteTokenSymbol"] != nil {
		p.QuoteTokenSymbol = pair["quoteTokenSymbol"].(string)
	}

	if pair["baseTokenDecimals"] != nil {
		p.BaseTokenDecimals = int(pair["baseTokenDecimals"].(float64))
	}

	if pair["quoteTokenDecimals"] != nil {
		p.QuoteTokenDecimals = int(pair["quoteTokenDecimals"].(float64))
	}

	if pair["rank"] != nil {
		p.Rank = int(pair["rank"].(float64))
	}

	return nil
	//TODO do we need the rest of the fields ?
}

func (p *Pair) MarshalJSON() ([]byte, error) {
	pair := map[string]interface{}{
		"baseTokenSymbol":    p.BaseTokenSymbol,
		"baseTokenDecimals":  p.BaseTokenDecimals,
		"quoteTokenSymbol":   p.QuoteTokenSymbol,
		"quoteTokenDecimals": p.QuoteTokenDecimals,
		"baseAsset":          p.BaseAsset,
		"quoteAsset":         p.QuoteAsset,
		"rank":               p.Rank,
		"active":             p.Active,
		"listed":             p.Listed,
	}

	return json.Marshal(pair)
}

type PairAssets struct {
	Name       string `json:"name" bson:"name"`
	BaseToken  string `json:"baseToken" bson:"baseToken"`
	QuoteToken string `json:"quoteToken" bson:"quoteToken"`
}

type PairAssetsRecord struct {
	Name       string `json:"name" bson:"name"`
	BaseToken  string `json:"baseToken" bson:"baseToken"`
	QuoteToken string `json:"quoteToken" bson:"quoteToken"`
}

type PairRecord struct {
	ID bson.ObjectId `json:"id" bson:"_id"`

	BaseTokenSymbol    string    `json:"baseTokenSymbol" bson:"baseTokenSymbol"`
	BaseAsset          string    `json:"baseAsset" bson:"baseAsset"`
	BaseTokenDecimals  int       `json:"baseTokenDecimals" bson:"baseTokenDecimals"`
	QuoteTokenSymbol   string    `json:"quoteTokenSymbol" bson:"quoteTokenSymbol"`
	QuoteAsset         string    `json:"quoteAsset" bson:"quoteAsset"`
	QuoteTokenDecimals int       `json:"quoteTokenDecimals" bson:"quoteTokenDecimals"`
	Active             bool      `json:"active" bson:"active"`
	Listed             bool      `json:"listed" bson:"listed"`
	Rank               int       `json:"rank" bson:"rank"`
	CreatedAt          time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" bson:"updatedAt"`
}

func (p *Pair) BaseTokenMultiplier() int64 {
	return int64(math.Pow(10, float64(p.BaseTokenDecimals)))
}

func (p *Pair) QuoteTokenMultiplier() int64 {
	return int64(math.Pow(10, float64(p.QuoteTokenDecimals)))
}

func (p *Pair) PairMultiplier() int64 {
	//	defaultMultiplier := math.Exp(10, 18)
	baseTokenMultiplier := int64(math.Pow(10, float64(p.BaseTokenDecimals)))
	return baseTokenMultiplier
	//	return math.Mul(defaultMultiplier, baseTokenMultiplier)
}

func (p *Pair) Code() string {
	code := p.BaseTokenSymbol + "/" + p.QuoteTokenSymbol + "::" + p.BaseAsset + "::" + p.QuoteAsset
	return code
}

func (p *Pair) AssetCode() string {
	code := p.BaseAsset + "::" + p.QuoteAsset
	return code
}

func (p *Pair) Name() string {
	name := p.BaseTokenSymbol + "/" + p.QuoteTokenSymbol
	return name
}

func (p *Pair) ParseAmount(a int64) float64 {
	nominator := a
	denominator := p.BaseTokenMultiplier()
	amount := float64(nominator) / float64(denominator)

	return amount
}

func (p *Pair) ParsePrice(pp float64) float64 {
	//	nominator := pp
	//	denominator := math.Mul(math.Exp(10, 18), p.QuoteTokenMultiplier())
	//	price := math.DivideToFloat(nominator, denominator)

	return pp
}

func (p *Pair) MinQuoteAmount() int64 {
	return 0
}

func (p *Pair) SetBSON(raw bson.Raw) error {
	decoded := &PairRecord{}

	err := raw.Unmarshal(decoded)
	if err != nil {
		return err
	}

	p.ID = decoded.ID
	p.BaseTokenSymbol = decoded.BaseTokenSymbol
	p.BaseAsset = decoded.BaseAsset
	p.BaseTokenDecimals = decoded.BaseTokenDecimals
	p.QuoteTokenSymbol = decoded.QuoteTokenSymbol
	p.QuoteAsset = decoded.QuoteAsset
	p.QuoteTokenDecimals = decoded.QuoteTokenDecimals
	p.Listed = decoded.Listed
	p.Active = decoded.Active
	p.Rank = decoded.Rank

	p.CreatedAt = decoded.CreatedAt
	p.UpdatedAt = decoded.UpdatedAt
	return nil
}

func (p *Pair) GetBSON() (interface{}, error) {
	return &PairRecord{
		ID:                 p.ID,
		BaseTokenSymbol:    p.BaseTokenSymbol,
		BaseAsset:          p.BaseAsset,
		BaseTokenDecimals:  p.BaseTokenDecimals,
		QuoteTokenSymbol:   p.QuoteTokenSymbol,
		QuoteAsset:         p.QuoteAsset,
		QuoteTokenDecimals: p.QuoteTokenDecimals,
		Active:             p.Active,
		Listed:             p.Listed,
		Rank:               p.Rank,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}, nil
}

func (p Pair) ValidateAssets() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.BaseAsset, validation.Required),
		validation.Field(&p.QuoteAsset, validation.Required),
	)
}

// Validate function is used to verify if an instance of
// struct satisfies all the conditions for a valid instance
func (p Pair) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.BaseAsset, validation.Required),
		validation.Field(&p.QuoteAsset, validation.Required),
		validation.Field(&p.BaseTokenSymbol, validation.Required),
		validation.Field(&p.QuoteTokenSymbol, validation.Required),
	)
}

// GetOrderBookKeys
func (p *Pair) GetOrderBookKeys() (sell, buy string) {
	return p.GetKVPrefix() + "::SELL", p.GetKVPrefix() + "::BUY"
}

func (p *Pair) GetKVPrefix() string {
	return p.BaseAsset + "::" + p.QuoteAsset
}

type PairData struct {
	Pair               PairID  `json:"pair,omitempty" bson:"_id"`
	Close              float64 `json:"close,omitempty" bson:"close"`
	Count              int64   `json:"count,omitempty" bson:"count"`
	High               float64 `json:"high,omitempty" bson:"high"`
	Low                float64 `json:"low,omitempty" bson:"low"`
	Open               float64 `json:"open,omitempty" bson:"open"`
	Volume             int64   `json:"volume,omitempty" bson:"volume"`
	QuoteVolume        int64   `json:"quoteVolume,omitempty" bson:"quoteVolume"`
	Timestamp          int64   `json:"timestamp,omitempty" bson:"timestamp"`
	OrderVolume        int64   `json:"orderVolume,omitempty" bson:"orderVolume"`
	OrderCount         int64   `json:"orderCount,omitempty" bson:"orderCount"`
	AverageOrderAmount int64   `json:"averageOrderAmount" bson:"averageOrderAmount"`
	AverageTradeAmount int64   `json:"averageTradeAmount" bson:"averageTradeAmount"`
	AskPrice           float64 `json:"askPrice,omitempty" bson:"askPrice"`
	BidPrice           float64 `json:"bidPrice,omitempty" bson:"bidPrice"`
	Price              float64 `json:"price,omitempty" bson:"price"`
	Rank               int     `json:"rank,omitempty" bson:"rank"`
}

func (p *PairData) MarshalJSON() ([]byte, error) {
	pairData := map[string]interface{}{
		"pair": map[string]interface{}{
			"pairName":   p.Pair.PairName,
			"baseToken":  p.Pair.BaseToken,
			"quoteToken": p.Pair.QuoteToken,
		},
		"timestamp": p.Timestamp,
		"rank":      p.Rank,
	}

	if p.Open != 0 {
		pairData["open"] = p.Open
	}

	if p.High != 0 {
		pairData["high"] = p.High
	}

	if p.Low != 0 {
		pairData["low"] = p.Low
	}

	if p.Volume != 0 {
		pairData["volume"] = p.Volume
	}

	if p.QuoteVolume != 0 {
		pairData["quoteVolume"] = p.QuoteVolume
	}

	if p.Close != 0 {
		pairData["close"] = p.Close
	}

	if p.Count != 0 {
		pairData["count"] = p.Count
	}

	if p.OrderVolume != 0 {
		pairData["orderVolume"] = p.OrderVolume
	}

	if p.OrderCount != 0 {
		pairData["orderCount"] = p.OrderCount
	}

	if p.AverageOrderAmount != 0 {
		pairData["averageOrderAmount"] = p.AverageOrderAmount
	}

	if p.AverageTradeAmount != 0 {
		pairData["averageTradeAmount"] = p.AverageTradeAmount
	}

	if p.AskPrice != 0 {
		pairData["askPrice"] = p.AskPrice
	}

	if p.BidPrice != 0 {
		pairData["bidPrice"] = p.BidPrice
	}

	if p.Price != 0 {
		pairData["price"] = p.Price
	}

	bytes, err := json.Marshal(pairData)
	return bytes, err
}

func (p *PairData) AssetCode() string {
	code := p.Pair.BaseToken + "::" + p.Pair.QuoteToken
	return code
}

//ToAPIData converts detailed data into public PairAPIData that contains
func (p *PairData) ToSimplifiedAPIData(pair *Pair) *SimplifiedPairAPIData {
	pairAPIData := SimplifiedPairAPIData{}
	pairAPIData.PairName = p.Pair.PairName
	pairAPIData.LastPrice = pair.ParsePrice(p.Close)
	pairAPIData.Volume = pair.ParseAmount(p.Volume)
	pairAPIData.QuoteVolume = pair.ParseAmount(p.QuoteVolume)
	pairAPIData.OrderVolume = pair.ParseAmount(p.OrderVolume)
	pairAPIData.AverageOrderAmount = pair.ParseAmount(p.AverageOrderAmount)
	pairAPIData.AverageTradeAmount = pair.ParseAmount(p.AverageTradeAmount)
	pairAPIData.TradeCount = int(p.Count)
	pairAPIData.OrderCount = int(p.OrderCount)

	return &pairAPIData
}

func (p *PairData) ToAPIData(pair *Pair) *PairAPIData {
	pairAPIData := PairAPIData{}
	pairAPIData.Pair = p.Pair
	pairAPIData.Open = pair.ParsePrice(p.Open)
	pairAPIData.High = pair.ParsePrice(p.High)
	pairAPIData.Low = pair.ParsePrice(p.Low)
	pairAPIData.Close = pair.ParsePrice(p.Close)
	pairAPIData.Volume = pair.ParseAmount(p.Volume)
	pairAPIData.QuoteVolume = pair.ParseAmount(p.QuoteVolume)
	pairAPIData.Timestamp = int(p.Timestamp)
	pairAPIData.OrderVolume = pair.ParseAmount(p.OrderVolume)
	pairAPIData.OrderCount = int(p.OrderCount)
	pairAPIData.TradeCount = int(p.Count)
	pairAPIData.AverageOrderAmount = pair.ParseAmount(p.AverageOrderAmount)
	pairAPIData.AverageTradeAmount = pair.ParseAmount(p.AverageTradeAmount)
	pairAPIData.AskPrice = pair.ParsePrice(p.AskPrice)
	pairAPIData.BidPrice = pair.ParsePrice(p.BidPrice)
	pairAPIData.Price = pair.ParsePrice(p.Price)
	pairAPIData.Rank = p.Rank

	return &pairAPIData
}

type PairAPIData struct {
	Pair               PairID  `json:"pair" bson:"_id"`
	Open               float64 `json:"open" bson:"open"`
	High               float64 `json:"high" bson:"high"`
	Low                float64 `json:"low" bson:"low"`
	Close              float64 `json:"close" bson:"close"`
	Volume             float64 `json:"volume" bson:"volume"`
	QuoteVolume        float64 `json:"quoteVolume" bson:"quoteVolume"`
	Timestamp          int     `json:"timestamp" bson:"timestamp"`
	OrderVolume        float64 `json:"orderVolume" bson:"orderVolume"`
	OrderCount         int     `json:"orderCount" bson:"orderCount"`
	TradeCount         int     `json:"tradeCount" bson:"tradeCount"`
	AverageOrderAmount float64 `json:"averageOrderAmount" bson:"averageOrderAmount"`
	AverageTradeAmount float64 `json:"averageTradeAmount" bson:"averageTradeAmount"`
	AskPrice           float64 `json:"askPrice" bson:"askPrice"`
	BidPrice           float64 `json:"bidPrice" bson:"bidPrice"`
	Price              float64 `json:"price" bson:"price"`
	Rank               int     `json:"rank" bson:"rank"`
}

//PairAPIData is a similar structure to PairData that contains human-readable data for a certain pair
type SimplifiedPairAPIData struct {
	PairName           string  `json:"pairName"`
	LastPrice          float64 `json:"lastPrice"`
	TradeCount         int     `json:"tradeCount"`
	OrderCount         int     `json:"orderCount"`
	Volume             float64 `json:"volume"`
	QuoteVolume        float64 `json:"quoteVolume"`
	OrderVolume        float64 `json:"orderVolume"`
	AverageOrderAmount float64 `json:"averageOrderAmount"`
	AverageTradeAmount float64 `json:"averageTradeAmount"`
}
