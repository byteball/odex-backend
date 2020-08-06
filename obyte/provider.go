package obyte

import (
	"encoding/json"
	"log"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
	"github.com/gorilla/websocket"
	"github.com/ybbus/jsonrpc"
)

type ObyteProvider struct {
	Client   jsonrpc.RPCClient
	WSClient *websocket.Conn
}

var operatorAddress string
var matcherFee float64
var affiliateFee float64
var feesUpdated bool

func NewObyteProvider() *ObyteProvider {
	client := jsonrpc.NewClient(app.Config.Obyte["http_url"])
	wsconn, _, err := websocket.DefaultDialer.Dial(app.Config.Obyte["ws_url"], nil)

	if err != nil {
		panic(err)
	}

	return &ObyteProvider{
		Client:   client,
		WSClient: wsconn,
	}
}

func (o *ObyteProvider) BalanceOf(owner string, token string) (int64, error) {
	var balance int64
	err := utils.Retry(3, func() error {
		err := o.Client.CallFor(&balance, "getBalance", owner, token)
		if err != nil {
			log.Println("error from getBalance: ", err)
		}
		return err
	})
	if err != nil {
		panic(err)
	}
	return balance, nil
}

func (o *ObyteProvider) GetBalances(owner string) map[string]int64 {
	type Balances struct {
		BalancesByAsset  map[string]int64 `json:"balances_by_asset"`
		BalancesBySymbol map[string]int64 `json:"balances_by_symbol"`
	}
	var balances Balances
	err := o.Client.CallFor(&balances, "getBalances", owner)
	if err != nil {
		panic(err)
	}
	log.Print("balances", balances)
	/*if balances == nil {
		log.Print("nil map")
		balances = make(map[string]int64)
	}*/
	return balances.BalancesBySymbol
}

func (o *ObyteProvider) GetOperatorAddress() string {
	if operatorAddress != "" {
		return operatorAddress
	}
	err := o.Client.CallFor(&operatorAddress, "getOperatorAddress", "x")
	if err != nil {
		panic(err)
	}
	return operatorAddress
}

func (o *ObyteProvider) GetFees() (float64, float64) {
	if feesUpdated {
		return matcherFee, affiliateFee
	}
	var fees [2]float64
	err := o.Client.CallFor(&fees, "getFees", "x")
	if err != nil {
		panic(err)
	}
	matcherFee = fees[0]
	affiliateFee = fees[1]
	feesUpdated = true
	return matcherFee, affiliateFee
}

func (o *ObyteProvider) Symbol(token string) (string, error) {
	var symbol string
	err := o.Client.CallFor(&symbol, "getSymbol", token)
	if err != nil {
		log.Println(err)
	}

	return symbol, err
}

func (o *ObyteProvider) Asset(symbol string) (string, error) {
	var asset string
	err := o.Client.CallFor(&asset, "getAsset", symbol)
	if err != nil {
		log.Println(err)
	}

	return asset, err
}

func (o *ObyteProvider) Decimals(token string) (uint8, error) {
	var decimals uint8
	err := o.Client.CallFor(&decimals, "getDecimals", token)
	if err != nil {
		log.Println(err)
	}

	return decimals, err
}

func (o *ObyteProvider) AddOrder(signedOrder *interface{}) (string, error) {
	log.Println("will rpc addOrder", utils.JSON(signedOrder))
	var hash string // order hash
	err := utils.Retry(3, func() error {
		err := o.Client.CallFor(&hash, "addOrder", signedOrder)
		if err != nil {
			log.Println("error from addOrder: ", err)
		}
		return err
	})
	/*if err != nil {
		panic(err)
	}*/

	return hash, err
}

func (o *ObyteProvider) CancelOrder(signedCancel *interface{}) error {
	log.Println("will rpc cancelOrder", utils.JSON(signedCancel))
	var resp string
	err := utils.Retry(3, func() error {
		err := o.Client.CallFor(&resp, "cancelOrder", signedCancel)
		if err != nil {
			log.Println("error from cancelOrder: ", err)
		}
		return err
	})

	return err
}

func (o *ObyteProvider) GetAuthorizedAddresses(address string) ([]string, error) {
	var authorizedAddresses []string
	err := o.Client.CallFor(&authorizedAddresses, "getAuthorizedAddresses", address)

	return authorizedAddresses, err
}

/*func (o *ObyteProvider) VerifySignature(order *types.Order) (string, error) {
	var id string // order hash
	err := o.Client.CallFor(&id, "verifySignature", order)

	return id, err
}

func (o *ObyteProvider) VerifyCancelSignature(order *types.OrderCancel) (string, error) {
	var addr string // who signed
	err := o.Client.CallFor(&addr, "verifyCancelSignature", order)

	return addr, err
}*/

func (o *ObyteProvider) ExecuteTrade(m *types.Matches) ([]string, error) {
	var arrTriggerUnits []string
	err := o.Client.CallFor(&arrTriggerUnits, "executeTrade", m)

	return arrTriggerUnits, err
}

func (o *ObyteProvider) ListenToEvents() (chan map[string]interface{}, error) {
	events := make(chan map[string]interface{})

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := o.WSClient.ReadMessage()
			if err != nil {
				panic(err)
				//log.Println("read:", err)
				//return
			}
			log.Printf("recv from wallet: %s", message)
			var ev map[string]interface{}
			err = json.Unmarshal(message, &ev)
			if err != nil {
				panic(err)
			}
			events <- ev
		}
	}()
	// to do
	return events, nil
}
