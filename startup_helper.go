package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/container"
	"github.com/johnny1110/crypto-exchange/controller"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/middleware"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func setupRouter(c *container.Container) *gin.Engine {
	router := gin.Default()

	// add middleware
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware())

	// create controller
	userController := controller.NewUserController(c.UserService)
	balanceController := controller.NewBalanceController(c.BalanceService)
	orderController := controller.NewOrderController(c.OrderService)
	adminController := controller.NewAdminController(c.AdminService)
	orderBookService := controller.NewOrderBookController(c.OrderBookService)

	// setup routes
	setupRoutes(router, c, userController, balanceController, orderController, adminController, orderBookService)

	return router
}

func setupRoutes(
	router *gin.Engine,
	c *container.Container,
	userController *controller.UserController,
	balanceController *controller.BalanceController,
	orderController *controller.OrderController,
	adminController *controller.AdminController,
	orderBookService *controller.OrderBookController,
) {
	// Health check
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public router
	public := router.Group("/api/v1")
	{
		// user etc.
		public.POST("/users/register", userController.Register)
		public.POST("/users/login", userController.Login)
		public.GET("/orderbooks/:market/snapshot", orderBookService.OrderbooksSnapshot)
	}

	// Auth router
	private := router.Group("/api/v1")
	private.Use(middleware.AuthMiddleware(c.CredentialCache))
	{
		// users
		private.GET("/users/profile", userController.GetProfile)
		private.POST("/users/logout", userController.Logout)
		// balances
		private.GET("/balances", balanceController.GetBalances)
		// orders
		private.POST("/orders/:market", orderController.PlaceOrder)
		private.DELETE("/orders/:market/:orderId", orderController.CancelOrder)

	}

	// Admin router
	admin := router.Group("/admin/api/v1")
	admin.Use(middleware.AdminMiddleware())
	{
		admin.POST("/manual-adjustment", adminController.ManualAdjustment)
		admin.POST("/test-make-market", adminController.TestMakeMarket)
	}
}

// initDB if testMode = true, everytime startup the app, it will rebuild database with schema and prepare mock data.
func initDB(testMode bool) (*sql.DB, error) {
	// for windows
	db, err := sql.Open("sqlite", "file:exg.db")
	// for Mac
	//db, err := sql.Open("sqlite3", "./exg.db")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Infof("Database initialized successfully")

	// Run SQL files on startup if testMode
	if testMode {
		if err := runSQLFilesWithTransaction(db); err != nil {
			return nil, fmt.Errorf("failed to run SQL files: %w", err)
		}
		log.Infof("DB schema and testing data initialized successfully")
	}

	return db, err
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

// initMarkets define markets here.
func initMarkets() []*market.MarketInfo {
	btcMarket := market.NewMarketInfo("BTC-USDT", "BTC", "USDT")
	ethMarket := market.NewMarketInfo("ETH-USDT", "ETH", "USDT")
	dotMarket := market.NewMarketInfo("DOT-USDT", "DOT", "USDT")

	return []*market.MarketInfo{btcMarket, ethMarket, dotMarket}
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
			log.Errorf("[RecoverOrderBook] failed to get latest price for market: %s", marketName)
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

func getEthClient() (*ethclient.Client, error) {
	return ethclient.Dial("http://localhost:8545")
}
