package types

import (
	"encoding/json"
	"errors"
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
	var token struct {
		ID        *string  `json:"id"`
		Symbol    *string  `json:"symbol"`
		Asset     *string  `json:"asset"`
		Decimals  *float64 `json:"decimals"`
		Active    *bool    `json:"active"`
		Listed    *bool    `json:"listed"`
		Quote     *bool    `json:"quote"`
		Rank      *float64 `json:"rank"`
		CreatedAt *string  `json:"createdAt"`
		UpdatedAt *string  `json:"updatedAt"`
	}

	// type mismatches are returned as errors rather than panicking
	err := json.Unmarshal(b, &token)
	if err != nil {
		return err
	}

	if token.Asset != nil {
		t.Asset = *token.Asset
	}

	if token.Listed != nil {
		t.Listed = *token.Listed
	}

	if token.Quote != nil {
		t.Quote = *token.Quote
	}

	if token.Active != nil {
		t.Active = *token.Active
	}

	if token.Decimals != nil {
		t.Decimals = int(*token.Decimals)
	}

	if token.Symbol != nil {
		t.Symbol = *token.Symbol
	}

	if token.ID != nil && *token.ID != "" {
		if !bson.IsObjectIdHex(*token.ID) {
			return errors.New("invalid token id")
		}
		t.ID = bson.ObjectIdHex(*token.ID)
	}

	if token.CreatedAt != nil {
		tm, _ := time.Parse(time.RFC3339Nano, *token.CreatedAt)
		t.CreatedAt = tm
	}

	if token.UpdatedAt != nil {
		tm, _ := time.Parse(time.RFC3339Nano, *token.UpdatedAt)
		t.UpdatedAt = tm
	}

	if token.Rank != nil {
		t.Rank = int(*token.Rank)
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
