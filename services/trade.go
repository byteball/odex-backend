package services

import (
	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
	"github.com/byteball/odex-backend/ws"
)

// TradeService struct with daos required, responsible for communicating with daos.
// TradeService functions are responsible for interacting with daos and implements business logics.
type TradeService struct {
	tradeDao interfaces.TradeDao
}

// NewTradeService returns a new instance of TradeService
func NewTradeService(TradeDao interfaces.TradeDao) *TradeService {
	return &TradeService{TradeDao}
}

// Subscribe
func (s *TradeService) Subscribe(c *ws.Client, bt, qt string) {
	socket := ws.GetTradeSocket()

	numTrades := 40
	trades, err := s.GetSortedTrades(bt, qt, numTrades)
	if err != nil {
		logger.Error(err)
		socket.SendErrorMessage(c, err.Error())
		return
	}

	id := utils.GetTradeChannelID(bt, qt)
	err = socket.Subscribe(id, c)
	if err != nil {
		logger.Error(err)
		socket.SendErrorMessage(c, err.Error())
		return
	}

	ws.RegisterConnectionUnsubscribeHandler(c, socket.UnsubscribeChannelHandler(id))
	socket.SendInitMessage(c, trades)
}

// Unsubscribe
func (s *TradeService) UnsubscribeChannel(c *ws.Client, bt, qt string) {
	socket := ws.GetTradeSocket()

	id := utils.GetTradeChannelID(bt, qt)
	socket.UnsubscribeChannel(id, c)
}

// Unsubscribe
func (s *TradeService) Unsubscribe(c *ws.Client) {
	socket := ws.GetTradeSocket()
	socket.Unsubscribe(c)
}

// GetByPairName fetches all the trades corresponding to a pair using pair's name
func (s *TradeService) GetByPairName(p string) ([]*types.Trade, error) {
	return s.tradeDao.GetByPairName(p)
}

// GetByPairAssets fetches all the trades corresponding to a pair using pair's asset IDs
func (s *TradeService) GetAllTradesByPairAssets(bt, qt string) ([]*types.Trade, error) {
	return s.tradeDao.GetAllTradesByPairAssets(bt, qt)
}

func (s *TradeService) GetSortedTradesByUserAddress(a string, limit ...int) ([]*types.Trade, error) {
	return s.tradeDao.GetSortedTradesByUserAddress(a, limit...)
}

func (s *TradeService) GetSortedTrades(bt, qt string, n int) ([]*types.Trade, error) {
	return s.tradeDao.GetSortedTrades(bt, qt, n)
}

// GetByUserAddress fetches all the trades corresponding to a user address
func (s *TradeService) GetByUserAddress(a string) ([]*types.Trade, error) {
	return s.tradeDao.GetByUserAddress(a)
}

// GetByHash fetches all trades corresponding to a trade hash
func (s *TradeService) GetByHash(h string) (*types.Trade, error) {
	return s.tradeDao.GetByHash(h)
}

func (s *TradeService) GetByMakerOrderHash(h string) ([]*types.Trade, error) {
	return s.tradeDao.GetByMakerOrderHash(h)
}

func (s *TradeService) GetByTakerOrderHash(h string) ([]*types.Trade, error) {
	return s.tradeDao.GetByTakerOrderHash(h)
}

func (s *TradeService) GetByTriggerUnitHash(h string) ([]*types.Trade, error) {
	return s.tradeDao.GetByTriggerUnitHash(h)
}

func (s *TradeService) GetByHashes(hashes []string) ([]*types.Trade, error) {
	return s.tradeDao.GetByHashes(hashes)
}

func (s *TradeService) GetByOrderHashes(hashes []string) ([]*types.Trade, error) {
	return s.tradeDao.GetByOrderHashes(hashes)
}

func (s *TradeService) UpdatePendingTrade(t *types.Trade, txh string) (*types.Trade, error) {
	t.Status = "PENDING"
	t.TxHash = txh

	updated, err := s.tradeDao.FindAndModify(t.Hash, t)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return updated, nil
}

func (s *TradeService) UpdateSuccessfulTrade(t *types.Trade) (*types.Trade, error) {
	t.Status = "SUCCESS"

	updated, err := s.tradeDao.FindAndModify(t.Hash, t)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return updated, nil
}

func (s *TradeService) UpdateTradeStatus(t *types.Trade, status string) (*types.Trade, error) {
	t.Status = status

	updated, err := s.tradeDao.FindAndModify(t.Hash, t)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return updated, nil
}

func (s *TradeService) UpdateTradeTxHash(tr *types.Trade, txh string) error {
	tr.TxHash = txh

	err := s.tradeDao.UpdateByHash(tr.Hash, tr)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
