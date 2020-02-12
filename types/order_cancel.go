package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// OrderCancel is a group of params used for canceling an order previously
// sent to the matching engine. The OrderId and OrderHash must correspond to the
// same order. To be valid and be able to be processed by the matching engine,
// the OrderCancel must include a signature by the Maker of the order corresponding
// to the OrderHash.
type OrderCancel struct {
	OrderHash   string `json:"orderHash"`
	UserAddress string `json:"userAddress"`
}

// NewOrderCancel returns a new empty OrderCancel object
func NewOrderCancel() *OrderCancel {
	return &OrderCancel{
		UserAddress: "",
		OrderHash:   "",
	}
}

// MarshalJSON returns the json encoded byte array representing the OrderCancel struct
func (oc *OrderCancel) MarshalJSON() ([]byte, error) {
	orderCancel := map[string]interface{}{
		"orderHash":   oc.OrderHash,
		"userAddress": oc.UserAddress,
	}

	return json.Marshal(orderCancel)
}

func (oc *OrderCancel) String() string {
	return fmt.Sprintf("\nOrderCancel:\nOrderHash: %x\nUserAddress: %x\n\n",
		oc.OrderHash, oc.UserAddress)
}

// UnmarshalJSON creates an OrderCancel object from a json byte string
func (oc *OrderCancel) UnmarshalJSON(b []byte) error {
	parsed := map[string]interface{}{}

	err := json.Unmarshal(b, &parsed)
	if err != nil {
		return err
	}

	if parsed["orderHash"] == nil {
		return errors.New("Order Hash is missing")
	}
	oc.OrderHash = parsed["orderHash"].(string)

	if parsed["userAddress"] == nil {
		return errors.New("userAddress is missing")
	}
	oc.UserAddress = parsed["userAddress"].(string)

	return nil
}

/*
// ComputeHash computes the hash of an order cancel message
func (oc *OrderCancel) ComputeHash() string {
	sha := sha256.New()
	sha.Write([]byte(oc.OrderHash))
	return base64.StdEncoding.EncodeToString(sha.Sum(nil))
}
*/
