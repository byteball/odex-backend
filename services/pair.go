package services

import (
	"time"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/globalsign/mgo/bson"
)

// PairService struct with daos required, responsible for communicating with daos.
// PairService functions are responsible for interacting with daos and implements business logics.
type PairService struct {
	pairDao  interfaces.PairDao
	tokenDao interfaces.TokenDao
	tradeDao interfaces.TradeDao
	orderDao interfaces.OrderDao
	provider interfaces.ObyteProvider
}

// NewPairService returns a new instance of balance service
func NewPairService(
	pairDao interfaces.PairDao,
	tokenDao interfaces.TokenDao,
	tradeDao interfaces.TradeDao,
	orderDao interfaces.OrderDao,
	provider interfaces.ObyteProvider,
) *PairService {

	return &PairService{pairDao, tokenDao, tradeDao, orderDao, provider}
}

func (s *PairService) CreatePairs(asset string) ([]*types.Pair, error) {
	quotes, err := s.tokenDao.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	base, err := s.tokenDao.GetByAsset(asset)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if base == nil {
		symbol, err := s.provider.Symbol(asset)
		if err != nil {
			logger.Error(err)
			return nil, ErrNoAsset
		}

		decimals, err := s.provider.Decimals(asset)
		if err != nil {
			logger.Error(err)
			return nil, ErrNoAsset
		}

		base = &types.Token{
			Symbol:   symbol,
			Asset:    asset,
			Decimals: int(decimals),
			Active:   true,
			Listed:   true,
			Quote:    false,
		}

		err = s.tokenDao.Create(base)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
	}

	pairs := []*types.Pair{}
	for _, q := range quotes {
		p, err := s.pairDao.GetByAsset(asset, q.Asset)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if p == nil {
			p := types.Pair{
				QuoteTokenSymbol:   q.Symbol,
				QuoteAsset:         q.Asset,
				QuoteTokenDecimals: q.Decimals,
				BaseTokenSymbol:    base.Symbol,
				BaseAsset:          base.Asset,
				BaseTokenDecimals:  base.Decimals,
				Active:             true,
				Listed:             true,
			}

			err := s.pairDao.Create(&p)
			if err != nil {
				logger.Error(err)
				return nil, err
			}

			pairs = append(pairs, &p)
		}
	}

	return pairs, nil
}

// Create function is responsible for inserting new pair in DB.
// It checks for existence of tokens in DB first
func (s *PairService) Create(pair *types.Pair) error {
	p, err := s.pairDao.GetByAsset(pair.BaseAsset, pair.QuoteAsset)
	if err != nil {
		logger.Error(err)
		return err
	}

	if p != nil {
		return ErrPairExists
	}

	quote, err := s.tokenDao.GetByAsset(pair.QuoteAsset)
	if err != nil {
		logger.Error(err)
		return err
	}

	if quote == nil {
		return ErrQuoteTokenNotFound
	}

	if !quote.Quote {
		return ErrQuoteTokenInvalid
	}

	base, err := s.tokenDao.GetByAsset(pair.BaseAsset)
	if err != nil {
		logger.Error(err)
		return err
	}

	if base == nil {
		symbol, err := s.provider.Symbol(pair.BaseAsset)
		if err != nil {
			logger.Error(err)
			return ErrNoAsset
		}

		decimals, err := s.provider.Decimals(pair.BaseAsset)
		if err != nil {
			logger.Error(err)
			return ErrNoAsset
		}

		token := types.Token{
			Symbol:   symbol,
			Asset:    pair.BaseAsset,
			Decimals: int(decimals),
			Rank:     0,
			Active:   true,
			Listed:   false,
			Quote:    false,
		}

		err = s.tokenDao.Create(&token)
		if err != nil {
			logger.Error(err)
			return err
		}

		pair.QuoteTokenSymbol = quote.Symbol
		pair.QuoteAsset = quote.Asset
		pair.QuoteTokenDecimals = quote.Decimals
		pair.BaseTokenSymbol = token.Symbol
		pair.BaseAsset = token.Asset
		pair.BaseTokenDecimals = token.Decimals
		pair.Active = true
		pair.Listed = false

		err = s.pairDao.Create(pair)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

// GetByID fetches details of a pair using its mongo ID
func (s *PairService) GetByID(id bson.ObjectId) (*types.Pair, error) {
	return s.pairDao.GetByID(id)
}

// GetByAsset fetches details of a pair using asset IDs of
// its constituting tokens
func (s *PairService) GetByAsset(bt, qt string) (*types.Pair, error) {
	return s.pairDao.GetByAsset(bt, qt)
}

// GetAll is reponsible for fetching all the pairs in the DB
func (s *PairService) GetAll() ([]types.Pair, error) {
	return s.pairDao.GetAll()
}

func (s *PairService) GetListedPairs() ([]types.Pair, error) {
	return s.pairDao.GetListedPairs()
}

func (s *PairService) GetUnlistedPairs() ([]types.Pair, error) {
	return s.pairDao.GetUnlistedPairs()
}

func (s *PairService) GetTokenPairData(bt, qt string) ([]*types.Tick, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	q := []bson.M{
		bson.M{
			"$match": bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lt":  end,
				},
				"status":     bson.M{"$in": []string{"SUCCESS", "COMMITTED"}},
				"baseToken":  bt,
				"quoteToken": qt,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"baseToken":  "$baseToken",
					"pairName":   "$pairName",
					"quoteToken": "$quoteToken",
				},
				"count":       bson.M{"$sum": 1},
				"open":        bson.M{"$first": "$price"},
				"high":        bson.M{"$max": "$price"},
				"low":         bson.M{"$min": "$price"},
				"close":       bson.M{"$last": "$price"},
				"volume":      bson.M{"$sum": "$amount"},
				"quoteVolume": bson.M{"$sum": "$quoteAmount"},
			},
		},
	}

	res, err := s.tradeDao.Aggregate(q)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *PairService) GetAllExactTokenPairData() ([]*types.PairData, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	pairs, err := s.pairDao.GetActivePairs()
	if err != nil {
		return nil, err
	}

	tradeDataQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lt":  end,
				},
				"status": bson.M{"$in": []string{"SUCCESS", "COMMITTED"}},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"count":       bson.M{"$sum": 1},
				"open":        bson.M{"$first": "$price"},
				"high":        bson.M{"$max": "$price"},
				"low":         bson.M{"$min": "$price"},
				"close":       bson.M{"$last": "$price"},
				"volume":      bson.M{"$sum": "$amount"},
				"quoteVolume": bson.M{"$sum": "$quoteAmount"},
			},
		},
	}

	bidsQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"side":   "BUY",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": 1},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
				"bestPrice": bson.M{"$max": "$price"},
			},
		},
	}

	asksQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"side":   "SELL",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": 1},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
				"bestPrice": bson.M{"$min": "$price"},
			},
		},
	}

	tradeData, err := s.tradeDao.Aggregate(tradeDataQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	bidsData, err := s.orderDao.Aggregate(bidsQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	asksData, err := s.orderDao.Aggregate(asksQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pairsData := []*types.PairData{}
	for _, p := range pairs {
		pairData := &types.PairData{
			Pair:               types.PairID{PairName: p.Name(), BaseToken: p.BaseAsset, QuoteToken: p.QuoteAsset},
			Open:               0,
			High:               0,
			Low:                0,
			Volume:             0,
			QuoteVolume:        0,
			Close:              0,
			Count:              0,
			OrderVolume:        0,
			OrderCount:         0,
			BidPrice:           0,
			AskPrice:           0,
			Price:              0,
			AverageOrderAmount: 0,
			AverageTradeAmount: 0,
		}

		for _, t := range tradeData {
			if t.AssetCode() == p.AssetCode() {
				pairData.Open = t.Open
				pairData.High = t.High
				pairData.Low = t.Low
				pairData.Volume = t.Volume
				pairData.QuoteVolume = t.QuoteVolume
				pairData.Close = t.Close
				pairData.Count = t.Count
				pairData.AverageTradeAmount = (t.Volume / t.Count)

			}
		}

		for _, o := range bidsData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume = o.OrderVolume
				pairData.OrderCount = o.OrderCount
				pairData.BidPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)
			}
		}

		for _, o := range asksData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume += o.OrderVolume
				pairData.OrderCount += o.OrderCount
				pairData.AskPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)

				if pairData.BidPrice != 0 && pairData.AskPrice != 0 {
					pairData.Price = (pairData.BidPrice + pairData.AskPrice) / 2
				} else {
					pairData.Price = 0
				}
			}
		}

		pairsData = append(pairsData, pairData)
	}

	return pairsData, nil
}

// Return a simplified version of the pair data
func (s *PairService) GetAllTokenPairData() ([]*types.PairAPIData, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	pairs, err := s.pairDao.GetActivePairs()
	if err != nil {
		return nil, err
	}

	tradeDataQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lt":  end,
				},
				"status": bson.M{"$in": []string{"SUCCESS", "COMMITTED"}},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"count":       bson.M{"$sum": 1},
				"open":        bson.M{"$first": "$price"},
				"high":        bson.M{"$max": "$price"},
				"low":         bson.M{"$min": "$price"},
				"close":       bson.M{"$last": "$price"},
				"volume":      bson.M{"$sum": "$amount"},
				"quoteVolume": bson.M{"$sum": "$quoteAmount"},
			},
		},
	}

	bidsQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"side":   "BUY",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": 1},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
				"bestPrice": bson.M{"$max": "$price"},
			},
		},
	}

	asksQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"side":   "SELL",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": 1},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
				"bestPrice": bson.M{"$min": "$price"},
			},
		},
	}

	tradeData, err := s.tradeDao.Aggregate(tradeDataQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	bidsData, err := s.orderDao.Aggregate(bidsQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	asksData, err := s.orderDao.Aggregate(asksQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pairsData := []*types.PairAPIData{}
	for _, p := range pairs {
		pairData := &types.PairData{
			Pair:               types.PairID{PairName: p.Name(), BaseToken: p.BaseAsset, QuoteToken: p.QuoteAsset},
			Open:               0,
			High:               0,
			Low:                0,
			Volume:             0,
			QuoteVolume:        0,
			Close:              0,
			Count:              0,
			OrderVolume:        0,
			OrderCount:         0,
			BidPrice:           0,
			AskPrice:           0,
			Price:              0,
			AverageOrderAmount: 0,
			AverageTradeAmount: 0,
		}

		for _, t := range tradeData {
			if t.AssetCode() == p.AssetCode() {
				pairData.Open = t.Open
				pairData.High = t.High
				pairData.Low = t.Low
				pairData.Volume = t.Volume
				pairData.QuoteVolume = t.QuoteVolume
				pairData.Close = t.Close
				pairData.Count = t.Count
				pairData.AverageTradeAmount = (t.Volume / t.Count)

			}
		}

		for _, o := range bidsData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume = o.OrderVolume
				pairData.OrderCount = o.OrderCount
				pairData.BidPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)
			}
		}

		for _, o := range asksData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume += o.OrderVolume
				pairData.OrderCount += o.OrderCount
				pairData.AskPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)

				if pairData.BidPrice != 0 && pairData.AskPrice != 0 {
					pairData.Price = (pairData.BidPrice + pairData.AskPrice) / 2
				} else {
					pairData.Price = 0
				}
			}
		}

		pairsData = append(pairsData, pairData.ToAPIData(&p))
	}

	return pairsData, nil
}

func (s *PairService) GetAllSimplifiedTokenPairData() ([]*types.SimplifiedPairAPIData, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	pairs, err := s.pairDao.GetActivePairs()
	if err != nil {
		return nil, err
	}

	tradeDataQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lt":  end,
				},
				"status": bson.M{"$in": []string{"SUCCESS", "COMMITTED"}},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"count":       bson.M{"$sum": 1},
				"open":        bson.M{"$first": "$price"},
				"high":        bson.M{"$max": "$price"},
				"low":         bson.M{"$min": "$price"},
				"close":       bson.M{"$last": "$price"},
				"volume":      bson.M{"$sum": "$amount"},
				"quoteVolume": bson.M{"$sum": "$quoteAmount"},
			},
		},
	}

	bidsQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"side":   "BUY",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": 1},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
				"bestPrice": bson.M{"$max": "$price"},
			},
		},
	}

	asksQuery := []bson.M{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$in": []string{"OPEN", "PARTIAL_FILLED"}},
				"side":   "SELL",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"pairName":   "$pairName",
					"baseToken":  "$baseToken",
					"quoteToken": "$quoteToken",
				},
				"orderCount": bson.M{"$sum": 1},
				"orderVolume": bson.M{
					"$sum": bson.M{
						"$subtract": []string{"$amount", "$filledAmount"},
					},
				},
				"bestPrice": bson.M{"$min": "$price"},
			},
		},
	}

	tradeData, err := s.tradeDao.Aggregate(tradeDataQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	bidsData, err := s.orderDao.Aggregate(bidsQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	asksData, err := s.orderDao.Aggregate(asksQuery)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pairsData := []*types.SimplifiedPairAPIData{}
	for _, p := range pairs {
		pairData := &types.PairData{
			Pair:               types.PairID{PairName: p.Name(), BaseToken: p.BaseAsset, QuoteToken: p.QuoteAsset},
			Open:               0,
			High:               0,
			Low:                0,
			Volume:             0,
			QuoteVolume:        0,
			Close:              0,
			Count:              0,
			OrderVolume:        0,
			OrderCount:         0,
			BidPrice:           0,
			AskPrice:           0,
			Price:              0,
			AverageOrderAmount: 0,
			AverageTradeAmount: 0,
		}

		for _, t := range tradeData {
			if t.AssetCode() == p.AssetCode() {
				pairData.Open = t.Open
				pairData.High = t.High
				pairData.Low = t.Low
				pairData.Volume = t.Volume
				pairData.QuoteVolume = t.QuoteVolume
				pairData.Close = t.Close
				pairData.Count = t.Count
				pairData.AverageTradeAmount = (t.Volume / t.Count)

			}
		}

		for _, o := range bidsData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume = o.OrderVolume
				pairData.OrderCount = o.OrderCount
				pairData.BidPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)
			}
		}

		for _, o := range asksData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume += o.OrderVolume
				pairData.OrderCount += o.OrderCount
				pairData.AskPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)

				if pairData.BidPrice != 0 && pairData.AskPrice != 0 {
					pairData.Price = (pairData.BidPrice + pairData.AskPrice) / 2
				} else {
					pairData.Price = 0
				}
			}
		}

		pairsData = append(pairsData, pairData.ToSimplifiedAPIData(&p))
	}

	return pairsData, nil
}
