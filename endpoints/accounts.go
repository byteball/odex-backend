package endpoints

import (
	"net/http"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/httputils"
	"github.com/gorilla/mux"
)

type accountEndpoint struct {
	accountService interfaces.AccountService
	orderService   interfaces.OrderService
	obyteProvider  interfaces.ObyteProvider
}

func ServeAccountResource(
	r *mux.Router,
	accountService interfaces.AccountService,
	orderService interfaces.OrderService,
	obyteProvider interfaces.ObyteProvider,
) {

	e := &accountEndpoint{accountService, orderService, obyteProvider}
	r.HandleFunc("/account/create", e.handleCreateAccount).Methods("POST")
	r.HandleFunc("/account/{address}", e.handleGetAccount).Methods("GET")
	r.HandleFunc("/account/authorized_addresses/{address}", e.handleGetAuthorizedAddresses).Methods("GET")
	r.HandleFunc("/account/balances/{address}", e.handleGetAccountTokenBalances).Methods("GET")
	r.HandleFunc("/account/{address}/{token}", e.handleGetAccountTokenBalance).Methods("GET")
}

func (e *accountEndpoint) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	addr := v.Get("address")

	if !isValidAddress(addr) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	a := addr
	existingAccount, err := e.accountService.GetByAddress(a)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if existingAccount != nil {
		httputils.WriteJSON(w, http.StatusOK, "Account already exists")
		return
	}

	acc := &types.Account{Address: a}
	err = e.accountService.Create(acc)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusCreated, acc)
}

func (e *accountEndpoint) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	addr := vars["address"]
	if !isValidAddress(addr) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	address := addr
	a, err := e.accountService.GetByAddress(address)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, a)
}

func (e *accountEndpoint) handleGetAccountTokenBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	a := vars["address"]
	if !isValidAddress(a) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	t := vars["token"]
	if !isValidAsset(t) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid token asset")
		return
	}

	addr := a
	tokenAsset := t

	b, err := e.accountService.GetTokenBalance(addr, tokenAsset)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, b)
}

func (e *accountEndpoint) handleGetAccountTokenBalances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	address := vars["address"]
	if !isValidAddress(address) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	balances := e.obyteProvider.GetBalances(address)
	logger.Info(balances)
	balances = e.orderService.AdjustBalancesForUncommittedTrades(address, balances)

	httputils.WriteJSON(w, http.StatusOK, balances)
}

func (e *accountEndpoint) handleGetAuthorizedAddresses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	address := vars["address"]
	if !isValidAddress(address) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	authorizedAddresses, err := e.obyteProvider.GetAuthorizedAddresses(address)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}
	logger.Info(authorizedAddresses)

	httputils.WriteJSON(w, http.StatusOK, authorizedAddresses)
}

func isValidAddress(address string) bool {
	return len(address) == 32
}

func isValidAsset(asset string) bool {
	return len(asset) == 44 || asset == "base"
}
