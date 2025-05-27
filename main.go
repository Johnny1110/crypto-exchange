package main

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/handlers"
	"github.com/johnny1110/crypto-exchange/market"
	"github.com/johnny1110/crypto-exchange/service"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./exg.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	markets := initMarkets()
	engine, err := core.NewMatchingEngine(markets)

	orderService := &service.OrderService{
		DB:     db,
		Engine: engine,
	}

	if err != nil {
		log.Fatalf("failed to init exchange core: %v", err)
	}

	r := gin.Default()
	// inject db into context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("engine", engine)
		c.Set("orderService", orderService)
		c.Next()
	})

	registerRouter(r)

	r.Run(":8080")
}

// initMarkets define markets here.
func initMarkets() []*market.MarketInfo {
	btcMarket := market.NewMarketInfo("BTC-USDT", "BTC", "USDT")
	ethMarket := market.NewMarketInfo("ETH-USDT", "ETH", "USDT")
	dotMarket := market.NewMarketInfo("DOT-USDT", "DOT", "USDT")

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
	auth.POST("/orders/:market", handlers.PlaceOrderHandler)
	auth.DELETE("/orders/:market/:orderID", handlers.CancelOrderHandler)
}

func healthcheck(context *gin.Context) {
	context.JSON(http.StatusOK, "OK")
}

func getEthClient() (*ethclient.Client, error) {
	return ethclient.Dial("http://localhost:8545")
}
