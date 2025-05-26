package main

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/handlers"
	"github.com/johnny1110/crypto-exchange/market"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const (
	hotWalletAddress    = "0x6812A4DB987C1bdaf14be68F8acD9c17948622b9"
	hotWalletPrivateKey = "b36b2f4b16ffbbbdb0369b597ce12ed24aaa729aa193cb12b6b070f4db566c7c"
)

//func main() {
//	e := echo.New()
//	e.HTTPErrorHandler = httpErrorHandler
//
//	ethClient, err := getEthClient()
//	if err != nil {
//		e.Logger.Fatal(err)
//	}
//	ex, err := exchange.NewMatchingEngine(ethClient, hotWalletAddress, hotWalletPrivateKey)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ex.InitOrderbooks()
//	e.GET("/healthcheck", func(c echo.Context) error { return c.String(http.StatusOK, "OK") })
//	e.POST("/order", ex.HandlePlaceOrder)
//	e.DELETE("/orderbook/:market/order/:id", ex.HandleDeleteOrder)
//	e.GET("/orderbook/:market", ex.HandleGetOrderBook)
//	e.GET("/orderbook/:market/orderIds", ex.HandleGetOrderIds)
//	e.POST("/user/register", ex.RegisterUser)
//	e.GET("/user/:userId/symbol/:symbol/balance", ex.QueryBalance)
//
//	e.Start(":3000")
//}

func main() {
	db, err := sql.Open("sqlite3", "./exg.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	markets := initMarkets()
	engine, err := core.NewMatchingEngine(markets)

	if err != nil {
		log.Fatalf("failed to init exchange core: %v", err)
	}

	r := gin.Default()
	// inject db into context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("engine", engine)
		c.Next()
	})

	registerRouter(r)

	r.Run(":8080")
}

// initMarkets define markets here.
func initMarkets() []*market.MarketInfo {
	btcMarket := market.NewMarketInfo("BTC/USDT", "BTC", "USDT")
	ethMarket := market.NewMarketInfo("ETH/USDT", "ETH", "USDT")
	dotMarket := market.NewMarketInfo("DOT/USDT", "DOT", "USDT")

	return []*market.MarketInfo{btcMarket, ethMarket, dotMarket}
}

func registerRouter(r *gin.Engine) {
	r.GET("/healthcheck", healthcheck)

	// user account
	r.POST("/users/register", handlers.Register)
	r.POST("/users/login", handlers.Login)

	// admin access
	r.POST("/admin/manual-adjustment", handlers.ManualAdjustment)

	// auth middleware
	auth := r.Group("/", handlers.AuthMiddleware)
	// auth protected
	auth.DELETE("/users/logout", handlers.Logout)
	auth.GET("/balances", handlers.GetBalance)
}

func healthcheck(context *gin.Context) {
	context.JSON(http.StatusOK, "OK")
}

func getEthClient() (*ethclient.Client, error) {
	return ethclient.Dial("http://localhost:8545")
}

//func httpErrorHandler(err error, c echo.Context) {
//	fmt.Println(err)
//}
