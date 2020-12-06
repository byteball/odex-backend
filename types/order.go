package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

// Order contains the data related to an order sent by the user
type Order struct {
	ID                  bson.ObjectId          `json:"id" bson:"_id"`
	UserAddress         string                 `json:"userAddress" bson:"userAddress"`
	MatcherAddress      string                 `json:"matcherAddress" bson:"matcherAddress"`
	AffiliateAddress    string                 `json:"affiliateAddress" bson:"affiliateAddress"`
	BaseToken           string                 `json:"baseToken" bson:"baseToken"`
	QuoteToken          string                 `json:"quoteToken" bson:"quoteToken"`
	Status              string                 `json:"status" bson:"status"`
	Side                string                 `json:"side" bson:"side"`
	Hash                string                 `json:"hash" bson:"hash"`
	Price               float64                `json:"price" bson:"price"`
	Amount              int64                  `json:"amount" bson:"amount"`
	FilledAmount        int64                  `json:"filledAmount" bson:"filledAmount"`
	RemainingSellAmount int64                  `json:"remainingSellAmount" bson:"remainingSellAmount"`
	PairName            string                 `json:"pairName" bson:"pairName"`
	OriginalOrder       map[string]interface{} `json:"originalOrder" bson:"originalOrder"`
	CreatedAt           time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt" bson:"updatedAt"`
}

func (o *Order) String() string {
	return fmt.Sprintf("Pair: %v, Price: %v, Hash: %v", o.PairName, o.Price, o.Hash)
}

// TODO: Verify userAddress, baseToken, quoteToken, etc. conditions are working
func (o *Order) Validate() error {
	if o.UserAddress == "" {
		return errors.New("Order 'userAddress' parameter is required")
	}

	if o.BaseToken == "" {
		return errors.New("Order 'baseToken' parameter is required")
	}

	if o.QuoteToken == "" {
		return errors.New("Order 'quoteToken' parameter is required")
	}

	if o.Amount == 0 {
		return errors.New("Order 'amount' parameter is required")
	}

	if o.Price == 0 {
		return errors.New("Order 'price' parameter is required")
	}

	if o.Side != "BUY" && o.Side != "SELL" {
		return errors.New("Order 'side' should be 'SELL' or 'BUY'")
	}

	if o.Amount <= 0 {
		return errors.New("Order 'amount' parameter should be strictly positive")
	}

	if o.Price <= 0 {
		return errors.New("Order 'price' parameter should be strictly positive")
	}

	return nil
}

/*
// ComputeHash calculates the orderRequest hash
func (o *Order) ComputeHash() string {
	sha := sha256.New()
	sha.Write([]byte(o.MatcherAddress))
	sha.Write([]byte(o.AffiliateAddress))
	sha.Write([]byte(o.UserAddress))
	sha.Write([]byte(o.BaseToken))
	sha.Write([]byte(o.QuoteToken))
	sha.Write(o.Amount.Bytes())
	sha.Write(o.PricePoint.Bytes())
	sha.Write(o.EncodedSide().Bytes())
	sha.Write(o.Nonce.Bytes())
	return base64.StdEncoding.EncodeToString(sha.Sum(nil))
}
*/

func (o *Order) Process(p *Pair) error {
	if o.FilledAmount == 0 {
		o.FilledAmount = 0
	}
	if o.RemainingSellAmount == 0 {
		o.RemainingSellAmount = o.SellAmount(p)
	}
	o.PairName = p.Name()
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) Pair() (*Pair, error) {
	if o.BaseToken == "" {
		return nil, errors.New("Base token is not set")
	}

	if o.QuoteToken == "" {
		return nil, errors.New("Quote token is set")
	}

	return &Pair{
		BaseAsset:  o.BaseToken,
		QuoteAsset: o.QuoteToken,
	}, nil
}

func (o *Order) RemainingAmount() int64 {
	return o.Amount - o.FilledAmount
}

func (o *Order) SellTokenSymbol() string {
	if o.Side == "BUY" {
		return o.QuoteTokenSymbol()
	}

	if o.Side == "SELL" {
		return o.BaseTokenSymbol()
	}

	return ""
}

//TODO handle error case
func (o *Order) SellToken() string {
	if o.Side == "BUY" {
		return o.QuoteToken
	} else {
		return o.BaseToken
	}
}

func (o *Order) BuyToken() string {
	if o.Side == "BUY" {
		return o.BaseToken
	} else {
		return o.QuoteToken
	}
}

func (o *Order) QuoteAmount(p *Pair) int64 {
	return int64(math.Round(float64(o.Amount) * o.Price))
}

func (o *Order) RemainingQuoteAmount() int64 {
	if o.Side == "BUY" {
		return o.RemainingSellAmount
	} else {
		return int64(math.Round(float64(o.RemainingSellAmount) * o.Price))
	}
}

// SellAmount
// If order is a "BUY", then sellToken = quoteToken
func (o *Order) SellAmount(p *Pair) int64 {
	//pairMultiplier := p.PairMultiplier()
	OriginalOrderData := o.OriginalOrder["signed_message"].(map[string]interface{})
	sellAmount := int64(OriginalOrderData["sell_amount"].(float64))
	if o.Side == "SELL" && sellAmount != o.Amount {
		panic("sell amount and amount mismatch")
	}
	return sellAmount
	/*if o.Side == "BUY" {
		return int64(math.Round(float64(o.Amount) * o.Price))
	} else {
		return o.Amount
	}*/
}

func (o *Order) OriginalPrice() float64 {
	OriginalOrderData := o.OriginalOrder["signed_message"].(map[string]interface{})
	return OriginalOrderData["price"].(float64)
}

func (o *Order) MatcherFeeRate() float64 {
	OriginalOrderData := o.OriginalOrder["signed_message"].(map[string]interface{})
	if (OriginalOrderData["matcher_fee_asset"].(string) == OriginalOrderData["sell_asset"].(string)){
		return OriginalOrderData["matcher_fee"].(float64) / OriginalOrderData["sell_amount"].(float64)
	} else if (OriginalOrderData["matcher_fee_asset"].(string) == OriginalOrderData["buy_asset"].(string)){
		return OriginalOrderData["matcher_fee"].(float64) / (OriginalOrderData["sell_amount"].(float64) * OriginalOrderData["price"].(float64))
	} else {
		panic("matcher fee asset not a pair asset")
	}
}

/*func (o *Order) RemainingSellAmount() int64 {
	//pairMultiplier := p.PairMultiplier()

	if o.Side == "BUY" {
		remainingAmount := o.Amount - o.FilledAmount
		return int64(math.Round(float64(remainingAmount) * o.Price))
	} else {
		return o.Amount - o.FilledAmount
	}
}*/

func (o *Order) RequiredSellAmount(p *Pair) int64 {
	var requiredSellTokenAmount int64

	//pairMultiplier := p.PairMultiplier()

	if o.Side == "BUY" {
		requiredSellTokenAmount = int64(math.Round(float64(o.Amount) * o.Price))
	} else {
		requiredSellTokenAmount = o.Amount
	}

	return requiredSellTokenAmount
}

func (o *Order) TotalRequiredSellAmount(p *Pair) int64 {
	var requiredSellTokenAmount int64

	//pairMultiplier := p.PairMultiplier()

	if o.Side == "BUY" {
		// selling the quote token, the fee is also paid in quote token
		OriginalOrderData := o.OriginalOrder["signed_message"].(map[string]interface{})
		requiredSellTokenAmount = int64(OriginalOrderData["sell_amount"].(float64)) + int64(OriginalOrderData["matcher_fee"].(float64))
	} else {
		requiredSellTokenAmount = o.Amount
	}

	if o.SellToken() == "base" {
		requiredSellTokenAmount += 10 * 1000 // add AA fees for 10 trades
	}

	return requiredSellTokenAmount
}

func (o *Order) BuyAmount(pairMultiplier int64) int64 {
	if o.Side == "SELL" {
		return o.Amount
	} else {
		return int64(math.Round(float64(o.Amount) * o.Price))
	}
}

//TODO handle error case ?
func (o *Order) EncodedSide() int64 {
	if o.Side == "BUY" {
		return 0
	} else {
		return 1
	}
}

func (o *Order) BuyTokenSymbol() string {
	if o.Side == "BUY" {
		return o.BaseTokenSymbol()
	}

	if o.Side == "SELL" {
		return o.QuoteTokenSymbol()
	}

	return ""
}

func (o *Order) PairCode() (string, error) {
	if o.PairName == "" {
		return "", errors.New("Pair name is required")
	}

	return o.PairName + "::" + o.BaseToken + "::" + o.QuoteToken, nil
}

func (o *Order) BaseTokenSymbol() string {
	if o.PairName == "" {
		return ""
	}

	return o.PairName[:strings.IndexByte(o.PairName, '/')]
}

func (o *Order) QuoteTokenSymbol() string {
	if o.PairName == "" {
		return ""
	}

	return o.PairName[strings.IndexByte(o.PairName, '/')+1:]
}

// JSON Marshal/Unmarshal interface

// MarshalJSON implements the json.Marshal interface
func (o *Order) MarshalJSON() ([]byte, error) {
	order := map[string]interface{}{
		"matcherAddress":      o.MatcherAddress,
		"affiliateAddress":    o.AffiliateAddress,
		"userAddress":         o.UserAddress,
		"baseToken":           o.BaseToken,
		"quoteToken":          o.QuoteToken,
		"side":                o.Side,
		"status":              o.Status,
		"pairName":            o.PairName,
		"amount":              o.Amount,
		"remainingSellAmount": o.RemainingSellAmount,
		"price":               o.Price,
		// NOTE: Currently removing this to simplify public API, might reinclude
		// later. An alternative would be to create additional simplified type
		"createdAt": o.CreatedAt.Format(time.RFC3339Nano),
		// "updatedAt": o.UpdatedAt.Format(time.RFC3339Nano),
		"originalOrder": o.OriginalOrder,
	}

	if o.FilledAmount != 0 {
		order["filledAmount"] = o.FilledAmount
	}

	if o.Hash != "" {
		order["hash"] = o.Hash
	}

	return json.Marshal(order)
}

func (o *Order) UnmarshalJSON(b []byte) error {
	order := map[string]interface{}{}

	err := json.Unmarshal(b, &order)
	if err != nil {
		return err
	}

	if order["id"] != nil && bson.IsObjectIdHex(order["id"].(string)) {
		o.ID = bson.ObjectIdHex(order["id"].(string))
	}

	if order["pairName"] != nil {
		o.PairName = order["pairName"].(string)
	}

	if order["matcherAddress"] != nil {
		o.MatcherAddress = order["matcherAddress"].(string)
	}

	if order["affiliateAddress"] != nil {
		o.AffiliateAddress = order["affiliateAddress"].(string)
	}

	if order["userAddress"] != nil {
		o.UserAddress = order["userAddress"].(string)
	}

	if order["baseToken"] != nil {
		o.BaseToken = order["baseToken"].(string)
	}

	if order["quoteToken"] != nil {
		o.QuoteToken = order["quoteToken"].(string)
	}

	if order["price"] != 0 {
		o.Price = order["price"].(float64) // FIX!!!
	}

	if order["amount"] != nil {
		switch order["amount"].(type) {
		case float64:
			o.Amount = int64(order["amount"].(float64))
		case string:
			o.Amount, err = strconv.ParseInt(order["amount"].(string), 10, 64)
			if err != nil {
				return errors.New("failed to parse amount")
			}
		default:
			return errors.New("unrecognized type of amount")
		}
	}

	if order["filledAmount"] != nil {
		o.FilledAmount = int64(order["filledAmount"].(float64))
	}

	if order["remainingSellAmount"] != nil {
		o.RemainingSellAmount = int64(order["remainingSellAmount"].(float64))
	}

	if order["hash"] != nil {
		o.Hash = order["hash"].(string)
	}

	if order["side"] != nil {
		o.Side = order["side"].(string)
	}

	if order["status"] != nil {
		o.Status = order["status"].(string)
	}

	if order["originalOrder"] != nil {
		o.OriginalOrder = order["originalOrder"].(map[string]interface{})
	}

	if order["createdAt"] != nil {
		t, _ := time.Parse(time.RFC3339Nano, order["createdAt"].(string))
		o.CreatedAt = t
	}

	if order["updatedAt"] != nil {
		t, _ := time.Parse(time.RFC3339Nano, order["updatedAt"].(string))
		o.UpdatedAt = t
	}

	return nil
}

// OrderRecord is the object that will be saved in the database
type OrderRecord struct {
	ID                  bson.ObjectId `json:"id" bson:"_id"`
	UserAddress         string        `json:"userAddress" bson:"userAddress"`
	MatcherAddress      string        `json:"matcherAddress" bson:"matcherAddress"`
	AffiliateAddress    string        `json:"affiliateAddress" bson:"affiliateAddress"`
	BaseToken           string        `json:"baseToken" bson:"baseToken"`
	QuoteToken          string        `json:"quoteToken" bson:"quoteToken"`
	Status              string        `json:"status" bson:"status"`
	Side                string        `json:"side" bson:"side"`
	Hash                string        `json:"hash" bson:"hash"`
	Price               float64       `json:"price" bson:"price"`
	Amount              int64         `json:"amount" bson:"amount"`
	FilledAmount        int64         `json:"filledAmount" bson:"filledAmount"`
	RemainingSellAmount int64         `json:"remainingSellAmount" bson:"remainingSellAmount"`

	OriginalOrder map[string]interface{} `json:"originalOrder" bson:"originalOrder"`

	PairName  string    `json:"pairName" bson:"pairName"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

func (o *Order) GetBSON() (interface{}, error) {
	or := OrderRecord{
		PairName:            o.PairName,
		MatcherAddress:      o.MatcherAddress,
		AffiliateAddress:    o.AffiliateAddress,
		UserAddress:         o.UserAddress,
		BaseToken:           o.BaseToken,
		QuoteToken:          o.QuoteToken,
		Status:              o.Status,
		Side:                o.Side,
		Hash:                o.Hash,
		Amount:              o.Amount,
		RemainingSellAmount: o.RemainingSellAmount,
		Price:               o.Price,
		OriginalOrder:       o.OriginalOrder,
		CreatedAt:           o.CreatedAt,
		UpdatedAt:           o.UpdatedAt,
	}

	if o.ID == "" {
		or.ID = bson.NewObjectId()
	} else {
		or.ID = o.ID
	}

	if o.FilledAmount != 0 {
		or.FilledAmount = o.FilledAmount
	}

	return or, nil
}

func (o *Order) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID                  bson.ObjectId          `json:"id,omitempty" bson:"_id"`
		PairName            string                 `json:"pairName" bson:"pairName"`
		MatcherAddress      string                 `json:"matcherAddress" bson:"matcherAddress"`
		AffiliateAddress    string                 `json:"affiliateAddress" bson:"affiliateAddress"`
		UserAddress         string                 `json:"userAddress" bson:"userAddress"`
		BaseToken           string                 `json:"baseToken" bson:"baseToken"`
		QuoteToken          string                 `json:"quoteToken" bson:"quoteToken"`
		Status              string                 `json:"status" bson:"status"`
		Side                string                 `json:"side" bson:"side"`
		Hash                string                 `json:"hash" bson:"hash"`
		Price               float64                `json:"price" bson:"price"`
		Amount              int64                  `json:"amount" bson:"amount"`
		FilledAmount        int64                  `json:"filledAmount" bson:"filledAmount"`
		RemainingSellAmount int64                  `json:"remainingSellAmount" bson:"remainingSellAmount"`
		OriginalOrder       map[string]interface{} `json:"originalOrder" bson:"originalOrder"`
		CreatedAt           time.Time              `json:"createdAt" bson:"createdAt"`
		UpdatedAt           time.Time              `json:"updatedAt" bson:"updatedAt"`
	})

	err := raw.Unmarshal(decoded)
	if err != nil {
		logger.Error(err)
		return err
	}

	o.ID = decoded.ID
	o.PairName = decoded.PairName
	o.MatcherAddress = decoded.MatcherAddress
	o.AffiliateAddress = decoded.AffiliateAddress
	o.UserAddress = decoded.UserAddress
	o.BaseToken = decoded.BaseToken
	o.QuoteToken = decoded.QuoteToken
	o.FilledAmount = decoded.FilledAmount
	o.RemainingSellAmount = decoded.RemainingSellAmount
	o.Status = decoded.Status
	o.Side = decoded.Side
	o.Hash = decoded.Hash
	o.OriginalOrder = decoded.OriginalOrder

	if decoded.Amount != 0 {
		o.Amount = decoded.Amount
	}

	if decoded.FilledAmount != 0 {
		o.FilledAmount = decoded.FilledAmount
	}

	if decoded.Price != 0 {
		o.Price = decoded.Price
	}

	o.CreatedAt = decoded.CreatedAt
	o.UpdatedAt = decoded.UpdatedAt

	return nil
}

type OrderBSONUpdate struct {
	*Order
}

func (o OrderBSONUpdate) GetBSON() (interface{}, error) {
	now := time.Now()

	set := bson.M{
		"pairName":            o.PairName,
		"matcherAddress":      o.MatcherAddress,
		"affiliateAddress":    o.AffiliateAddress,
		"userAddress":         o.UserAddress,
		"baseToken":           o.BaseToken,
		"quoteToken":          o.QuoteToken,
		"status":              o.Status,
		"side":                o.Side,
		"price":               o.Price,
		"amount":              o.Amount,
		"remainingSellAmount": o.RemainingSellAmount,
		"originalOrder":       o.OriginalOrder,
		"updatedAt":           now,
	}

	if o.FilledAmount != 0 {
		set["filledAmount"] = o.FilledAmount
	}

	setOnInsert := bson.M{
		"_id":       bson.NewObjectId(),
		"hash":      o.Hash,
		"createdAt": now,
	}

	update := bson.M{
		"$set":         set,
		"$setOnInsert": setOnInsert,
	}

	return update, nil
}

type OrderData struct {
	Pair        PairID  `json:"id,omitempty" bson:"_id"`
	OrderVolume int64   `json:"orderVolume,omitempty" bson:"orderVolume"`
	OrderCount  int64   `json:"orderCount,omitempty" bson:"orderCount"`
	BestPrice   float64 `json:"bestPrice,omitempty" bson:"bestPrice"`
}

func (o *OrderData) AssetCode() string {
	code := o.Pair.BaseToken + "::" + o.Pair.QuoteToken
	return code
}

func (o *OrderData) ConvertedVolume(p *Pair, exchangeRate float64) float64 {
	valueAsToken := float64(o.OrderVolume) / float64(p.BaseTokenMultiplier())
	value := valueAsToken / exchangeRate

	return value
}

func (o *OrderData) MarshalJSON() ([]byte, error) {
	orderData := map[string]interface{}{
		"pair": map[string]interface{}{
			"pairName":   o.Pair.PairName,
			"baseToken":  o.Pair.BaseToken,
			"quoteToken": o.Pair.QuoteToken,
		},
	}

	if o.OrderVolume != 0 {
		orderData["orderVolume"] = o.OrderVolume
	}

	if o.OrderCount != 0 {
		orderData["orderCount"] = o.OrderCount
	}

	if o.BestPrice != 0 {
		orderData["bestPrice"] = o.BestPrice
	}

	bytes, err := json.Marshal(orderData)
	return bytes, err
}

// UnmarshalJSON creates a trade object from a json byte string
func (o *OrderData) UnmarshalJSON(b []byte) error {
	orderData := map[string]interface{}{}
	err := json.Unmarshal(b, &orderData)

	if err != nil {
		return err
	}

	if orderData["pair"] != nil {
		pair := orderData["pair"].(map[string]interface{})
		o.Pair = PairID{
			PairName:   pair["pairName"].(string),
			BaseToken:  pair["baseToken"].(string),
			QuoteToken: pair["quoteToken"].(string),
		}
	}

	if orderData["orderVolume"] != nil {
		o.OrderVolume = int64(orderData["orderVolume"].(float64))
	}

	if orderData["orderCount"] != nil {
		o.OrderCount = int64(orderData["orderCount"].(float64))
	}

	if orderData["bestPrice"] != nil {
		o.BestPrice = orderData["bestPrice"].(float64) // FIX
	}

	return nil
}

func (o *OrderData) GetBSON() (interface{}, error) {
	type PairID struct {
		PairName   string `json:"pairName" bson:"pairName"`
		BaseToken  string `json:"baseToken" bson:"baseToken"`
		QuoteToken string `json:"quoteToken" bson:"quoteToken"`
	}

	count := o.OrderCount

	volume := o.OrderVolume

	bestPrice := o.BestPrice

	return struct {
		ID          PairID  `json:"id,omitempty" bson:"_id"`
		OrderVolume int64   `json:"orderCount" bson:"orderCount"`
		OrderCount  int64   `json:"orderVolume" bson:"orderVolume"`
		BestPrice   float64 `json:"bestPrice" bson:"bestPrice"`
	}{
		ID: PairID{
			o.Pair.PairName,
			o.Pair.BaseToken,
			o.Pair.QuoteToken,
		},
		OrderVolume: volume,
		OrderCount:  count,
		BestPrice:   bestPrice,
	}, nil
}

func (o *OrderData) SetBSON(raw bson.Raw) error {
	type PairIDRecord struct {
		PairName   string `json:"pairName" bson:"pairName"`
		BaseToken  string `json:"baseToken" bson:"baseToken"`
		QuoteToken string `json:"quoteToken" bson:"quoteToken"`
	}

	decoded := new(struct {
		Pair        PairIDRecord `json:"pair,omitempty" bson:"_id"`
		OrderCount  int64        `json:"orderCount" bson:"orderCount"`
		OrderVolume int64        `json:"orderVolume" bson:"orderVolume"`
		BestPrice   float64      `json:"bestPrice" bson:"bestPrice"`
	})

	err := raw.Unmarshal(decoded)
	if err != nil {
		return err
	}

	o.Pair = PairID{
		PairName:   decoded.Pair.PairName,
		BaseToken:  decoded.Pair.BaseToken,
		QuoteToken: decoded.Pair.QuoteToken,
	}

	orderCount := decoded.OrderCount
	orderVolume := decoded.OrderVolume
	bestPrice := decoded.BestPrice

	o.OrderCount = orderCount
	o.OrderVolume = orderVolume
	o.BestPrice = bestPrice

	return nil
}
