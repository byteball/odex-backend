package testutils

import (
	"github.com/byteball/odex-backend/types"
)

func GetZRXWETHTestPair() *types.Pair {
	return &types.Pair{
		BaseTokenSymbol:    "ZRX",
		BaseAsset:   "0x2034842261b82651885751fc293bba7ba5398156",
		BaseTokenDecimals:  18,
		QuoteTokenSymbol:   "WETH",
		QuoteAsset:  "0x276e16ada4b107332afd776691a7fbbaede168ef",
		QuoteTokenDecimals: 18,
	}
}
