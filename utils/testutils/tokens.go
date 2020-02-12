package testutils

import (
	"github.com/byteball/odex-backend/types"
)

func GetTestZRXToken() types.Token {
	return types.Token{
		Symbol: "ZRX",
		Asset:  "7l7GzugRUz9b/q7M+A1K9IIQ5yWnqlB6CyImXx73TQs=",
	}
}

func GetTestWETHToken() types.Token {
	return types.Token{
		Symbol: "WETH",
		Asset:  "PP7/+yQc6+XAZ1WBsUzGwcmfIlInIRLsLlfoWJc/3kY=",
	}
}
