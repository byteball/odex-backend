package endpoints

import "github.com/byteball/odex-backend/utils"

var logger = utils.Logger

// maxListLimit caps rows returned by list endpoints; a missing, zero or oversized limit means "max".
const maxListLimit = 1000

func clampLimit(limit int) int {
	if limit <= 0 || limit > maxListLimit {
		return maxListLimit
	}

	return limit
}
