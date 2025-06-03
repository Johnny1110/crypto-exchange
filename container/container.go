package container

import (
	"database/sql"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/repository"
	repositoryImpl "github.com/johnny1110/crypto-exchange/repository/impl"
	"github.com/johnny1110/crypto-exchange/security"
	"github.com/johnny1110/crypto-exchange/service"
	serviceImpl "github.com/johnny1110/crypto-exchange/service/impl"
	"log"
)

// Container including all service and repo
type Container struct {
	// Database
	DB *sql.DB

	// Repositories
	UserRepo    repository.IUserRepository
	BalanceRepo repository.IBalanceRepository
	OrderRepo   repository.IOrderRepository
	TradeRepo   repository.ITradeRepository

	// Services
	UserService      service.IUserService
	BalanceService   service.IBalanceService
	OrderService     service.IOrderService
	OrderBookService service.IOrderBookService
	AdminService     service.IAdminService

	// Cache and Security
	CredentialCache *security.CredentialCache
	MatchingEngine  *core.MatchingEngine
}

// NewContainer do DI
func NewContainer(db *sql.DB, engine *core.MatchingEngine) *Container {
	c := &Container{
		DB:             db,
		MatchingEngine: engine,
	}

	// init cache
	c.CredentialCache = security.NewCredentialCache()

	// init repositories
	c.initRepositories()

	// init services
	c.initServices()

	return c
}

func (c *Container) initRepositories() {
	c.UserRepo = repositoryImpl.NewUserRepository()
	c.BalanceRepo = repositoryImpl.NewBalanceRepository()
	c.OrderRepo = repositoryImpl.NewOrderRepository()
	c.TradeRepo = repositoryImpl.NewTradeRepository()
}

func (c *Container) initServices() {
	c.UserService = serviceImpl.NewIUserService(c.DB, c.UserRepo, c.BalanceRepo, c.CredentialCache)
	c.BalanceService = serviceImpl.NewIBalanceService(c.DB, c.UserRepo, c.BalanceRepo)
	c.OrderService = serviceImpl.NewIOrderService(c.DB, c.MatchingEngine, c.OrderRepo, c.TradeRepo, c.BalanceRepo)
	c.OrderBookService = serviceImpl.NewIOrderBookService(c.MatchingEngine)
	c.AdminService = serviceImpl.NewIAdminService(c.DB, c.UserRepo, c.BalanceRepo, c.OrderService)
}

// Cleanup 清理資源
func (c *Container) Cleanup() {
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}
