package ws

import (
	sync "github.com/sasha-s/go-deadlock"
)

// OrderConn is websocket order connection struct
// It holds the reference to connection and the channel of type OrderMessage

type OrderConnection []*Client

var orderConnections map[string]OrderConnection
var mutex sync.Mutex

// GetOrderConn returns the connection associated with an order ID
func GetOrderConnections(address string) OrderConnection {
	mutex.Lock()
	defer mutex.Unlock()
	c := orderConnections[address]
	if c == nil {
		logger.Warning("No connection found for address", address)
		return nil
	}

	return orderConnections[address]
}

func OrderSocketUnsubscribeHandler(a string) func(client *Client) {
	return func(client *Client) {
		logger.Info("In unsubscription handler")
		mutex.Lock()
		defer mutex.Unlock()
		orderConnection := orderConnections[a]
		if orderConnection == nil {
			logger.Info("No subscriptions")
		}

		if orderConnection != nil {
			logger.Info("%v connections before unsubscription", len(orderConnections[a]))
			for i, c := range orderConnection {
				if client == c {
					orderConnection = append(orderConnection[:i], orderConnection[i+1:]...)
				}
			}

		}

		orderConnections[a] = orderConnection
		logger.Info("%v connections after unsubscription", len(orderConnections[a]))
	}
}

// RegisterOrderConnection registers a connection with and orderID.
// It is called whenever a message is recieved over order channel
func RegisterOrderConnection(a string, c *Client) {
	logger.Info("Registering new order connection")
	mutex.Lock()
	defer mutex.Unlock()

	if orderConnections == nil {
		orderConnections = make(map[string]OrderConnection)
	}

	if orderConnections[a] == nil {
		logger.Info("Registering a new order connection")
		orderConnections[a] = OrderConnection{c}
		RegisterConnectionUnsubscribeHandler(c, OrderSocketUnsubscribeHandler(a))
		logger.Info("Number of connections for %s: %v", a, len(orderConnections))
	}

	if orderConnections[a] != nil {
		if !IsClientConnected(a, c) {
			logger.Info("Registering a new order connection")
			orderConnections[a] = append(orderConnections[a], c)
			RegisterConnectionUnsubscribeHandler(c, OrderSocketUnsubscribeHandler(a))
			logger.Info("Number of connections for %s: %v", a, len(orderConnections))
		}
	}
}

func IsClientConnected(a string, client *Client) bool {
	for _, c := range orderConnections[a] {
		if c == client {
			logger.Info("Client is connected")
			return true
		}
	}

	logger.Info("Client is not connected")
	return false
}

func SendOrderMessage(msgType string, a string, payload interface{}) {
	conn := GetOrderConnections(a)
	if conn == nil {
		return
	}

	for _, c := range conn {
		go c.SendMessage(OrderChannel, msgType, payload)
	}
}
