package interfaces

import (
	"time"

	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/ws"
	"github.com/globalsign/mgo/bson"
)

type OrderDao interface {
	Create(o *types.Order) error
	Update(id bson.ObjectId, o *types.Order) error
	Upsert(id bson.ObjectId, o *types.Order) error
	Delete(orders ...*types.Order) error
	DeleteByHashes(hashes ...string) error
	UpdateAllByHash(h string, o *types.Order) error
	UpdateByHash(h string, o *types.Order) error
	UpsertByHash(h string, o *types.Order) error
	GetByID(id bson.ObjectId) (*types.Order, error)
	GetByHash(h string) (*types.Order, error)
	GetByHashes(hashes []string) ([]*types.Order, error)

	GetByUserAddress(addr string, limit ...int) ([]*types.Order, error)
	GetCurrentByUserAddress(a string, limit ...int) ([]*types.Order, error)
	GetCurrentByUserAddressAndSignerAddress(address string, signer string) ([]*types.Order, error)
	GetHistoryByUserAddress(a string, limit ...int) ([]*types.Order, error)
	GetMatchingBuyOrders(o *types.Order) ([]*types.Order, error)
	GetMatchingSellOrders(o *types.Order) ([]*types.Order, error)
	GetExpiredOrders() ([]*types.Order, error)
	UpdateOrderFilledAmount(h string, value int64) error
	UpdateOrderFilledAmounts(h []string, values []int64) ([]*types.Order, error)
	UpdateOrderStatusesByHashes(status string, hashes ...string) ([]*types.Order, error)
	GetUserLockedBalance(account string, token string) (int64, []*types.Order, error)
	UpdateOrderStatus(h string, status string) error
	GetRawOrderBook(*types.Pair) ([]*types.Order, error)
	GetOrderBook(*types.Pair) ([]map[string]interface{}, []map[string]interface{}, error)
	GetOrderBookPrice(p *types.Pair, pp float64, side string) (int64, string, float64, error)
	FindAndModify(h string, o *types.Order) (*types.Order, error)
	Drop() error
	Aggregate(q []bson.M) ([]*types.OrderData, error)
}

type AccountDao interface {
	Create(account *types.Account) (err error)
	GetAll() (res []types.Account, err error)
	GetByID(id bson.ObjectId) (*types.Account, error)
	GetByAddress(owner string) (response *types.Account, err error)
	GetTokenBalances(owner string) (map[string]*types.TokenBalance, error)
	GetTokenBalance(owner string, token string) (*types.TokenBalance, error)
	UpdateTokenBalance(owner string, token string, tokenBalance *types.TokenBalance) (err error)
	UpdateBalance(owner string, token string, balance int64) (err error)
	FindOrCreate(addr string) (*types.Account, error)
	Drop()
}

type PairDao interface {
	Create(o *types.Pair) error
	GetAll() ([]types.Pair, error)
	GetActivePairs() ([]types.Pair, error)
	GetByID(id bson.ObjectId) (*types.Pair, error)
	GetByName(name string) (*types.Pair, error)
	GetByTokenSymbols(baseTokenSymbol, quoteTokenSymbol string) (*types.Pair, error)
	GetByAsset(baseToken, quoteToken string) (*types.Pair, error)
	GetDefaultPairs() ([]types.Pair, error)
	GetListedPairs() ([]types.Pair, error)
	GetUnlistedPairs() ([]types.Pair, error)
}

type TradeDao interface {
	Create(o ...*types.Trade) error
	Update(t *types.Trade) error
	UpdateByHash(h string, t *types.Trade) error
	GetAll() ([]types.Trade, error)
	Aggregate(q []bson.M) ([]*types.Tick, error)
	GetByPairName(name string) ([]*types.Trade, error)
	GetErroredTradeCount(start, end time.Time) (int, error)
	GetByHash(h string) (*types.Trade, error)
	GetByMakerOrderHash(h string) ([]*types.Trade, error)
	GetByTakerOrderHash(h string) ([]*types.Trade, error)
	GetByTriggerUnitHash(h string) ([]*types.Trade, error)
	GetByHashes(hashes []string) ([]*types.Trade, error)
	GetByOrderHashes(hashes []string) ([]*types.Trade, error)
	GetSortedTrades(bt, qt string, n int) ([]*types.Trade, error)
	GetSortedTradesByUserAddress(a string, limit ...int) ([]*types.Trade, error)
	GetUncommittedTradesByUserAddress(a string) []*types.Trade
	GetNTradesByPairAssets(bt, qt string, n int) ([]*types.Trade, error)
	GetTradesByPairAssets(bt, qt string, n int) ([]*types.Trade, error)
	GetAllTradesByPairAssets(bt, qt string) ([]*types.Trade, error)
	FindAndModify(h string, t *types.Trade) (*types.Trade, error)
	GetByUserAddress(a string) ([]*types.Trade, error)
	UpdateTradeStatus(h string, status string) error
	UpdateTradeStatuses(status string, hashes ...string) ([]*types.Trade, error)
	UpdateTradeStatusesByOrderHashes(status string, hashes ...string) ([]*types.Trade, error)
	Drop()
}

type TokenDao interface {
	Create(token *types.Token) error
	GetAll() ([]types.Token, error)
	GetByID(id bson.ObjectId) (*types.Token, error)
	GetByAsset(asset string) (*types.Token, error)
	GetBySymbol(symbol string) (*types.Token, error)
	GetByAssetOrSymbol(assetOrSymbol string) (*types.Token, error)
	GetQuoteTokens() ([]types.Token, error)
	GetBaseTokens() ([]types.Token, error)
	GetListedTokens() ([]types.Token, error)
	GetUnlistedTokens() ([]types.Token, error)
	GetListedBaseTokens() ([]types.Token, error)
	Drop() error
}

type Engine interface {
	HandleOrders(msg *rabbitmq.Message) error
	// RecoverOrders(matches types.Matches) error
	// CancelOrder(order *types.Order) (*types.EngineResponse, error)
	// DeleteOrder(o *types.Order) error
}

type InfoService interface {
	GetExchangeData() (*types.ExchangeData, error)
	GetExchangeStats() (*types.ExchangeStats, error)
	GetPairStats() (*types.PairStats, error)
}

type OHLCVService interface {
	Unsubscribe(c *ws.Client)
	UnsubscribeChannel(c *ws.Client, p *types.SubscriptionPayload)
	Subscribe(c *ws.Client, p *types.SubscriptionPayload)
	GetOHLCV(p []types.PairAssets, duration int64, unit string, timeInterval ...int64) ([]*types.Tick, error)
}

type OrderService interface {
	GetByID(id bson.ObjectId) (*types.Order, error)
	GetByHash(h string) (*types.Order, error)
	GetByHashes(hashes []string) ([]*types.Order, error)
	GetByUserAddress(a string, limit ...int) ([]*types.Order, error)
	GetCurrentByUserAddress(a string, limit ...int) ([]*types.Order, error)
	GetHistoryByUserAddress(a string, limit ...int) ([]*types.Order, error)
	NewOrder(o *types.Order) error
	CancelOrder(oc *types.OrderCancel) error
	HandleEngineResponse(res *types.EngineResponse) error
	GetSenderAddresses(oc *types.OrderCancel) (string, string, error)
	CheckIfBalancesAreSufficientAndCancel(address string, balances map[string]int64)
	CancelOrdersSignedByRevokedSigner(address string, signer string)
	CancelExpiredOrders()
	AdjustBalancesForUncommittedTrades(address string, balances map[string]int64) map[string]int64
	FixOrderStatus(o *types.Order)
}

type OrderBookService interface {
	GetOrderBook(bt, qt string) (map[string]interface{}, error)
	GetRawOrderBook(bt, qt string) (*types.RawOrderBook, error)
	SubscribeOrderBook(c *ws.Client, bt, qt string)
	UnsubscribeOrderBook(c *ws.Client)
	UnsubscribeOrderBookChannel(c *ws.Client, bt, qt string)
	SubscribeRawOrderBook(c *ws.Client, bt, qt string)
	UnsubscribeRawOrderBook(c *ws.Client)
	UnsubscribeRawOrderBookChannel(c *ws.Client, bt, qt string)
}

type PairService interface {
	Create(pair *types.Pair) error
	CreatePairs(token string) ([]*types.Pair, error)
	GetByID(id bson.ObjectId) (*types.Pair, error)
	GetByAsset(bt, qt string) (*types.Pair, error)
	GetTokenPairData(bt, qt string) ([]*types.Tick, error)
	GetAllExactTokenPairData() ([]*types.PairData, error)
	GetAllSimplifiedTokenPairData() ([]*types.SimplifiedPairAPIData, error)
	GetAllTokenPairData() ([]*types.PairAPIData, error)
	GetAll() ([]types.Pair, error)
	GetListedPairs() ([]types.Pair, error)
	GetUnlistedPairs() ([]types.Pair, error)
}

type TokenService interface {
	Create(token *types.Token) error
	GetByID(id bson.ObjectId) (*types.Token, error)
	GetByAsset(a string) (*types.Token, error)
	GetBySymbol(s string) (*types.Token, error)
	GetByAssetOrSymbol(a string) (*types.Token, error)
	CheckByAssetOrSymbol(a string) (*types.Token, error)
	GetAll() ([]types.Token, error)
	GetQuoteTokens() ([]types.Token, error)
	GetBaseTokens() ([]types.Token, error)
	GetListedTokens() ([]types.Token, error)
	GetUnlistedTokens() ([]types.Token, error)
	GetListedBaseTokens() ([]types.Token, error)
}

type TradeService interface {
	GetByPairName(p string) ([]*types.Trade, error)
	GetAllTradesByPairAssets(bt, qt string) ([]*types.Trade, error)
	GetSortedTrades(bt, qt string, n int) ([]*types.Trade, error)
	GetSortedTradesByUserAddress(a string, limit ...int) ([]*types.Trade, error)
	GetByUserAddress(a string) ([]*types.Trade, error)
	GetByHash(h string) (*types.Trade, error)
	GetByOrderHashes(h []string) ([]*types.Trade, error)
	GetByMakerOrderHash(h string) ([]*types.Trade, error)
	GetByTakerOrderHash(h string) ([]*types.Trade, error)
	GetByTriggerUnitHash(h string) ([]*types.Trade, error)
	GetByHashes(hashes []string) ([]*types.Trade, error)
	UpdateTradeTxHash(tr *types.Trade, txh string) error
	UpdateSuccessfulTrade(t *types.Trade) (*types.Trade, error)
	UpdateTradeStatus(t *types.Trade, status string) (*types.Trade, error)
	UpdatePendingTrade(t *types.Trade, txh string) (*types.Trade, error)
	Subscribe(c *ws.Client, bt, qt string)
	UnsubscribeChannel(c *ws.Client, bt, qt string)
	Unsubscribe(c *ws.Client)
}

type AccountService interface {
	GetAll() ([]types.Account, error)
	Create(account *types.Account) error
	GetByID(id bson.ObjectId) (*types.Account, error)
	GetByAddress(address string) (*types.Account, error)
	FindOrCreate(address string) (*types.Account, error)
	GetTokenBalance(owner string, token string) (*types.TokenBalance, error)
	GetTokenBalances(owner string) (map[string]*types.TokenBalance, error)
}

type ValidatorService interface {
	ValidateOperatorAddress(o *types.Order) error
	ValidateBalance(o *types.Order) error
	ValidateAvailableBalance(o *types.Order, uncommittedDeltas map[string]int64, balanceLockedInMemoryOrders int64) error
	//VerifySignature(o *types.Order) (string, error)
	//VerifyCancelSignature(oc *types.OrderCancel) (string, error)
}

type PriceService interface {
	GetDollarMarketPrices(baseCurrencies []string) (map[string]float64, error)
	GetMultipleMarketPrices(baseCurrencies []string, quoteCurrencies []string) (map[string]map[string]float64, error)
}

type ObyteProvider interface {
	BalanceOf(owner string, token string) (int64, error)
	GetBalances(owner string) map[string]int64
	GetOperatorAddress() string
	GetFees() (float64, float64)
	Decimals(token string) (uint8, error)
	Symbol(token string) (string, error)
	Asset(symbol string) (string, error)
	//VerifySignature(order *types.Order) (string, error)
	//VerifyCancelSignature(oc *types.OrderCancel) (string, error)
	AddOrder(signedOrder *interface{}) (string, error)
	CancelOrder(signedCancel *interface{}) error
	GetAuthorizedAddresses(address string) ([]string, error)
	ExecuteTrade(m *types.Matches) ([]string, error)
	ListenToEvents() (chan map[string]interface{}, error)
}
