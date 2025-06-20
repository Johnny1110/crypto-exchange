package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/johnny1110/crypto-exchange/container"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/ws"
	"github.com/labstack/gommon/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func initLogger(logDir string, level log.Lvl) error {
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	logFile, _ := os.OpenFile(
		fmt.Sprintf("%s/app_%s.log", logDir, today),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.SetLevel(level)
	return nil
}

func runSQLFilesWithTransaction(db *sql.DB) error {
	sqlFiles := []string{
		"./doc/db_schema/schema.sql",
		"./doc/db_schema/testing_data.sql",
	}

	for _, filePath := range sqlFiles {
		if err := executeSQLFileWithTransaction(db, filePath); err != nil {
			return fmt.Errorf("failed to execute %s: %w", filePath, err)
		}
		log.Infof("Successfully executed: %s", filePath)
	}

	return nil
}

func executeSQLFileWithTransaction(db *sql.DB, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("SQL file does not exist: %s", filePath)
	}

	// Read the SQL file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file %s: %w", filePath, err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	// Split the content by semicolons to handle multiple statements
	statements := strings.Split(string(content), ";")

	// Execute each statement within the transaction
	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("failed to execute statement %d in %s: %w\nStatement: %s",
				i+1, filepath.Base(filePath), err, statement)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func recoverOrderBook(c *container.Container) error {
	log.Infof("[RecoverOrderBook] start")
	markets := c.MatchingEngine.Markets()

	ctx := context.Background()

	for _, marketName := range markets {
		log.Infof("[RecoverOrderBook] trying to recover market: %s", marketName)
		openOrderStatuses := []model.OrderStatus{model.ORDER_STATUS_NEW, model.ORDER_STATUS_PARTIAL}
		orderDTOs, err := c.OrderService.QueryOrdersByMarketAndStatuses(ctx, marketName, openOrderStatuses)
		if len(orderDTOs) == 0 {
			log.Infof("[RecoverOrderBook] no order found in market: %s", marketName)
			continue
		}
		latestPrice, err := c.TradeRepo.GetMarketLatestPrice(ctx, c.DB, marketName)
		if err != nil {
			log.Warnf("[RecoverOrderBook] failed to get latest price for market: %s, using default 0.0", marketName)
			latestPrice = 0.0
		}

		orders := convertOrderDTOsToEngineOrders(orderDTOs)
		err = c.MatchingEngine.RecoverOrderBook(marketName, orders, latestPrice)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertOrderDTOsToEngineOrders(orderDTOs []*dto.Order) []*model.Order {
	orders := make([]*model.Order, 0, len(orderDTOs))
	for _, o := range orderDTOs {
		orders = append(orders, o.ToEngineOrder())
	}
	return orders
}

func startUpAllScheduler(c *container.Container) {
	err := c.OrderBookSnapshotScheduler.Start()
	if err != nil {
		panic(err)
	}

	err = c.MarketDataScheduler.Start()
	if err != nil {
		panic(err)
	}

	err = c.LQDTScheduler.Start()
	if err != nil {
		panic(err)
	}

	err = c.WSDataFeederScheduler.Start()
	if err != nil {
		panic(err)
	}
}

func setupWebSocket(c *container.Container) {
	go func() {
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws.HandleWebSocket(c.WSHub, w, r)
		})

		log.Info("WebSocket listen on :8081")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()
}

func getEthClient() (*ethclient.Client, error) {
	return ethclient.Dial("http://localhost:8545")
}
