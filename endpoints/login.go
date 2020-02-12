package endpoints

import (
	"encoding/json"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/ws"
	"github.com/gorilla/mux"
)

type loginEndpoint struct {
}

// ServeLoginResource sets up the routing of login endpoints and the corresponding handlers.
func ServeLoginResource(
	r *mux.Router,
) {
	e := &loginEndpoint{}
	ws.RegisterChannel(ws.LoginChannel, e.loginWebsocket)
}

func (e *loginEndpoint) loginWebsocket(input interface{}, c *ws.Client) {
	b, _ := json.Marshal(input)
	var ev *types.WebsocketEvent
	if err := json.Unmarshal(b, &ev); err != nil {
		logger.Error(err)
		return
	}

	socket := ws.GetLoginSocket()
	if ev.Type != "SUBSCRIBE" && ev.Type != "UNSUBSCRIBE" {
		logger.Info("Event Type", ev.Type)
		err := map[string]string{"Message": "Invalid payload"}
		socket.SendErrorMessage(c, err)
		return
	}

	b, _ = json.Marshal(ev.Payload)
	var sessionId string
	err := json.Unmarshal(b, &sessionId)
	if err != nil {
		logger.Error(err)
		return
	}

	if ev.Type == "SUBSCRIBE" {
		if sessionId == "" {
			err := map[string]string{"Message": "Invalid sessionId"}
			socket.SendErrorMessage(c, err)
			return
		}

		socket.Subscribe(sessionId, c)
	}

	if ev.Type == "UNSUBSCRIBE" {
		socket.Unsubscribe(c)
	}
}
