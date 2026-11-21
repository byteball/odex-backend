package ws

import (
	sync "github.com/sasha-s/go-deadlock"

	"github.com/byteball/odex-backend/types"
	"github.com/gorilla/websocket"
)

type Client struct {
	*websocket.Conn
	mu       sync.Mutex
	RpcMutex sync.Mutex
	send     chan types.WebsocketMessage
	done     chan struct{}
	closed   bool
}

// sendBufferSize bounds the memory held for a client that reads slower than we publish.
const sendBufferSize = 256

// TODO: refactor into non-global variables
var unsubscribeHandlers map[*Client][]func(*Client)
var subscriptionMutex sync.Mutex

func NewClient(c *websocket.Conn) *Client {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()
	conn := &Client{Conn: c, mu: sync.Mutex{}, RpcMutex: sync.Mutex{}, send: make(chan types.WebsocketMessage, sendBufferSize), done: make(chan struct{}), closed: false}

	if unsubscribeHandlers == nil {
		unsubscribeHandlers = make(map[*Client][]func(*Client))
	}

	if unsubscribeHandlers[conn] == nil {
		unsubscribeHandlers[conn] = make([]func(*Client), 0)
	}

	return conn
}

// SendMessage constructs the message with proper structure to be sent over websocket
func (c *Client) SendMessage(channel string, msgType string, payload interface{}, h ...string) {
	e := types.WebsocketEvent{
		Type:    msgType,
		Payload: payload,
	}

	if len(h) > 0 {
		e.Hash = h[0]
	}

	m := types.WebsocketMessage{
		Channel: channel,
		Event:   e,
	}

	c.enqueue(m)
}

// enqueue never blocks: a client whose buffer is full is dropped rather than letting messages pile up.
func (c *Client) enqueue(m types.WebsocketMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		logger.Info("trying to send to a closed ws connection")
		return
	}

	select {
	case c.send <- m:
	default:
		logger.Warning("ws client send buffer is full, closing the connection")
		c.closed = true
		go c.closeConnection()
	}
}

func (c *Client) closeConnection() {
	subscriptionMutex.Lock()
	handlers := unsubscribeHandlers[c]
	delete(unsubscribeHandlers, c)
	subscriptionMutex.Unlock()

	for _, unsub := range handlers {
		unsub(c)
	}

	c.mu.Lock()
	c.closed = true
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	c.mu.Unlock()

	c.Close()
}

func (c *Client) SendOrderErrorMessage(err error, h string) {
	p := map[string]interface{}{
		"message": err.Error(),
		"hash":    h,
	}

	e := types.WebsocketEvent{
		Type:    "ERROR",
		Payload: p,
	}

	m := types.WebsocketMessage{
		Channel: OrderChannel,
		Event:   e,
	}

	c.enqueue(m)
}
