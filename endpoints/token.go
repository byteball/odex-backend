package endpoints

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/services"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/httputils"
	"github.com/gorilla/mux"
)

type tokenEndpoint struct {
	tokenService interfaces.TokenService
}

// ServeTokenResource sets up the routing of token endpoints and the corresponding handlers.
func ServeTokenResource(
	r *mux.Router,
	tokenService interfaces.TokenService,
) {
	e := &tokenEndpoint{tokenService}
	r.UseEncodedPath()
	r.HandleFunc("/tokens/base", e.HandleGetBaseTokens).Methods("GET")
	r.HandleFunc("/tokens/quote", e.HandleGetQuoteTokens).Methods("GET")
	r.HandleFunc("/tokens/{assetOrSymbol}", e.HandleGetToken).Methods("GET")
	//	r.HandleFunc("/tokens/{asset}", e.HandleCreateToken).Methods("POST")
	r.HandleFunc("/tokens/check/{assetOrSymbol}", e.HandleCheckToken).Methods("GET")
	r.HandleFunc("/tokens", e.HandleGetTokens).Methods("GET")
	r.HandleFunc("/tokens", e.HandleCreateTokens).Methods("POST")
}

func (e *tokenEndpoint) HandleCreateTokens(w http.ResponseWriter, r *http.Request) {
	var t types.Token
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&t)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	defer r.Body.Close()

	err = e.tokenService.Create(&t)
	if err != nil {
		if err == services.ErrTokenExists {
			httputils.WriteError(w, http.StatusBadRequest, "")
			return
		} else {
			logger.Error(err)
			httputils.WriteError(w, http.StatusInternalServerError, "")
			return
		}
	}

	httputils.WriteJSON(w, http.StatusCreated, t)
}

func (e *tokenEndpoint) HandleGetTokens(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()

	var res []types.Token
	var err error

	switch v.Get("listed") {
	case "":
		res, err = e.tokenService.GetAll()
	case "true":
		res, err = e.tokenService.GetListedTokens()
	case "false":
		res, err = e.tokenService.GetUnlistedTokens()
	}

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

func (e *tokenEndpoint) HandleGetQuoteTokens(w http.ResponseWriter, r *http.Request) {
	res, err := e.tokenService.GetQuoteTokens()
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *tokenEndpoint) HandleGetBaseTokens(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()

	var res []types.Token
	var err error

	switch v.Get("listed") {
	case "":
		res, err = e.tokenService.GetBaseTokens()
	case "true":
		res, err = e.tokenService.GetListedBaseTokens()
	case "false":
		res, err = e.tokenService.GetUnlistedTokens()
	}

	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *tokenEndpoint) HandleGetToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	assetOrSymbol, err := url.PathUnescape(vars["assetOrSymbol"])
	logger.Info("HandleGetToken", assetOrSymbol)
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "PathUnescape: "+err.Error())
		return
	}
	/*if !isValidAsset(asset) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Asset "+asset)
		return
	}*/

	res, err := e.tokenService.GetByAssetOrSymbol(assetOrSymbol)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *tokenEndpoint) HandleCheckToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	assetOrSymbol, err := url.PathUnescape(vars["assetOrSymbol"])
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "PathUnescape: "+err.Error())
		return
	}
	logger.Info("HandleCheckToken", assetOrSymbol)
	/*if !isValidAsset(asset) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Asset")
		return
	}*/

	t, err := e.tokenService.CheckByAssetOrSymbol(assetOrSymbol)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, t)
}

/*
func (e *tokenEndpoint) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	asset := vars["asset"]
	if !isValidAsset(asset) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Asset")
	}

	t, err := e.tokenService.CheckByAsset(asset)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	defer r.Body.Close()

	err = e.tokenService.Create(t)
	if err != nil {
		if err == services.ErrTokenExists {
			httputils.WriteError(w, http.StatusBadRequest, "")
			return
		} else {
			logger.Error(err)
			httputils.WriteError(w, http.StatusInternalServerError, "")
			return
		}
	}

	httputils.WriteJSON(w, http.StatusCreated, t)
}
*/
