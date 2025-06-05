package main

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/container"
	"github.com/johnny1110/crypto-exchange/controller"
	"github.com/johnny1110/crypto-exchange/middleware"
	"net/http"
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
		private.DELETE("/orders/:orderId", orderController.CancelOrder)
		private.GET("/orders", orderController.GetOrders)

	}

	// Admin router
	admin := router.Group("/admin/api/v1")
	admin.Use(middleware.AdminMiddleware())
	{
		admin.POST("/manual-adjustment", adminController.ManualAdjustment)
		admin.POST("/test-make-market", adminController.TestMakeMarket)
	}
}
