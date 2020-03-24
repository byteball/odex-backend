package endpoints

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/byteball/odex-backend/interfaces"
	"github.com/byteball/odex-backend/utils/httputils"
	"github.com/gorilla/mux"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/ws"
)

type orderEndpoint struct {
	orderService   interfaces.OrderService
	accountService interfaces.AccountService
	obyteProvider  interfaces.ObyteProvider
}

// ServeOrderResource sets up the routing of order endpoints and the corresponding handlers.
func ServeOrderResource(
	r *mux.Router,
	orderService interfaces.OrderService,
	accountService interfaces.AccountService,
	obyteProvider interfaces.ObyteProvider,
) {
	e := &orderEndpoint{orderService, accountService, obyteProvider}
	r.HandleFunc("/orders/history", e.handleGetOrderHistory).Methods("GET")
	r.HandleFunc("/orders/current", e.handleGetCurrentOrders).Methods("GET")
	r.HandleFunc("/orders/positions", e.handleGetCurrentOrders).Methods("GET")
	r.HandleFunc("/orders", e.handleGetOrders).Methods("GET")
	ws.RegisterChannel(ws.OrderChannel, e.ws)
}

func (e *orderEndpoint) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	addr := v.Get("address")
	limit := v.Get("limit")

	if addr == "" {
		httputils.WriteError(w, http.StatusBadRequest, "address Parameter Missing")
		return
	}

	if !isValidAddress(addr) {
		httputils.WriteError(w, http.StatusBadRequest, "Invalid Address")
		return
	}

	var err error
	var orders []*types.Order
	address := addr

	if limit == "" {
		orders, err = e.orderService.GetByUserAddress(address)
	} else {
		lim, _ := strconv.Atoi(limit)
		orders, err = e.orderService.GetByUserAddress(address, lim)
	}

	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if orders == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Order{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, orders)
}

func (e *orderEndpoint) handleGetCurrentOrders(w http.ResponseWriter, r *http.Request) {
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

	var err error
	var orders []*types.Order
	address := addr

	if limit == "" {
		orders, err = e.orderService.GetCurrentByUserAddress(address)
	} else {
		lim, _ := strconv.Atoi(limit)
		orders, err = e.orderService.GetCurrentByUserAddress(address, lim)
	}

	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "")
		return
	}

	if orders == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Order{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, orders)
}

func (e *orderEndpoint) handleGetOrderHistory(w http.ResponseWriter, r *http.Request) {
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

	var err error
	var orders []*types.Order
	address := addr

	if limit == "" {
		orders, err = e.orderService.GetHistoryByUserAddress(address)
	} else {
		lim, _ := strconv.Atoi(limit)
		orders, err = e.orderService.GetHistoryByUserAddress(address, lim)
	}

	if err != nil {
		logger.Error(err)
		httputils.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if orders == nil {
		httputils.WriteJSON(w, http.StatusOK, []types.Order{})
		return
	}

	httputils.WriteJSON(w, http.StatusOK, orders)
}

// ws function handles incoming websocket messages on the order channel
func (e *orderEndpoint) ws(input interface{}, c *ws.Client) {
	msg := &types.WebsocketEvent{}

	bytes, _ := json.Marshal(input)
	if err := json.Unmarshal(bytes, &msg); err != nil {
		logger.Error(err)
		go c.SendMessage(ws.OrderChannel, "ERROR", err.Error())
	}

	switch msg.Type {
	case "ADDRESS":
		e.handleAddress(msg, c)
	case "NEW_ORDER":
		e.handleNewOrder(msg, c)
	case "CANCEL_ORDER":
		e.handleCancelOrder(msg, c)
	default:
		log.Print("Response with error")
	}
}

// handleNewOrder handles NewOrder message. New order messages are transmitted to the order service after being unmarshalled
func (e *orderEndpoint) handleNewOrder(ev *types.WebsocketEvent, c *ws.Client) {
	c.RpcMutex.Lock()
	_, err := e.obyteProvider.AddOrder(&ev.Payload)
	c.RpcMutex.Unlock()
	if err != nil {
		logger.Error(err)
		//payload := map[string]string{"hash": hash, "error": err.Error()}
		payload := err.Error()
		go c.SendMessage(ws.OrderChannel, "ERROR", payload)
		return
	}
	/*o := &types.Order{}

	bytes, err := json.Marshal(ev.Payload)
	if err != nil {
		logger.Error(err)
		go c.SendMessage(ws.OrderChannel, "ERROR", err.Error())
		return
	}

	err = json.Unmarshal(bytes, &o)
	if err != nil {
		logger.Error(err)
		c.SendOrderErrorMessage(err, o.Hash)
		return
	}

	//o.Hash = o.ComputeHash()
	ws.RegisterOrderConnection(o.UserAddress, c)

	acc, err := e.accountService.FindOrCreate(o.UserAddress)
	if err != nil {
		logger.Error(err)
		c.SendOrderErrorMessage(err, o.Hash)
	}

	if acc.IsBlocked {
		go c.SendMessage(ws.OrderChannel, "ERROR", errors.New("Account is blocked"))
	}

	err = e.orderService.NewOrder(o)
	if err != nil {
		logger.Error(err)
		c.SendOrderErrorMessage(err, o.Hash)
		return
	}*/
}

// handleCancelOrder handles CancelOrder message.
func (e *orderEndpoint) handleCancelOrder(ev *types.WebsocketEvent, c *ws.Client) {
	c.RpcMutex.Lock()
	err := e.obyteProvider.CancelOrder(&ev.Payload)
	c.RpcMutex.Unlock()
	if err != nil {
		logger.Error(err)
		go c.SendMessage(ws.OrderChannel, "ERROR", err.Error())
		return
	}

	/*bytes, err := json.Marshal(ev.Payload)
	oc := &types.OrderCancel{}

	err = oc.UnmarshalJSON(bytes)
	if err != nil {
		logger.Error(err)
		c.SendOrderErrorMessage(err, oc.OrderHash)
		return
	}

	addr, err := e.orderService.GetSenderAddress(oc)
	if err != nil {
		logger.Error(err)
		c.SendOrderErrorMessage(err, oc.OrderHash)
		return
	}

	//ws.RegisterOrderConnection(addr, c)
	if !ws.IsClientConnected(addr, c) {
		err := errors.New("client not connected to this address")
		logger.Error(err)
		c.SendOrderErrorMessage(err, oc.OrderHash)
		return
	}

	if !ws.IsClientConnectedToSession(oc.SessionId, c) {
		err := errors.New("client not connected to this session id")
		logger.Error(err)
		c.SendOrderErrorMessage(err, oc.OrderHash)
		return
	}

	orderErr := e.orderService.CancelOrder(oc)
	if orderErr != nil {
		logger.Error(err)
		c.SendOrderErrorMessage(orderErr, oc.OrderHash)
		return
	}*/
}

func (e *orderEndpoint) handleAddress(ev *types.WebsocketEvent, c *ws.Client) {
	if reflect.TypeOf(ev.Payload).Kind() != reflect.String {
		logger.Error("bad type of payload")
		go c.SendMessage(ws.OrderChannel, "ERROR", errors.New("bad type of payload"))
		return
	}

	address := ev.Payload.(string)

	ws.RegisterOrderConnection(address, c)

	acc, err := e.accountService.FindOrCreate(address)
	if err != nil {
		logger.Error(err)
		go c.SendMessage(ws.OrderChannel, "ERROR", err)
		return
	}

	if acc.IsBlocked {
		go c.SendMessage(ws.OrderChannel, "ERROR", errors.New("Account is blocked"))
	}
}
