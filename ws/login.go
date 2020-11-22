package ws

import (
	"errors"
	sync "github.com/sasha-s/go-deadlock"
)

var loginSocket *LoginSocket

// LoginSocket holds the map of connections subscribed to session ids
// corresponding to the key/event they have subscribed to.
type LoginSocket struct {
	subscriptions     map[string]map[*Client]bool
	subscriptionsList map[*Client][]string
	mu                sync.Mutex
}

func NewLoginSocket() *LoginSocket {
	return &LoginSocket{
		subscriptions:     make(map[string]map[*Client]bool),
		subscriptionsList: make(map[*Client][]string),
		mu:                sync.Mutex{},
	}
}

func GetLoginSocket() *LoginSocket {
	if loginSocket == nil {
		loginSocket = NewLoginSocket()
	}

	return loginSocket
}

// Subscribe registers a new websocket connections to the login channel updates
func (s *LoginSocket) Subscribe(sessionId string, c *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c == nil {
		return errors.New("No connection found")
	}

	if s.subscriptions[sessionId] == nil {
		s.subscriptions[sessionId] = make(map[*Client]bool)
	}

	s.subscriptions[sessionId][c] = true

	if s.subscriptionsList[c] == nil {
		s.subscriptionsList[c] = []string{}
	}

	s.subscriptionsList[c] = append(s.subscriptionsList[c], sessionId)

	return nil
}

// UnsubscribeHandler unsubscribes a connection from a certain login channel id
func (s *LoginSocket) UnsubscribeChannelHandler(sessionId string) func(c *Client) {
	return func(c *Client) {
		s.UnsubscribeChannel(sessionId, c)
	}
}

func (s *LoginSocket) UnsubscribeHandler() func(c *Client) {
	return func(c *Client) {
		s.Unsubscribe(c)
	}
}

// Unsubscribe removes a websocket connection from the login channel updates
func (s *LoginSocket) UnsubscribeChannel(sessionId string, c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subscriptions[sessionId][c] {
		s.subscriptions[sessionId][c] = false
		delete(s.subscriptions[sessionId], c)
	}
}

func (s *LoginSocket) Unsubscribe(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionIds := s.subscriptionsList[c]
	if sessionIds == nil {
		return
	}

	for _, id := range s.subscriptionsList[c] {
		if s.subscriptions[id][c] {
			s.subscriptions[id][c] = false
			delete(s.subscriptions[id], c)
		}
	}
}

func IsClientConnectedToSession(sessionId string, c *Client) bool {
	GetLoginSocket()
	if loginSocket.subscriptionsList[c] == nil {
		return false
	}

	if loginSocket.subscriptions[sessionId] == nil {
		return false
	}

	if loginSocket.subscriptions[sessionId][c] {
		return true
	}
	return false
}

func (s *LoginSocket) LinkAddressToClient(sessionId string, address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, active := range loginSocket.subscriptions[sessionId] {
		if active {
			RegisterOrderConnection(address, conn)
		}
	}
}

// BroadcastMessage broadcasts login message to all subscribed sockets
func (s *LoginSocket) SendMessageBySession(sessionId string, p interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger.Info("SendMessageBySession on login channel", sessionId, p)

	// go func() {
	for conn, active := range loginSocket.subscriptions[sessionId] {
		if active {
			s.SendUpdateMessage(conn, p)
		}
	}
	// }()
}

// SendMessage sends a websocket message on the login channel
func (s *LoginSocket) SendMessage(c *Client, msgType string, p interface{}) {
	logger.Info("SendMessage on login channel", msgType, p)
	go c.SendMessage(LoginChannel, msgType, p)
}

// SendErrorMessage sends an error message on the login channel
func (s *LoginSocket) SendErrorMessage(c *Client, p interface{}) {
	go c.SendMessage(LoginChannel, "ERROR", p)
}

// SendInitMessage is responsible for sending message on trade ohlcv channel at subscription
func (s *LoginSocket) SendInitMessage(c *Client, p interface{}) {
	go c.SendMessage(LoginChannel, "INIT", p)
}

// SendUpdateMessage is responsible for sending message on trade ohlcv channel at subscription
func (s *LoginSocket) SendUpdateMessage(c *Client, p interface{}) {
	logger.Info("SendUpdateMessage on login channel", p)
	go c.SendMessage(LoginChannel, "UPDATE", p)
}
