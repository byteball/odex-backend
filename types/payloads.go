package types

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation"
)

// NewOrderPayload is the struct in which the order request sent by the
// user is populated
type NewOrderPayload struct {
	PairName         string  `json:"pairName"`
	MatcherAddress   string  `json:"matcherAddress"`
	AffiliateAddress string  `json:"affiliateAddress"`
	UserAddress      string  `json:"userAddress"`
	BaseToken        string  `json:"baseToken"`
	QuoteToken       string  `json:"quoteToken"`
	Side             string  `json:"side"`
	Amount           int64   `json:"amount"`
	Price            float64 `json:"price"`
	Hash             string  `json:"hash"`
}

func (p NewOrderPayload) MarshalJSON() ([]byte, error) {
	encoded := map[string]interface{}{
		"pairName":         p.PairName,
		"matcherAddress":   p.MatcherAddress,
		"affiliateAddress": p.AffiliateAddress,
		"userAddress":      p.UserAddress,
		"amount":           p.Amount,
		"price":            p.Price,
		"side":             p.Side,
		"hash":             p.Hash,
	}

	return json.Marshal(encoded)
}

func (p *NewOrderPayload) UnmarshalJSON(b []byte) error {
	decoded := map[string]interface{}{}

	err := json.Unmarshal(b, &decoded)
	if err != nil {
		return err
	}

	if decoded["pairName"] != nil {
		p.PairName = decoded["pairName"].(string)
	}

	if decoded["userAddress"] != nil {
		p.UserAddress = decoded["userAddress"].(string)
	}

	if decoded["matcherAddress"] != nil {
		p.MatcherAddress = decoded["matcherAddress"].(string)
	}

	if decoded["affiliateAddress"] != nil {
		p.AffiliateAddress = decoded["affiliateAddress"].(string)
	}

	if decoded["amount"] != nil {
		p.Amount = int64(decoded["amount"].(float64))
	}

	if decoded["price"] != nil {
		p.Price = decoded["price"].(float64) // FIX
	}

	if decoded["side"] != nil {
		p.Side = decoded["side"].(string)
	}

	if decoded["hash"] != nil {
		p.Hash = decoded["hash"].(string)
	}
	return nil
}

// Validate validates the NewOrderPayload fields.
func (p NewOrderPayload) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Amount, validation.Required),
		validation.Field(&p.Price, validation.Required),
		validation.Field(&p.UserAddress, validation.Required),
		validation.Field(&p.BaseToken, validation.Required),
		validation.Field(&p.QuoteToken, validation.Required),
		validation.Field(&p.Side, validation.Required),
		// validation.Field(&m.Signature, validation.Required),
	)
}

// ToOrder converts the NewOrderPayload to Order
func (p *NewOrderPayload) ToOrder() (o *Order, err error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	o = &Order{
		UserAddress: p.UserAddress,
		BaseToken:   p.BaseToken,
		QuoteToken:  p.QuoteToken,
		Amount:      p.Amount,
		Side:        p.Side,
		Price:       p.Price,
		//	Hash:        p.ComputeHash(),
	}

	return o, nil
}

func (p *NewOrderPayload) EncodedSide() int64 {
	if p.Side == "BUY" {
		return 0
	} else {
		return 1
	}
}

/*
// ComputeHash calculates the orderRequest hash
func (p *NewOrderPayload) ComputeHash() string {
	sha := sha256.New()
	sha.Write([]byte(p.MatcherAddress))
	sha.Write([]byte(p.AffiliateAddress))
	sha.Write([]byte(p.UserAddress))
	sha.Write([]byte(p.BaseToken))
	sha.Write([]byte(p.QuoteToken))
	sha.Write(p.Amount.Bytes())
	sha.Write(p.PricePoint.Bytes())
	sha.Write(p.EncodedSide().Bytes())
	sha.Write(p.Nonce.Bytes())
	return base64.StdEncoding.EncodeToString(sha.Sum(nil))
}
*/
