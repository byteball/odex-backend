package endpoints

import (
	"net/http"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/httputils"
	"github.com/gorilla/mux"
)

type infoEndpoint struct {
	tokenService  interfaces.TokenService
	infoService   interfaces.InfoService
	obyteProvider interfaces.ObyteProvider
}

func ServeInfoResource(
	r *mux.Router,
	tokenService interfaces.TokenService,
	infoService interfaces.InfoService,
	obyteProvider interfaces.ObyteProvider,
) {

	e := &infoEndpoint{tokenService, infoService, obyteProvider}
	r.HandleFunc("/info", e.handleGetInfo)
	r.HandleFunc("/info/exchange", e.handleGetExchangeInfo)
	r.HandleFunc("/info/operators", e.handleGetOperatorsInfo)
	r.HandleFunc("/info/fees", e.handleGetFeeInfo)
	r.HandleFunc("/stats/trading", e.handleGetTradingStats)
	// r.HandleFunc("/stats/all", e.handleGetStats)
	// r.HandleFunc("/stats/pairs", e.handleGetPairStats)
}

func (e *infoEndpoint) handleGetInfo(w http.ResponseWriter, r *http.Request) {
	operator_address := e.obyteProvider.GetOperatorAddress()
	matcherFee, affiliateFee := e.obyteProvider.GetFees()

	operators := [1]string{operator_address}

	quotes, err := e.tokenService.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		return
	}

	fees := map[string]map[string]float64{}
	for _, q := range quotes {
		fees[q.Symbol] = map[string]float64{
			"matcherFee":   matcherFee,
			"affiliateFee": affiliateFee,
		}
	}

	res := map[string]interface{}{
		"operatorAddress": operator_address,
		"fees":            fees,
		"operators":       operators,
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *infoEndpoint) handleGetExchangeInfo(w http.ResponseWriter, r *http.Request) {
	operator_address := e.obyteProvider.GetOperatorAddress()

	res := map[string]string{"operatorAddress": operator_address}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *infoEndpoint) handleGetOperatorsInfo(w http.ResponseWriter, r *http.Request) {
	operator_address := e.obyteProvider.GetOperatorAddress()

	addresses := []string{operator_address}
	res := map[string][]string{"operators": addresses}
	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *infoEndpoint) handleGetFeeInfo(w http.ResponseWriter, r *http.Request) {
	matcherFee, affiliateFee := e.obyteProvider.GetFees()
	quotes, err := e.tokenService.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		return
	}

	fees := map[string]map[string]float64{}
	for _, q := range quotes {
		fees[q.Symbol] = map[string]float64{
			"matcherFee":   matcherFee,
			"affiliateFee": affiliateFee,
		}
	}

	httputils.WriteJSON(w, http.StatusOK, fees)
}

func (e *infoEndpoint) handleGetTradingStats(w http.ResponseWriter, r *http.Request) {
	res, err := e.infoService.GetExchangeStats()
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if res == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Pair{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *infoEndpoint) handleGetPairStats(w http.ResponseWriter, r *http.Request) {
	res, err := e.infoService.GetPairStats()
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if res == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Pair{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *infoEndpoint) handleGetStats(w http.ResponseWriter, r *http.Request) {
	res, err := e.infoService.GetExchangeData()
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if res == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Pair{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}
