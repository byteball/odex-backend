package services

import (
	"github.com/byteball/odex-backend/interfaces"
	"github.com/globalsign/mgo/bson"

	"github.com/byteball/odex-backend/types"
)

// TokenService struct with daos required, responsible for communicating with daos.
// TokenService functions are responsible for interacting with daos and implements business logics.
type TokenService struct {
	tokenDao interfaces.TokenDao
	provider interfaces.ObyteProvider
}

// NewTokenService returns a new instance of TokenService
func NewTokenService(tokenDao interfaces.TokenDao, provider interfaces.ObyteProvider) *TokenService {
	return &TokenService{tokenDao, provider}
}

// Create inserts a new token into the database
func (s *TokenService) Create(token *types.Token) error {
	t, err := s.tokenDao.GetByAsset(token.Asset)
	if err != nil {
		logger.Error(err)
		return err
	}

	if t != nil {
		return ErrTokenExists
	}

	err = s.tokenDao.Create(token)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// GetByID fetches the detailed document of a token using its mongo ID
func (s *TokenService) GetByID(id bson.ObjectId) (*types.Token, error) {
	return s.tokenDao.GetByID(id)
}

// GetByAsset fetches the detailed document of a token using its asset ID
func (s *TokenService) GetByAsset(asset string) (*types.Token, error) {
	return s.tokenDao.GetByAsset(asset)
}

func (s *TokenService) CheckByAsset(asset string) (*types.Token, error) {
	t, err := s.tokenDao.GetByAsset(asset)
	if err != nil {
		return nil, err
	}
	if t != nil {
		return t, nil
	}
	symbol, err := s.provider.Symbol(asset)
	if err != nil {
		return nil, err
	}
	decimals, err := s.provider.Decimals(asset)
	if err != nil {
		return nil, err
	}
	t = &types.Token{
		Symbol:   symbol,
		Asset:  asset,
		Decimals: int(decimals),
		Active:   true,
		Listed:   false,
		Quote:    false,
		Rank:     0,
	}
	/*err = s.Create(t)
	if err != nil {
		return nil, err
	}*/
	return t, nil
}

// GetAll fetches all the tokens from db
func (s *TokenService) GetAll() ([]types.Token, error) {
	return s.tokenDao.GetAll()
}

func (s *TokenService) GetListedTokens() ([]types.Token, error) {
	return s.tokenDao.GetListedTokens()
}

func (s *TokenService) GetUnlistedTokens() ([]types.Token, error) {
	return s.tokenDao.GetUnlistedTokens()
}

// GetQuote fetches all the quote tokens from db
func (s *TokenService) GetQuoteTokens() ([]types.Token, error) {
	return s.tokenDao.GetQuoteTokens()
}

// GetBase fetches all the quote tokens from db
func (s *TokenService) GetBaseTokens() ([]types.Token, error) {
	return s.tokenDao.GetBaseTokens()
}

func (s *TokenService) GetListedBaseTokens() ([]types.Token, error) {
	return s.tokenDao.GetListedBaseTokens()
}
