package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/globalsign/mgo/bson"
)

// Trade struct holds arguments corresponding to a "Taker Order"
type Trade struct {
	ID                       bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Taker                    string        `json:"taker" bson:"taker"`
	Maker                    string        `json:"maker" bson:"maker"`
	BaseToken                string        `json:"baseToken" bson:"baseToken"`
	QuoteToken               string        `json:"quoteToken" bson:"quoteToken"`
	MakerOrderHash           string        `json:"makerOrderHash" bson:"makerOrderHash"`
	TakerOrderHash           string        `json:"takerOrderHash" bson:"takerOrderHash"`
	Hash                     string        `json:"hash" bson:"hash"`
	TxHash                   string        `json:"txHash" bson:"txHash"`
	PairName                 string        `json:"pairName" bson:"pairName"`
	CreatedAt                time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt                time.Time     `json:"updatedAt" bson:"updatedAt"`
	Price                    float64       `json:"price" bson:"price"`
	Status                   string        `json:"status" bson:"status"`
	Amount                   int64         `json:"amount" bson:"amount"`
	QuoteAmount              int64         `json:"quoteAmount" bson:"quoteAmount"`
	RemainingTakerSellAmount int64         `json:"remainingTakerSellAmount" bson:"remainingTakerSellAmount"`
	RemainingMakerSellAmount int64         `json:"remainingMakerSellAmount" bson:"remainingMakerSellAmount"`
	MakerSide                string        `json:"makerSide" bson:"makerSide"`
}

type TradeRecord struct {
	ID                       bson.ObjectId `json:"id" bson:"_id"`
	Taker                    string        `json:"taker" bson:"taker"`
	Maker                    string        `json:"maker" bson:"maker"`
	BaseToken                string        `json:"baseToken" bson:"baseToken"`
	QuoteToken               string        `json:"quoteToken" bson:"quoteToken"`
	MakerOrderHash           string        `json:"makerOrderHash" bson:"makerOrderHash"`
	TakerOrderHash           string        `json:"takerOrderHash" bson:"takerOrderHash"`
	Hash                     string        `json:"hash" bson:"hash"`
	TxHash                   string        `json:"txHash" bson:"txHash"`
	PairName                 string        `json:"pairName" bson:"pairName"`
	CreatedAt                time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt                time.Time     `json:"updatedAt" bson:"updatedAt"`
	Price                    float64       `json:"price" bson:"price"`
	Amount                   int64         `json:"amount" bson:"amount"`
	QuoteAmount              int64         `json:"quoteAmount" bson:"quoteAmount"`
	RemainingTakerSellAmount int64         `json:"remainingTakerSellAmount" bson:"remainingTakerSellAmount"`
	RemainingMakerSellAmount int64         `json:"remainingMakerSellAmount" bson:"remainingMakerSellAmount"`
	Status                   string        `json:"status" bson:"status"`
	MakerSide                string        `json:"makerSide" bson:"makerSide"`
}

// NewTrade returns a new unsigned trade corresponding to an Order, amount and taker address
func NewTrade(mo *Order, to *Order, amount int64, price float64) *Trade {
	t := &Trade{
		Maker:          mo.UserAddress,
		Taker:          to.UserAddress,
		BaseToken:      mo.BaseToken,
		QuoteToken:     mo.QuoteToken,
		MakerOrderHash: mo.Hash,
		TakerOrderHash: to.Hash,
		PairName:       mo.PairName,
		Amount:         amount,
		Price:          price,
		Status:         "PENDING",
	}

	t.QuoteAmount = t.CalcQuoteAmount()
	t.Hash = t.ComputeHash()

	return t
}

func (t *Trade) Validate() error {
	if t.Taker == "" {
		return errors.New("Trade 'taker' parameter is required'")
	}

	if t.Maker == "" {
		return errors.New("Trade 'maker' parameter is required")
	}

	if t.TakerOrderHash == "" {
		return errors.New("Trade 'takerOrderHash' parameter is required")
	}

	if t.MakerOrderHash == "" {
		return errors.New("Trade 'makerOrderHash' parameter is required")
	}

	if t.BaseToken == "" {
		return errors.New("Trade 'baseToken' parameter is required")
	}

	if t.QuoteToken == "" {
		return errors.New("Trade 'quoteToken' parameter is required")
	}

	if t.Amount == 0 {
		return errors.New("Trade 'amount' parameter is required")
	}

	if t.QuoteAmount == 0 {
		return errors.New("Trade 'quoteAmount' parameter is required")
	}

	if t.Price == 0 {
		return errors.New("Trade 'price' paramter is required")
	}

	if t.Price <= 0 {
		return errors.New("Trade 'price' parameter should be positive")
	}

	if t.Amount <= 0 {
		return errors.New("Trade 'amount' parameter should be positive")
	}

	if t.QuoteAmount <= 0 {
		return errors.New("Trade 'quoteAmount' parameter should be positive")
	}

	//TODO add validations for hashes and addresses
	return nil
}

// MarshalJSON returns the json encoded byte array representing the trade struct
func (t *Trade) MarshalJSON() ([]byte, error) {
	trade := map[string]interface{}{
		"taker":                    t.Taker,
		"maker":                    t.Maker,
		"status":                   t.Status,
		"hash":                     t.Hash,
		"pairName":                 t.PairName,
		"price":                    t.Price,
		"amount":                   t.Amount,
		"quoteAmount":              t.QuoteAmount,
		"remainingTakerSellAmount": t.RemainingTakerSellAmount,
		"remainingMakerSellAmount": t.RemainingMakerSellAmount,
		"makerSide":                t.MakerSide,
		"createdAt":                t.CreatedAt.Format(time.RFC3339Nano),
	}

	if t.BaseToken != "" {
		trade["baseToken"] = t.BaseToken
	}

	if t.QuoteToken != "" {
		trade["quoteToken"] = t.QuoteToken
	}

	if t.TxHash != "" {
		trade["txHash"] = t.TxHash
	}

	if t.TakerOrderHash != "" {
		trade["takerOrderHash"] = t.TakerOrderHash
	}

	if t.MakerOrderHash != "" {
		trade["makerOrderHash"] = t.MakerOrderHash
	}

	return json.Marshal(trade)
}

// UnmarshalJSON creates a trade object from a json byte string
func (t *Trade) UnmarshalJSON(b []byte) error {
	trade := map[string]interface{}{}

	err := json.Unmarshal(b, &trade)
	if err != nil {
		return err
	}

	if trade["makerOrderHash"] == nil {
		return errors.New("Order Hash is not set")
	} else {
		t.MakerOrderHash = trade["makerOrderHash"].(string)
	}

	if trade["takerOrderHash"] != nil {
		t.TakerOrderHash = trade["takerOrderHash"].(string)
	}

	if trade["hash"] == nil {
		return errors.New("Hash is not set")
	} else {
		t.Hash = trade["hash"].(string)
	}

	if trade["quoteToken"] == nil {
		return errors.New("Quote token is not set")
	} else {
		t.QuoteToken = trade["quoteToken"].(string)
	}

	if trade["baseToken"] == nil {
		return errors.New("Base token is not set")
	} else {
		t.BaseToken = trade["baseToken"].(string)
	}

	if trade["maker"] == nil {
		return errors.New("Maker is not set")
	} else {
		t.Taker = trade["taker"].(string)
	}

	if trade["taker"] == nil {
		return errors.New("Taker is not set")
	} else {
		t.Maker = trade["maker"].(string)
	}

	if trade["id"] != nil && bson.IsObjectIdHex(trade["id"].(string)) {
		t.ID = bson.ObjectIdHex(trade["id"].(string))
	}

	if trade["txHash"] != nil {
		t.TxHash = trade["txHash"].(string)
	}

	if trade["pairName"] != nil {
		t.PairName = trade["pairName"].(string)
	}

	if trade["status"] != nil {
		t.Status = trade["status"].(string)
	}

	if trade["price"] != nil {
		t.Price = trade["price"].(float64) // FIX
	}

	if trade["amount"] != nil {
		t.Amount = int64(trade["amount"].(float64))
	}

	if trade["quoteAmount"] != nil {
		t.QuoteAmount = int64(trade["quoteAmount"].(float64))
	}

	if trade["remainingTakerSellAmount"] != nil {
		t.RemainingTakerSellAmount = int64(trade["remainingTakerSellAmount"].(float64))
	}

	if trade["remainingMakerSellAmount"] != nil {
		t.RemainingMakerSellAmount = int64(trade["remainingMakerSellAmount"].(float64))
	}

	if trade["createdAt"] != nil {
		tm, _ := time.Parse(time.RFC3339Nano, trade["createdAt"].(string))
		t.CreatedAt = tm
	}

	if trade["makerSide"] != nil {
		t.MakerSide = trade["makerSide"].(string)
	}

	return nil
}

func (t *Trade) CalcQuoteAmount() int64 {
	// pairMultiplier := p.PairMultiplier()
	return int64(math.Round(float64(t.Amount) * t.Price))
}

func (t *Trade) GetBSON() (interface{}, error) {
	tr := TradeRecord{
		ID:                       t.ID,
		PairName:                 t.PairName,
		Maker:                    t.Maker,
		Taker:                    t.Taker,
		BaseToken:                t.BaseToken,
		QuoteToken:               t.QuoteToken,
		MakerOrderHash:           t.MakerOrderHash,
		Hash:                     t.Hash,
		TxHash:                   t.TxHash,
		TakerOrderHash:           t.TakerOrderHash,
		CreatedAt:                t.CreatedAt,
		UpdatedAt:                t.UpdatedAt,
		Price:                    t.Price,
		Status:                   t.Status,
		Amount:                   t.Amount,
		QuoteAmount:              t.QuoteAmount,
		RemainingTakerSellAmount: t.RemainingTakerSellAmount,
		RemainingMakerSellAmount: t.RemainingMakerSellAmount,
		MakerSide:                t.MakerSide,
	}

	return tr, nil
}

func (t *Trade) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID                       bson.ObjectId `json:"id,omitempty" bson:"_id"`
		PairName                 string        `json:"pairName" bson:"pairName"`
		Taker                    string        `json:"taker" bson:"taker"`
		Maker                    string        `json:"maker" bson:"maker"`
		BaseToken                string        `json:"baseToken" bson:"baseToken"`
		QuoteToken               string        `json:"quoteToken" bson:"quoteToken"`
		MakerOrderHash           string        `json:"makerOrderHash" bson:"makerOrderHash"`
		TakerOrderHash           string        `json:"takerOrderHash" bson:"takerOrderHash"`
		Hash                     string        `json:"hash" bson:"hash"`
		TxHash                   string        `json:"txHash" bson:"txHash"`
		CreatedAt                time.Time     `json:"createdAt" bson:"createdAt"`
		UpdatedAt                time.Time     `json:"updatedAt" bson:"updatedAt"`
		Price                    float64       `json:"price" bson:"price"`
		Status                   string        `json:"status" bson:"status"`
		Amount                   int64         `json:"amount" bson:"amount"`
		QuoteAmount              int64         `json:"quoteAmount" bson:"quoteAmount"`
		RemainingTakerSellAmount int64         `json:"remainingTakerSellAmount" bson:"remainingTakerSellAmount"`
		RemainingMakerSellAmount int64         `json:"remainingMakerSellAmount" bson:"remainingMakerSellAmount"`
		MakerSide                string        `json:"makerSide" bson:"makerSide"`
	})

	err := raw.Unmarshal(decoded)
	if err != nil {
		return err
	}

	t.ID = decoded.ID
	t.PairName = decoded.PairName
	t.Taker = decoded.Taker
	t.Maker = decoded.Maker
	t.BaseToken = decoded.BaseToken
	t.QuoteToken = decoded.QuoteToken
	t.MakerOrderHash = decoded.MakerOrderHash
	t.TakerOrderHash = decoded.TakerOrderHash
	t.Hash = decoded.Hash
	t.TxHash = decoded.TxHash
	t.Status = decoded.Status
	t.Amount = decoded.Amount
	t.QuoteAmount = decoded.QuoteAmount
	t.RemainingTakerSellAmount = decoded.RemainingTakerSellAmount
	t.RemainingMakerSellAmount = decoded.RemainingMakerSellAmount
	t.Price = decoded.Price
	t.MakerSide = decoded.MakerSide

	t.CreatedAt = decoded.CreatedAt
	t.UpdatedAt = decoded.UpdatedAt
	return nil
}

// ComputeHash returns hashes the trade
// The OrderHash, Amount, and Taker attributes must be
// set before attempting to compute the trade hash
func (t *Trade) ComputeHash() string {
	sha := sha256.New()
	sha.Write([]byte(t.MakerOrderHash))
	sha.Write([]byte(t.TakerOrderHash))
	return base64.StdEncoding.EncodeToString(sha.Sum(nil))
}

func (t *Trade) Pair() (*Pair, error) {
	if t.BaseToken == "" {
		return nil, errors.New("Base token is not set")
	}

	if t.QuoteToken == "" {
		return nil, errors.New("Quote token is set")
	}

	return &Pair{
		BaseAsset:  t.BaseToken,
		QuoteAsset: t.QuoteToken,
	}, nil
}

type TradeBSONUpdate struct {
	*Trade
}

func (t TradeBSONUpdate) GetBSON() (interface{}, error) {
	now := time.Now()

	set := bson.M{
		"taker":                    t.Taker,
		"maker":                    t.Maker,
		"baseToken":                t.BaseToken,
		"quoteToken":               t.QuoteToken,
		"makerOrderHash":           t.MakerOrderHash,
		"takerOrderHash":           t.TakerOrderHash,
		"txHash":                   t.TxHash,
		"pairName":                 t.PairName,
		"status":                   t.Status,
		"remainingTakerSellAmount": t.RemainingTakerSellAmount,
		"remainingMakerSellAmount": t.RemainingMakerSellAmount,
		"makerSide":                t.MakerSide,
	}

	if t.Price != 0 {
		set["price"] = t.Price
	}

	if t.Amount != 0 {
		set["amount"] = t.Amount
	}

	if t.QuoteAmount != 0 {
		set["quoteAmount"] = t.QuoteAmount
	}

	setOnInsert := bson.M{
		"_id":       bson.NewObjectId(),
		"hash":      t.Hash,
		"createdAt": now,
	}

	update := bson.M{
		"$set":         set,
		"$setOnInsert": setOnInsert,
	}

	return update, nil
}
