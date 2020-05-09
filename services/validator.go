package services

import (
	"fmt"
	"strings"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
)

type ValidatorService struct {
	obyteProvider interfaces.ObyteProvider
	accountDao    interfaces.AccountDao
	orderDao      interfaces.OrderDao
	pairDao       interfaces.PairDao
}

func NewValidatorService(
	obyteProvider interfaces.ObyteProvider,
	accountDao interfaces.AccountDao,
	orderDao interfaces.OrderDao,
	pairDao interfaces.PairDao,
) *ValidatorService {

	return &ValidatorService{
		obyteProvider,
		accountDao,
		orderDao,
		pairDao,
	}
}

func (s *ValidatorService) ValidateOperatorAddress(o *types.Order) error {
	// operator_address := s.obyteProvider.GetOperatorAddress()
	// if o.MatcherAddress != operator_address && o.AffiliateAddress != operator_address {
	// 	return errors.New("Order 'matcherAddress' or 'affiliateAddress' parameter is incorrect")
	// }
	return nil
}

func (s *ValidatorService) ValidateAvailableBalance(o *types.Order, uncommittedDeltas map[string]int64, balanceLockedInMemoryOrders int64) error {

	pair, err := s.pairDao.GetByAsset(o.BaseToken, o.QuoteToken)
	if err != nil {
		logger.Error(err)
		return err
	}

	totalRequiredAmount := o.TotalRequiredSellAmount(pair)

	var sellTokenBalance int64

	// we implement retries in the case the provider connection fell asleep
	err = utils.Retry(3, func() error {
		sellTokenBalance, err = s.obyteProvider.BalanceOf(o.UserAddress, o.SellToken())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(err)
		return err
	}

	sellTokenBalance += uncommittedDeltas[o.SellTokenSymbol()]

	//Sell Token Balance
	if sellTokenBalance < totalRequiredAmount {
		return fmt.Errorf("Insufficient %v Balance: have %v, need %v for order %v at %v", o.SellTokenSymbol(), sellTokenBalance, totalRequiredAmount, o.Hash, o.Price)
	}

	sellTokenLockedBalance, orders, err := s.orderDao.GetUserLockedBalance(o.UserAddress, o.SellToken())
	if err != nil {
		logger.Error(err)
		return err
	}

	availableSellTokenBalance := sellTokenBalance - sellTokenLockedBalance - balanceLockedInMemoryOrders

	if availableSellTokenBalance < totalRequiredAmount {
		var orderLines []string
		for _, o := range orders {
			orderLines = append(orderLines, fmt.Sprintf("%v for %v at %v", o.Hash, o.RemainingSellAmount, o.Price))
		}
		return fmt.Errorf("Insufficient %v available: have %v, need %v for order %v at %v, total balance: %v, open orders:\n%v", o.SellTokenSymbol(), availableSellTokenBalance, totalRequiredAmount, o.Hash, o.Price, sellTokenBalance, strings.Join(orderLines, "\n"))
	}

	return nil
}

func (s *ValidatorService) ValidateBalance(o *types.Order) error {

	pair, err := s.pairDao.GetByAsset(o.BaseToken, o.QuoteToken)
	if err != nil {
		logger.Error(err)
		return err
	}

	totalRequiredAmount := o.TotalRequiredSellAmount(pair)

	var sellTokenBalance int64

	// we implement retries in the case the provider connection fell asleep
	err = utils.Retry(3, func() error {
		sellTokenBalance, err = s.obyteProvider.BalanceOf(o.UserAddress, o.SellToken())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(err)
		return err
	}

	//Sell Token Balance
	if sellTokenBalance < totalRequiredAmount {
		return fmt.Errorf("Insufficient %v Balance: have %v, need %v for order %v at %v", o.SellTokenSymbol(), sellTokenBalance, totalRequiredAmount, o.Hash, o.Price)
	}

	return nil
}

/*func (s *ValidatorService) VerifySignature(o *types.Order) (string, error) {
	id, err := s.obyteProvider.VerifySignature(o)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *ValidatorService) VerifyCancelSignature(oc *types.OrderCancel) (string, error) {
	addr, err := s.obyteProvider.VerifyCancelSignature(oc)
	if err != nil {
		return "", err
	}
	return addr, nil
}*/
