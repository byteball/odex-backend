package ws

import (
	"errors"
	"fmt"
)

const (
	TradeChannel        = "trades"
	RawOrderBookChannel = "raw_orderbook"
	OrderChannel        = "orders"
	OrderBookChannel    = "orderbook"
	OHLCVChannel        = "ohlcv"
	LoginChannel        = "login"
	BalancesChannel     = "balances"
)

var socketChannels map[string]func(interface{}, *Client)

func RegisterChannel(channel string, fn func(interface{}, *Client)) error {
	if channel == "" {
		return errors.New("Channel can not be an empty string")
	}

	if fn == nil {
		return errors.New("Handler should not be nil")
	}

	ch := getChannels()
	if ch[channel] != nil {
		return fmt.Errorf("Channel already registered")
	}

	ch[channel] = fn
	return nil
}

func getChannels() map[string]func(interface{}, *Client) {
	if socketChannels == nil {
		socketChannels = make(map[string]func(interface{}, *Client))
	}

	return socketChannels
}

// removeSubscription drops channelID from the client's list so that closed clients are not retained.
// The caller must hold the socket's lock.
func removeSubscription(list map[*Client][]string, c *Client, channelID string) {
	remaining := make([]string, 0, len(list[c]))
	for _, id := range list[c] {
		if id != channelID {
			remaining = append(remaining, id)
		}
	}

	if len(remaining) == 0 {
		delete(list, c)
	} else {
		list[c] = remaining
	}
}
