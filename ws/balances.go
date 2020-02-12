package ws

func SendBalancesMessage(msgType string, address string, balances map[string]int64, event string) {
	logger.Info("SendBalancesMessage", address, balances, event)
	conn := GetOrderConnections(address)
	if conn == nil {
		return
	}

	payload := map[string]interface{}{
		"balances": balances,
		"event":    event,
	}

	for _, c := range conn {
		go c.SendMessage(BalancesChannel, msgType, payload)
	}
}
