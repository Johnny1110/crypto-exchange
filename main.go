package main

import (
	"github.com/johnny1110/crypto-exchange/container"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
)

func main() {
	err := initLogger("logs", log.INFO)
	if err != nil {
		panic(err)
	}

	db, err := initDB(false)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	engine, err := core.NewMatchingEngine(settings.ALL_MARKETS)

	if err != nil {
		log.Fatalf("failed to init matching-engine: %v", err)
	}

	c := container.NewContainer(db, engine)
	defer c.Cleanup()

	router := setupRouter(c)
	setupWebSocket(c)

	// Recover OrderBook from db data.
	err = recoverOrderBook(c)
	if err != nil {
		log.Fatalf("failed to recover orderbook: %v", err)
	}

	startUpAllScheduler(c)

	log.Infof("Exchange Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
