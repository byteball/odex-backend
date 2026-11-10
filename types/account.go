package types

import (
	"time"

	"encoding/json"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Account corresponds to a single Obyte address. It contains a list of token balances for that address
type Account struct {
	ID            bson.ObjectId            `json:"-" bson:"_id"`
	Address       string                   `json:"address" bson:"address"`
	TokenBalances map[string]*TokenBalance `json:"tokenBalances" bson:"tokenBalances"`
	IsBlocked     bool                     `json:"isBlocked" bson:"isBlocked"`
	CreatedAt     time.Time                `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time                `json:"updatedAt" bson:"updatedAt"`
}

// TokenBalance holds the Balance and the Locked balance values for a single asset
type TokenBalance struct {
	Asset          string `json:"asset" bson:"asset"`
	Symbol         string `json:"symbol" bson:"symbol"`
	Balance        int64  `json:"balance" bson:"balance"`
	PendingBalance int64  `json:"pendingBalance" bson:"pendingBalance"`
	LockedBalance  int64  `json:"lockedBalance" bson:"lockedBalance"`
}

// AccountRecord corresponds to what is stored in the DB.
type AccountRecord struct {
	ID            bson.ObjectId                 `json:"id" bson:"_id"`
	Address       string                        `json:"address" bson:"address"`
	TokenBalances map[string]TokenBalanceRecord `json:"tokenBalances" bson:"tokenBalances"`
	IsBlocked     bool                          `json:"isBlocked" bson:"isBlocked"`
	CreatedAt     time.Time                     `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time                     `json:"updatedAt" bson:"updatedAt"`
}

// TokenBalanceRecord corresponds to a TokenBalance struct that is stored in the DB.
type TokenBalanceRecord struct {
	Asset          string `json:"asset" bson:"asset"`
	Symbol         string `json:"symbol" bson:"symbol"`
	Balance        int64  `json:"balance" bson:"balance"`
	PendingBalance int64  `json:"pendingBalance" base:"pendingBalance"`
	LockedBalance  int64  `json:"lockedBalance" bson:"lockedBalance"`
}

// GetBSON implements bson.Getter
func (a *Account) GetBSON() (interface{}, error) {
	ar := AccountRecord{
		IsBlocked: a.IsBlocked,
		Address:   a.Address,
	}

	tokenBalances := make(map[string]TokenBalanceRecord)

	for key, value := range a.TokenBalances {
		tokenBalances[key] = TokenBalanceRecord{
			Asset:          value.Asset,
			Symbol:         value.Symbol,
			Balance:        value.Balance,
			LockedBalance:  value.LockedBalance,
			PendingBalance: value.PendingBalance,
		}
	}

	ar.TokenBalances = tokenBalances

	if a.ID == "" {
		ar.ID = bson.NewObjectId()
	} else {
		ar.ID = a.ID
	}

	return ar, nil
}

// SetBSON implemenets bson.Setter
func (a *Account) SetBSON(raw bson.Raw) error {
	decoded := &AccountRecord{}

	err := raw.Unmarshal(decoded)
	if err != nil {
		return err
	}

	a.TokenBalances = make(map[string]*TokenBalance)
	for key, value := range decoded.TokenBalances {

		balance := value.Balance
		lockedBalance := value.LockedBalance
		pendingBalance := value.PendingBalance

		a.TokenBalances[key] = &TokenBalance{
			Asset:          value.Asset,
			Symbol:         value.Symbol,
			Balance:        balance,
			LockedBalance:  lockedBalance,
			PendingBalance: pendingBalance,
		}
	}

	a.Address = decoded.Address
	a.ID = decoded.ID
	a.IsBlocked = decoded.IsBlocked
	a.CreatedAt = decoded.CreatedAt
	a.UpdatedAt = decoded.UpdatedAt

	return nil
}

// JSON Marshal/Unmarshal interface

// MarshalJSON implements the json.Marshal interface
func (a *Account) MarshalJSON() ([]byte, error) {
	account := map[string]interface{}{
		"id":        a.ID,
		"address":   a.Address,
		"isBlocked": a.IsBlocked,
		"createdAt": a.CreatedAt.String(),
		"updatedAt": a.UpdatedAt.String(),
	}

	tokenBalance := make(map[string]interface{})

	for address, balance := range a.TokenBalances {
		tokenBalance[address] = map[string]interface{}{
			"asset":          balance.Asset,
			"symbol":         balance.Symbol,
			"balance":        balance.Balance,
			"lockedBalance":  balance.LockedBalance,
			"pendingBalance": balance.PendingBalance,
		}
	}

	account["tokenBalances"] = tokenBalance
	return json.Marshal(account)
}

func (a *Account) UnmarshalJSON(b []byte) error {
	var account struct {
		ID            *string `json:"id"`
		Address       *string `json:"address"`
		TokenBalances map[string]*struct {
			Asset          *string  `json:"asset"`
			Symbol         *string  `json:"symbol"`
			Balance        *float64 `json:"balance"`
			LockedBalance  *float64 `json:"lockedBalance"`
			PendingBalance *float64 `json:"pendingBalance"`
		} `json:"tokenBalances"`
	}

	err := json.Unmarshal(b, &account)
	if err != nil {
		return err
	}

	if account.ID != nil && bson.IsObjectIdHex(*account.ID) {
		a.ID = bson.ObjectIdHex(*account.ID)
	}

	if account.Address != nil {
		a.Address = *account.Address
	}

	if account.TokenBalances != nil {
		a.TokenBalances = make(map[string]*TokenBalance)
		for asset, balance := range account.TokenBalances {
			if !isValidAsset(asset) || balance == nil {
				continue
			}

			tb := &TokenBalance{}

			if balance.Asset != nil && isValidAsset(*balance.Asset) {
				tb.Asset = *balance.Asset
			}

			if balance.Symbol != nil {
				tb.Symbol = *balance.Symbol
			}

			if balance.Balance != nil {
				tb.Balance = int64(*balance.Balance)
			}

			if balance.LockedBalance != nil {
				tb.LockedBalance = int64(*balance.LockedBalance)
			}

			if balance.PendingBalance != nil {
				tb.PendingBalance = int64(*balance.PendingBalance)
			}

			a.TokenBalances[asset] = tb
		}
	}

	return nil
}

// Validate enforces the account model
func (a Account) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Address, validation.Required),
	)
}

type AccountBSONUpdate struct {
	*Account
}

func (a *AccountBSONUpdate) GetBSON() (interface{}, error) {
	now := time.Now()
	tokenBalances := make(map[string]TokenBalanceRecord)

	//TODO validate this. All the fields have to be set
	for key, value := range a.TokenBalances {
		tokenBalances[key] = TokenBalanceRecord{
			Asset:          value.Asset,
			Symbol:         value.Symbol,
			Balance:        value.Balance,
			LockedBalance:  value.LockedBalance,
			PendingBalance: value.PendingBalance,
		}
	}

	set := bson.M{
		"updatedAt": now,
		"address":   a.Address,
	}

	setOnInsert := bson.M{
		"_id":       bson.NewObjectId(),
		"createdAt": now,
	}

	update := bson.M{
		"$set":         set,
		"$setOnInsert": setOnInsert,
	}

	return update, nil
}

func isValidAsset(asset string) bool {
	return len(asset) == 44 || asset == "base"
}
