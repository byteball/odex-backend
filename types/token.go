package types

import (
	"encoding/json"
	"time"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Token struct is used to model the token data in the system and DB
type Token struct {
	ID       bson.ObjectId `json:"-" bson:"_id"`
	Symbol   string        `json:"symbol" bson:"symbol"`
	Asset    string        `json:"asset" bson:"asset"`
	Decimals int           `json:"decimals" bson:"decimals"`
	Active   bool          `json:"active" bson:"active"`
	Listed   bool          `json:"listed" bson:"listed"`
	Quote    bool          `json:"quote" bson:"quote"`
	Rank     int           `json:"rank,omitempty" bson:"rank,omitempty"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// TokenRecord is the struct which is stored in db
type TokenRecord struct {
	ID       bson.ObjectId `json:"-" bson:"_id"`
	Symbol   string        `json:"symbol" bson:"symbol"`
	Asset    string        `json:"asset" bson:"asset"`
	Decimals int           `json:"decimals" bson:"decimals"`
	Active   bool          `json:"active" bson:"active"`
	Listed   bool          `json:"listed" bson:"listed"`
	Quote    bool          `json:"quote" bson:"quote"`
	Rank     int           `json:"rank,omitempty" bson:"rank,omitempty"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// Validate function is used to verify if an instance of
// struct satisfies all the conditions for a valid instance
func (t Token) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Symbol, validation.Required),
		validation.Field(&t.Asset, validation.Required),
	)
}

func (t *Token) MarshalJSON() ([]byte, error) {
	token := map[string]interface{}{
		"id":        t.ID,
		"symbol":    t.Symbol,
		"asset":     t.Asset,
		"decimals":  t.Decimals,
		"active":    t.Active,
		"listed":    t.Listed,
		"quote":     t.Quote,
		"createdAt": t.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt": t.UpdatedAt.Format(time.RFC3339Nano),
		"rank":      t.Rank,
	}

	return json.Marshal(token)
}

func (t *Token) UnmarshalJSON(b []byte) error {
	token := map[string]interface{}{}

	err := json.Unmarshal(b, &token)
	if err != nil {
		return err
	}

	if token["asset"] != nil {
		t.Asset = token["asset"].(string)
	}

	if token["listed"] != nil {
		t.Listed = token["listed"].(bool)
	}

	if token["quote"] != nil {
		t.Quote = token["quote"].(bool)
	}

	if token["active"] != nil {
		t.Active = token["active"].(bool)
	}

	if token["decimals"] != nil {
		t.Decimals = int(token["decimals"].(float64))
	}

	if token["symbol"] != nil {
		t.Symbol = token["symbol"].(string)
	}

	if token["id"] != nil && token["id"].(string) != "" {
		t.ID = bson.ObjectIdHex(token["id"].(string))
	}

	if token["createdAt"] != nil {
		tm, _ := time.Parse(time.RFC3339Nano, token["createdAt"].(string))
		t.CreatedAt = tm
	}

	if token["updatedAt"] != nil {
		tm, _ := time.Parse(time.RFC3339Nano, token["updatedAt"].(string))
		t.UpdatedAt = tm
	}

	if token["rank"] != nil {
		t.Rank = int(token["rank"].(float64))
	}

	return nil
}

// GetBSON implements bson.Getter
func (t *Token) GetBSON() (interface{}, error) {
	tr := TokenRecord{
		ID:        t.ID,
		Symbol:    t.Symbol,
		Asset:     t.Asset,
		Decimals:  t.Decimals,
		Active:    t.Active,
		Listed:    t.Listed,
		Quote:     t.Quote,
		Rank:      t.Rank,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	return tr, nil
}

// SetBSON implemenets bson.Setter
func (t *Token) SetBSON(raw bson.Raw) error {
	decoded := &TokenRecord{}

	err := raw.Unmarshal(decoded)
	if err != nil {
		return err
	}

	t.ID = decoded.ID
	t.Symbol = decoded.Symbol
	if decoded.Asset != "" {
		t.Asset = decoded.Asset
	}

	t.Decimals = decoded.Decimals
	t.Active = decoded.Active
	t.Listed = decoded.Listed
	t.Quote = decoded.Quote
	t.CreatedAt = decoded.CreatedAt
	t.UpdatedAt = decoded.UpdatedAt
	t.Rank = decoded.Rank

	return nil
}
