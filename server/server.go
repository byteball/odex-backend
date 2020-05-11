package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/daos"
	"github.com/byteball/odex-backend/endpoints"
	"github.com/byteball/odex-backend/errors"
	"github.com/byteball/odex-backend/obyte"
	"github.com/byteball/odex-backend/operator"
	"github.com/byteball/odex-backend/rabbitmq"
	"github.com/byteball/odex-backend/services"
	"github.com/byteball/odex-backend/ws"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"

	"github.com/byteball/odex-backend/engine"
)

func Start() {
	env := os.Getenv("GO_ENV")

	if err := app.LoadConfig("./config", env); err != nil {
		panic(err)
	}

	if err := errors.LoadMessages(app.Config.ErrorFile); err != nil {
		panic(err)
	}

	// connect to the database
	_, err := daos.InitSession(nil)
	if err != nil {
		panic(err)
	}

	rabbitConn := rabbitmq.InitConnection(app.Config.RabbitMQURL)

	provider := obyte.NewObyteProvider()

	router := NewRouter(provider, rabbitConn)
	router.HandleFunc("/socket", ws.ConnectionEndpoint)

	// certManager := autocert.Manager{
	// 	Prompt:     autocert.AcceptTOS,
	// 	HostPolicy: autocert.HostWhitelist("engine.amp.exchange"),
	// 	Cache:      autocert.DirCache("/certs"),
	// }

	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Authorization", "Access-Control-Allow-Origin"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	// start the server
	if app.Config.EnableTLS {
		log.Printf("server %v starting on port :443", app.Version)
		err := http.ListenAndServeTLS(":443",
			"/etc/ssl/matching-engine/server_certificate.pem",
			"/etc/ssl/matching-engine/server_key.pem",
			handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router),
		)

		if err != nil {
			panic(err)
		}
		// server := &http.Server{
		// 	Addr:      ":443",
		// 	Handler:
		// 	TLSConfig: &tls.Config{GetCertificate: certManager.GetCertificate},
		// }

		// go handleCerts(&certManager)

		// err := server.ListenAndServeTLS("", "")
		// if err != nil {
		// 	panic(err)
		// }

	} else {
		address := fmt.Sprintf(":%v", app.Config.ServerPort)
		log.Printf("server %v starting at %v\n", app.Version, address)
		err := http.ListenAndServe(address, handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router))
		if err != nil {
			log.Fatal("The process exited with error:", err.Error())
		}
	}
}

func handleCerts(certManager *autocert.Manager) {
	err := http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	if err != nil {
		log.Print(err)
		panic(err)
	}
}

func NewRouter(
	provider *obyte.ObyteProvider,
	rabbitConn *rabbitmq.Connection,
) *mux.Router {

	r := mux.NewRouter()

	// get daos for dependency injection
	orderDao := daos.NewOrderDao()
	tokenDao := daos.NewTokenDao()
	pairDao := daos.NewPairDao()
	tradeDao := daos.NewTradeDao()
	accountDao := daos.NewAccountDao()

	// get services for injection
	accountService := services.NewAccountService(accountDao, tokenDao)
	ohlcvService := services.NewOHLCVService(tradeDao)
	tokenService := services.NewTokenService(tokenDao, provider)
	tradeService := services.NewTradeService(tradeDao)
	validatorService := services.NewValidatorService(provider, accountDao, orderDao, pairDao)
	priceService := services.NewPriceService()

	infoService := services.NewInfoService(pairDao, tokenDao, tradeDao, orderDao, priceService)
	pairService := services.NewPairService(pairDao, tokenDao, tradeDao, orderDao, provider)
	orderService := services.NewOrderService(orderDao, pairDao, accountDao, tradeDao, validatorService, rabbitConn)
	orderBookService := services.NewOrderBookService(pairDao, tokenDao, orderDao)
	// cronService := crons.NewCronService(ohlcvService)

	// instantiate engine
	eng := engine.NewEngine(rabbitConn, orderDao, tradeDao, pairDao, provider, orderService)

	// deploy operator
	op, err := operator.NewOperator(
		tradeService,
		orderService,
		accountService,
		provider,
		rabbitConn,
	)

	if err != nil {
		panic(err)
	}

	// deploy http and ws endpoints
	endpoints.ServeInfoResource(r, tokenService, infoService, provider)
	endpoints.ServeAccountResource(r, accountService, orderService, provider)
	endpoints.ServeTokenResource(r, tokenService)
	endpoints.ServePairResource(r, pairService, tokenService)
	endpoints.ServeOrderBookResource(r, orderBookService)
	endpoints.ServeOHLCVResource(r, ohlcvService)
	endpoints.ServeTradeResource(r, tradeService)
	endpoints.ServeOrderResource(r, orderService, accountService, provider)
	endpoints.ServeLoginResource(r)

	//initialize rabbitmq subscriptions
	rabbitConn.SubscribeOrders(eng.HandleOrders)
	rabbitConn.SubscribeTrades(op.HandleTrades)
	rabbitConn.SubscribeOperator(orderService.HandleOperatorMessages)
	rabbitConn.SubscribeEngineResponses(orderService.HandleEngineResponse)

	// cronService.InitCrons()
	return r
}
