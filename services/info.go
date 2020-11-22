package services

import (
	"log"
	"time"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
	"github.com/globalsign/mgo/bson"
)

type InfoService struct {
	pairDao      interfaces.PairDao
	tokenDao     interfaces.TokenDao
	tradeDao     interfaces.TradeDao
	orderDao     interfaces.OrderDao
	priceService interfaces.PriceService
}

func NewInfoService(
	pairDao interfaces.PairDao,
	tokenDao interfaces.TokenDao,
	tradeDao interfaces.TradeDao,
	orderDao interfaces.OrderDao,
	priceService interfaces.PriceService,
) *InfoService {

	return &InfoService{
		pairDao,
		tokenDao,
		tradeDao,
		orderDao,
		priceService,
	}
}

func (s *InfoService) GetExchangeStats() (*types.ExchangeStats, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	tokens, err := s.tokenDao.GetBaseTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	quoteTokens, err := s.tokenDao.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	erroredTradeCount, err := s.tradeDao.GetErroredTradeCount(start, end)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tokenSymbols := []string{}
	for _, t := range tokens {
		tokenSymbols = append(tokenSymbols, t.Symbol)
	}

	for _, t := range quoteTokens {
		tokenSymbols = append(tokenSymbols, t.Symbol)
	}

	rates, err := s.priceService.GetDollarMarketPrices(tokenSymbols)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pairs, err := s.pairDao.GetDefaultPairs()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tradesQuery := []bson.M{
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

	tradeData, err := s.tradeDao.Aggregate(tradesQuery)
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

	var totalOrders int
	var totalTrades int
	var totalVolume float64
	var totalBuyOrderAmount float64
	var totalSellOrderAmount float64
	var totalSellOrders int
	var totalBuyOrders int
	pairTradeCounts := map[string]int{}
	tokenTradeCounts := map[string]int{}

	for _, p := range pairs {
		for _, t := range tradeData {
			if t.AssetCode() == p.AssetCode() {
				totalTrades += int(t.Count)

				if exchangeRate := rates[p.BaseTokenSymbol]; exchangeRate != 0 {
					totalVolume = totalVolume + t.ConvertedVolume(&p, rates[p.BaseTokenSymbol])
				}

				pairTradeCounts[p.Name()] = int(t.Count)
				tokenTradeCounts[p.BaseTokenSymbol] += int(t.Count)
			}
		}

		for _, o := range bidsData {
			if o.AssetCode() == p.AssetCode() {
				// change and replace by equivalent dollar volume instead of order count
				if exchangeRate := rates[p.BaseTokenSymbol]; exchangeRate != 0 {
					totalBuyOrderAmount = totalBuyOrderAmount + o.ConvertedVolume(&p, rates[p.BaseTokenSymbol])
				}
				totalBuyOrders += int(o.OrderCount)
			}
		}

		for _, o := range asksData {
			if o.AssetCode() == p.AssetCode() {
				// change and replace by equivalent dollar volume instead of order count
				if exchangeRate := rates[p.BaseTokenSymbol]; exchangeRate != 0 {
					totalSellOrderAmount = totalSellOrderAmount + o.ConvertedVolume(&p, rates[p.BaseTokenSymbol])
				}
				totalSellOrders += int(o.OrderCount)
			}
		}
	}

	mostTradedToken, _ := utils.MaxIntMap(tokenTradeCounts)
	mostTradedPair, _ := utils.MaxIntMap(pairTradeCounts)
	totalOrders = totalSellOrders + totalBuyOrders
	totalOrderAmount := totalBuyOrderAmount + totalSellOrderAmount

	tradeSuccessRatio := float64(1)
	if totalTrades > 0 {
		tradeSuccessRatio = float64(totalTrades-erroredTradeCount) / float64(totalTrades)
	}

	stats := &types.ExchangeStats{
		TotalOrders:          totalOrders,
		TotalTrades:          totalTrades,
		TotalBuyOrderAmount:  totalBuyOrderAmount,
		TotalSellOrderAmount: totalSellOrderAmount,
		TotalVolume:          totalVolume,
		TotalOrderAmount:     totalOrderAmount,
		TotalBuyOrders:       totalBuyOrders,
		TotalSellOrders:      totalSellOrders,
		MostTradedToken:      mostTradedToken,
		MostTradedPair:       mostTradedPair,
		TradeSuccessRatio:    tradeSuccessRatio,
	}

	log.Printf("%+v\n", stats)

	return stats, nil
}

func (s *InfoService) GetExchangeData() (*types.ExchangeData, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	tokens, err := s.tokenDao.GetBaseTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	quoteTokens, err := s.tokenDao.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	erroredTradeCount, err := s.tradeDao.GetErroredTradeCount(start, end)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tokenSymbols := []string{}
	for _, t := range tokens {
		tokenSymbols = append(tokenSymbols, t.Symbol)
	}

	for _, t := range quoteTokens {
		tokenSymbols = append(tokenSymbols, t.Symbol)
	}

	rates, err := s.priceService.GetDollarMarketPrices(tokenSymbols)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pairs, err := s.pairDao.GetDefaultPairs()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tradesQuery := []bson.M{
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

	tradeData, err := s.tradeDao.Aggregate(tradesQuery)
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

	var pairAPIData []*types.PairAPIData
	var totalOrders int
	var totalTrades int
	var totalVolume float64
	var totalBuyOrderAmount float64
	var totalSellOrderAmount float64
	var totalSellOrders int
	var totalBuyOrders int
	pairTradeCounts := map[string]int{}
	tokenTradeCounts := map[string]int{}

	// //total orderbook volume per quote token
	// totalOrderBookVolume := map[string]int{}

	for _, p := range pairs {
		pairData := &types.PairData{
			Pair:               types.PairID{PairName: p.Name(), BaseToken: p.BaseAsset, QuoteToken: p.QuoteAsset},
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
				pairData.Volume = t.Volume
				pairData.QuoteVolume = t.QuoteVolume
				pairData.Close = t.Close
				pairData.Count = t.Count
				pairData.AverageTradeAmount = t.Volume / t.Count

				totalTrades += int(t.Count)

				if exchangeRate := rates[p.BaseTokenSymbol]; exchangeRate != 0 {
					totalVolume = totalVolume + t.ConvertedVolume(&p, rates[p.BaseTokenSymbol])
				}

				pairTradeCounts[p.Name()] = int(t.Count)
				tokenTradeCounts[p.BaseTokenSymbol] += int(t.Count)
			}
		}

		for _, o := range bidsData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume = o.OrderVolume
				pairData.OrderCount = o.OrderCount
				pairData.BidPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)

				if exchangeRate := rates[p.BaseTokenSymbol]; exchangeRate != 0 {
					totalBuyOrderAmount = totalBuyOrderAmount + o.ConvertedVolume(&p, rates[p.BaseTokenSymbol])
				}

				totalBuyOrders += int(o.OrderCount)
			}
		}

		for _, o := range asksData {
			if o.AssetCode() == p.AssetCode() {
				pairData.OrderVolume += o.OrderVolume
				pairData.OrderCount += o.OrderCount
				pairData.AskPrice = o.BestPrice
				pairData.AverageOrderAmount = (pairData.OrderVolume / pairData.OrderCount)

				if exchangeRate := rates[p.BaseTokenSymbol]; exchangeRate != 0 {
					totalSellOrderAmount = totalSellOrderAmount + o.ConvertedVolume(&p, rates[p.BaseTokenSymbol])
				}

				totalSellOrders += int(o.OrderCount)

				//TODO change price into orderbook price
				if pairData.BidPrice != 0 && pairData.AskPrice != 0 {
					pairData.Price = (pairData.BidPrice + pairData.AskPrice) / 2
				} else {
					pairData.Price = 0
				}
			}
		}

		pairAPIData = append(pairAPIData, pairData.ToAPIData(&p))
	}

	mostTradedToken, _ := utils.MaxIntMap(tokenTradeCounts)
	mostTradedPair, _ := utils.MaxIntMap(pairTradeCounts)
	totalOrders = totalSellOrders + totalBuyOrders
	totalOrderAmount := totalBuyOrderAmount + totalSellOrderAmount
	tradeSuccessRatio := float64(totalTrades-erroredTradeCount) / float64(totalTrades)

	exchangeData := &types.ExchangeData{
		PairData:             pairAPIData,
		TotalOrders:          totalOrders,
		TotalTrades:          totalTrades,
		TotalBuyOrderAmount:  totalBuyOrderAmount,
		TotalSellOrderAmount: totalSellOrderAmount,
		TotalVolume:          totalVolume,
		TotalOrderAmount:     totalOrderAmount,
		TotalBuyOrders:       totalSellOrders,
		TotalSellOrders:      totalBuyOrders,
		MostTradedToken:      mostTradedToken,
		MostTradedPair:       mostTradedPair,
		TradeSuccessRatio:    tradeSuccessRatio,
	}

	return exchangeData, nil
}

func (s *InfoService) GetPairStats() (*types.PairStats, error) {
	now := time.Now()
	end := time.Unix(now.Unix(), 0)
	start := time.Unix(now.AddDate(0, 0, -7).Unix(), 0)
	//one, _ := bson.ParseDecimal128("1")

	tokens, err := s.tokenDao.GetBaseTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	quoteTokens, err := s.tokenDao.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tokenSymbols := []string{}
	for _, t := range tokens {
		tokenSymbols = append(tokenSymbols, t.Symbol)
	}

	for _, t := range quoteTokens {
		tokenSymbols = append(tokenSymbols, t.Symbol)
	}

	pairs, err := s.pairDao.GetDefaultPairs()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tradesQuery := []bson.M{
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

	tradeData, err := s.tradeDao.Aggregate(tradesQuery)
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

	var pairStatistics types.PairStats

	for _, p := range pairs {
		pairData := &types.PairData{
			Pair:               types.PairID{PairName: p.Name(), BaseToken: p.BaseAsset, QuoteToken: p.QuoteAsset},
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

				//TODO change price into orderbook price
				if pairData.BidPrice != 0 && pairData.AskPrice != 0 {
					pairData.Price = (pairData.BidPrice + pairData.AskPrice) / 2
				} else {
					pairData.Price = 0
				}
			}
		}

		pairStatistics = append(pairStatistics, pairData.ToAPIData(&p))
	}

	return &pairStatistics, nil
}
