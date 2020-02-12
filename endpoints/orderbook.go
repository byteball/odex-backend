package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/httputils"
	"github.com/byteball/odex-backend/ws"
	"github.com/gorilla/mux"
)

type OrderBookEndpoint struct {
	orderBookService interfaces.OrderBookService
}

// ServePairResource sets up the routing of pair endpoints and the corresponding handlers.
func ServeOrderBookResource(
	r *mux.Router,
	orderBookService interfaces.OrderBookService,
) {
	e := &OrderBookEndpoint{orderBookService}
	r.HandleFunc("/orderbook/raw", e.handleGetRawOrderBook)
	r.HandleFunc("/orderbook", e.handleGetOrderBook)
	ws.RegisterChannel(ws.OrderBookChannel, e.orderBookWebSocket)
	ws.RegisterChannel(ws.RawOrderBookChannel, e.rawOrderBookWebSocket)
}

// orderBookEndpoint
func (e *OrderBookEndpoint) handleGetOrderBook(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	bt := v.Get("baseToken")
	qt := v.Get("quoteToken")

	if bt == "" {
		httputils.WriteError(w, http.StatusBadRequest, "baseToken Parameter missing")
		return
	}

	if qt == "" {
		httputils.WriteError(w, http.StatusBadRequest, "quoteToken Parameter missing")
		return
	}

	if !isValidAsset(bt) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Base Token Asset")
		return
	}

	if !isValidAsset(qt) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Quote Token Asset")
		return
	}

	baseAsset := bt
	quoteAsset := qt
	ob, err := e.orderBookService.GetOrderBook(baseAsset, quoteAsset)
	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, ob)
}

// orderBookEndpoint
func (e *OrderBookEndpoint) handleGetRawOrderBook(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	bt := v.Get("baseToken")
	qt := v.Get("quoteToken")

	if bt == "" {
		httputils.WriteError(w, http.StatusBadRequest, "baseToken Parameter missing")
		return
	}

	if qt == "" {
		httputils.WriteError(w, http.StatusBadRequest, "quoteToken Parameter missing")
		return
	}

	if !isValidAsset(bt) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Base Token Asset")
		return
	}

	if !isValidAsset(qt) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Quote Token Asset")
		return
	}

	baseAsset := bt
	quoteAsset := qt
	ob, err := e.orderBookService.GetRawOrderBook(baseAsset, quoteAsset)
	if err != nil {
		httputils.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	httputils.WriteJSON(w, http.StatusOK, ob)
}

// liteOrderBookWebSocket
func (e *OrderBookEndpoint) rawOrderBookWebSocket(input interface{}, c *ws.Client) {
	b, _ := json.Marshal(input)
	var ev *types.WebsocketEvent

	err := json.Unmarshal(b, &ev)
	if err != nil {
		logger.Error(err)
		return
	}

	socket := ws.GetRawOrderBookSocket()

	b, _ = json.Marshal(ev.Payload)
	var p *types.SubscriptionPayload

	err = json.Unmarshal(b, &p)
	if err != nil {
		logger.Error(err)
	}

	if ev.Type == "UNSUBSCRIBE" {
		e.orderBookService.UnsubscribeRawOrderBook(c)
		return
	}

	if p.BaseToken == "" {
		msg := map[string]string{"Message": "Invalid Base Token"}
		socket.SendErrorMessage(c, msg)
		return
	}

	if p.QuoteToken == "" {
		msg := map[string]string{"Message": "Invalid Quote Token"}
		socket.SendErrorMessage(c, msg)
		return
	}

	if ev.Type == "SUBSCRIBE" {
		e.orderBookService.SubscribeRawOrderBook(c, p.BaseToken, p.QuoteToken)
	}
}

func (e *OrderBookEndpoint) orderBookWebSocket(input interface{}, c *ws.Client) {
	b, _ := json.Marshal(input)
	var ev *types.WebsocketEvent
	err := json.Unmarshal(b, &ev)
	if err != nil {
		logger.Error(err)
		return
	}

	socket := ws.GetOrderBookSocket()

	b, _ = json.Marshal(ev.Payload)
	var p *types.SubscriptionPayload
	err = json.Unmarshal(b, &p)
	if err != nil {
		logger.Error(err)
		msg := map[string]string{"Message": "Internal server error"}
		socket.SendErrorMessage(c, msg)
		return
	}

	if ev.Type == "UNSUBSCRIBE" {
		e.orderBookService.UnsubscribeOrderBook(c)
		return
	}

	if p.BaseToken == "" {
		msg := map[string]string{"Message": "Empty base token in orderbook"}
		socket.SendErrorMessage(c, msg)
		return
	}

	if p.QuoteToken == "" {
		msg := map[string]string{"Message": "Empty quote token in orderbook"}
		socket.SendErrorMessage(c, msg)
		return
	}

	if ev.Type == "SUBSCRIBE" {
		e.orderBookService.SubscribeOrderBook(c, p.BaseToken, p.QuoteToken)
	}

}
