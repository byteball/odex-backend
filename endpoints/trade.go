package endpoints

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/httputils"
	"github.com/byteball/odex-backend/ws"
	"github.com/gorilla/mux"
)

type tradeEndpoint struct {
	tradeService interfaces.TradeService
}

// ServeTradeResource sets up the routing of trade endpoints and the corresponding handlers.
// TODO trim down to one single endpoint with the 3 following params: base, quote, address
func ServeTradeResource(
	r *mux.Router,
	tradeService interfaces.TradeService,
) {
	e := &tradeEndpoint{tradeService}
	r.HandleFunc("/trades/pair", e.HandleGetTradeHistory)
	r.HandleFunc("/trades", e.HandleGetTrades).Methods("GET")
	ws.RegisterChannel(ws.TradeChannel, e.tradeWebsocket)
}

// history is reponsible for handling pair's trade history requests
func (e *tradeEndpoint) HandleGetTradeHistory(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	bt := v.Get("baseToken")
	qt := v.Get("quoteToken")
	l := v.Get("limit")

	if bt == "" {
		httputils.WriteError(w, http.StatusBadRequest, "baseToken Parameter missing")
		return
	}

	if qt == "" {
		httputils.WriteError(w, http.StatusBadRequest, "quoteToken Parameter missing")
		return
	}

	if !isValidAsset(bt) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid base token asset in trade")
		return
	}

	if !isValidAsset(qt) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid quote token asset in trade")
		return
	}

	limit := 20
	if l != "" {
		limit, _ = strconv.Atoi(l)
	}

	baseToken := bt
	quoteToken := qt
	res, err := e.tradeService.GetSortedTrades(baseToken, quoteToken, limit)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if res == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Trade{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

// get is reponsible for handling user's trade history requests
func (e *tradeEndpoint) HandleGetTrades(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	addr := v.Get("address")
	limit := v.Get("limit")

	if addr == "" {
		httputils.WriteError(w, http.StatusBadRequest, "address Parameter missing")
		return
	}

	if !isValidAddress(addr) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	lim := 100
	if limit != "" {
		lim, _ = strconv.Atoi(limit)
	}

	address := addr
	res, err := e.tradeService.GetSortedTradesByUserAddress(address, lim)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if res == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Trade{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, res)
}

func (e *tradeEndpoint) tradeWebsocket(input interface{}, c *ws.Client) {
	b, _ := json.Marshal(input)
	var ev *types.WebsocketEvent
	if err := json.Unmarshal(b, &ev); err != nil {
		logger.Error(err)
		return
	}

	socket := ws.GetTradeSocket()
	if ev.Type != "SUBSCRIBE" && ev.Type != "UNSUBSCRIBE" {
		logger.Info("Event Type", ev.Type)
		err := map[string]string{"Message": "Invalid payload"}
		socket.SendErrorMessage(c, err)
		return
	}

	b, _ = json.Marshal(ev.Payload)
	var p *types.SubscriptionPayload
	err := json.Unmarshal(b, &p)
	if err != nil {
		logger.Error(err)
		return
	}

	if ev.Type == "SUBSCRIBE" {
		if p.BaseToken == "" {
			err := map[string]string{"Message": "Empty base token in trade"}
			socket.SendErrorMessage(c, err)
			return
		}

		if p.QuoteToken == "" {
			err := map[string]string{"Message": "Empty quote token in trade"}
			socket.SendErrorMessage(c, err)
			return
		}

		e.tradeService.Subscribe(c, p.BaseToken, p.QuoteToken)
	}

	if ev.Type == "UNSUBSCRIBE" {
		if p == nil {
			e.tradeService.Unsubscribe(c)
			return
		}

		e.tradeService.UnsubscribeChannel(c, p.BaseToken, p.QuoteToken)
	}
}
