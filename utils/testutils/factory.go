package testutils

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/byteball/odex-backend/types"
)

type Wallet struct {
	Address string
}

// Orderfactory simplifies creating orders, trades and cancelOrders objects
// Pair is the token pair for which the order is created
type OrderFactory struct {
	Wallet *Wallet
	Pair   *types.Pair
	Params *OrderParams
	// Client         *ethclient.Client
}

type OrderParams struct {
	MatcherAddress string
}

// NewOrderFactory returns an order factory from a given token pair and a given wallet
// TODO: Refactor this function to send back an error
func NewOrderFactory(p *types.Pair, w *Wallet, matcherAddress string) (*OrderFactory, error) {
	// rpcClient, err := rpc.DialWebsocket(context.Background(), "ws://127.0.0.1:8546", "")
	// if err != nil {
	// 	log.Printf("Could not create order factory")
	// 	return nil, err
	// }

	// client := ethclient.NewClient(rpcClient)

	params := &OrderParams{
		MatcherAddress: matcherAddress,
	}

	return &OrderFactory{
		Pair:   p,
		Wallet: w,
		Params: params,
		// Client:         client,
	}, nil
}

// GetWallet returns the order factory wallet
func (f *OrderFactory) GetWallet() *Wallet {
	return f.Wallet
}

// GetAddress returns the order factory address
func (f *OrderFactory) GetAddress() string {
	return f.Wallet.Address
}

func (f *OrderFactory) GetMatcherAddress() string {
	return f.Params.MatcherAddress
}

// SetMatcherAddress changes the default exchange address for orders created by this factory
func (f *OrderFactory) SetMatcherAddress(addr string) error {
	f.Params.MatcherAddress = addr
	return nil
}

// NewOrderMessage creates an order with the given params and returns a new PLACE_ORDER message
func (f *OrderFactory) NewOrderMessage(baseToken, quoteToken string, amount, pricepoint int64) (*types.WebsocketMessage, *types.Order, error) {
	o, err := f.NewOrder(baseToken, quoteToken, amount, pricepoint)
	if err != nil {
		return nil, nil, err
	}

	m := types.NewOrderWebsocketMessage(o)

	return m, o, nil
}

func (f *OrderFactory) NewCancelOrderMessage(o *types.Order) (*types.WebsocketMessage, *types.OrderCancel, error) {
	oc, err := f.NewCancelOrder(o)
	if err != nil {
		log.Print(err)
		return nil, nil, err
	}

	m := types.NewOrderCancelWebsocketMessage(oc)
	return m, oc, nil
}

// NewOrder returns a new order with the given params. The order is signed by the factory wallet.
func (f *OrderFactory) NewOrder(baseToken string, quoteToken string, amount int64, pricepoint int64) (*types.Order, error) {
	o := &types.Order{}

	o.UserAddress = f.Wallet.Address
	o.MatcherAddress = f.Params.MatcherAddress
	o.BaseToken = baseToken
	o.QuoteToken = quoteToken
	o.Price = float64(pricepoint)
	o.Amount = amount
	o.Status = "OPEN"
	//o.Sign(f.Wallet)

	return o, nil
}

func (f *OrderFactory) NewLargeOrder(baseToken string, quoteToken string, amount int64, pricepoint float64) (*types.Order, error) {
	o := &types.Order{}

	o.UserAddress = f.Wallet.Address
	o.MatcherAddress = f.Params.MatcherAddress
	o.BaseToken = baseToken
	o.QuoteToken = quoteToken
	o.Amount = amount
	o.Price = pricepoint
	o.Status = "OPEN"
	//o.Sign(f.Wallet)

	return o, nil
}

func (f *OrderFactory) NewBuyOrderMessage(price float64, amount int64) (*types.WebsocketMessage, *types.Order, error) {
	o, err := f.NewBuyOrder(price, amount)
	if err != nil {
		return nil, nil, err
	}

	m := types.NewOrderWebsocketMessage(&o)

	return m, &o, nil
}

func (f *OrderFactory) NewSellOrderMessage(price float64, amount int64) (*types.WebsocketMessage, *types.Order, error) {
	o, err := f.NewSellOrder(price, amount)
	if err != nil {
		return nil, nil, err
	}

	m := types.NewOrderWebsocketMessage(&o)

	return m, &o, nil
}

func (f *OrderFactory) NewCancelOrder(o *types.Order) (*types.OrderCancel, error) {
	oc := &types.OrderCancel{}

	oc.OrderHash = o.Hash
	//oc.Sign(f.Wallet)
	return oc, nil
}

// NewBuyOrder creates a new buy order from the order factory
func (f *OrderFactory) NewBuyOrder(pricepoint float64, value int64, filled ...int64) (types.Order, error) {
	o := types.Order{}

	// Transform decimal value into rounded points value (ex: 0.01 ETH => 1 ETH)
	amountPoints := int64(value * 100)
	etherPoints := int64(1)

	o.Amount = (etherPoints * amountPoints) / 100
	o.UserAddress = f.Wallet.Address
	o.MatcherAddress = f.Params.MatcherAddress
	o.BaseToken = f.Pair.BaseAsset
	o.QuoteToken = f.Pair.QuoteAsset
	o.Side = "BUY"

	o.PairName = f.Pair.Name()
	o.Price = float64(pricepoint)
	o.CreatedAt = time.Now()

	if filled == nil {
		o.FilledAmount = 0
		o.RemainingSellAmount = int64(math.Round(float64(o.Amount) * o.Price))
		o.Status = "OPEN"
	} else if value == filled[0] {
		o.FilledAmount = o.Amount
		o.RemainingSellAmount = 0
		o.Status = "FILLED"
	} else {
		filledPoints := int64(filled[0] * 100)
		o.FilledAmount = (etherPoints * filledPoints) / 100
		o.RemainingSellAmount = int64(math.Round(float64(o.Amount-o.FilledAmount) * o.Price))
		o.Status = "PARTIAL_FILLED"
	}

	o.OriginalOrder = map[string]interface{}{
		"signed_message": map[string]interface{}{
			"sell_amount": pricepoint * float64(value),
			"price":       1 / pricepoint,
			"matcher_fee": 0,
		},
	}

	//o.Sign(f.Wallet)
	o.Hash = fmt.Sprintf("%s %s %v %v", o.UserAddress, o.PairName, pricepoint, value)
	return o, nil
}

// NewBuyOrder returns a new order with the given params. The order is signed by the factory wallet
// NewBuyOrder computes the AmountBuy and AmountSell parameters from the given amount and price.
// Currently, the amount, price and order type are also kept. This could be amended in the future
// (meaning we would let the engine compute OrderBuy, Amount and Price. Ultimately this does not really
// matter except maybe for convenience/readability purposes)
func (f *OrderFactory) NewSellOrder(pricepoint float64, value int64, filled ...int64) (types.Order, error) {
	o := types.Order{}

	amountPoints := int64(value * 100)
	etherPoints := int64(1)

	o.Amount = (etherPoints * amountPoints) / 100
	o.UserAddress = f.Wallet.Address
	o.MatcherAddress = f.Params.MatcherAddress
	o.BaseToken = f.Pair.BaseAsset
	o.QuoteToken = f.Pair.QuoteAsset
	o.Side = "SELL"

	o.Price = float64(pricepoint)
	o.CreatedAt = time.Now()
	o.PairName = f.Pair.Name()

	if filled == nil {
		o.FilledAmount = 0
		o.RemainingSellAmount = o.Amount
		o.Status = "OPEN"
	} else if value == filled[0] {
		o.FilledAmount = o.Amount
		o.RemainingSellAmount = 0
		o.Status = "FILLED"
	} else {
		filledPoints := int64(filled[0])
		o.FilledAmount = etherPoints * filledPoints
		o.RemainingSellAmount = o.Amount - o.FilledAmount
		o.Status = "PARTIAL_FILLED"
	}

	o.OriginalOrder = map[string]interface{}{
		"signed_message": map[string]interface{}{
			"sell_amount": float64(value),
			"price":       pricepoint,
			"matcher_fee": 0,
		},
	}

	//o.Sign(f.Wallet)
	o.Hash = fmt.Sprintf("%s %s %v %v", o.UserAddress, o.PairName, pricepoint, value)
	return o, nil
}

// NewTrade returns a new trade with the given params. The trade is signed by the factory wallet.
func (f *OrderFactory) NewTrade(o *types.Order, amount int64) (types.Trade, error) {
	t := types.Trade{}

	t.Maker = o.UserAddress
	t.Taker = f.Wallet.Address
	t.BaseToken = o.BaseToken
	t.QuoteToken = o.QuoteToken
	t.MakerOrderHash = o.Hash
	t.Amount = amount
	t.QuoteAmount = t.CalcQuoteAmount()

	return t, nil
}
