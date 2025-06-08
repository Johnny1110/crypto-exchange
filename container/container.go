package container

import (
	"database/sql"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/repository"
	repositoryImpl "github.com/johnny1110/crypto-exchange/repository/impl"
	"github.com/johnny1110/crypto-exchange/scheduler"
	"github.com/johnny1110/crypto-exchange/security"
	"github.com/johnny1110/crypto-exchange/service"
	serviceImpl "github.com/johnny1110/crypto-exchange/service/impl"
	"github.com/johnny1110/crypto-exchange/service/impl/amm"
	"log"
	"net/http"
	"time"
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
	UserService       service.IUserService
	BalanceService    service.IBalanceService
	OrderService      service.IOrderService
	OrderBookService  service.IOrderBookService
	AdminService      service.IAdminService
	CacheService      service.ICacheService
	MarketDataService service.IMarketDataService

	// Cache and Security
	CredentialCache *security.CredentialCache
	MatchingEngine  *core.MatchingEngine

	// Scheduler
	MarketDataScheduler        scheduler.Scheduler
	OrderBookSnapshotScheduler scheduler.Scheduler
	LQDTScheduler              scheduler.Scheduler

	// Proxy
	AmmExFuncProxy amm.IAmmExchangeFuncProxy
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

	// init proxy()
	c.initProxy()

	// init Scheduler
	c.initScheduler()

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
	c.OrderService = serviceImpl.NewIOrderService(c.DB, c.MatchingEngine, c.OrderRepo, c.TradeRepo, c.BalanceRepo)
	c.OrderBookService = serviceImpl.NewIOrderBookService(c.MatchingEngine)
	c.AdminService = serviceImpl.NewIAdminService(c.DB, c.UserRepo, c.BalanceRepo, c.OrderService)
	c.CacheService = serviceImpl.NewCacheService()
	c.MarketDataService = serviceImpl.NewMarketDataService(c.DB, c.TradeRepo, c.CacheService)
	c.BalanceService = serviceImpl.NewIBalanceService(c.DB, c.UserRepo, c.BalanceRepo, c.MarketDataService)
}

// Cleanup clean
func (c *Container) Cleanup() {
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

func (c *Container) initScheduler() {
	c.MarketDataScheduler = scheduler.NewMarketDataScheduler(c.MarketDataService, c.CacheService, 30*time.Second)
	c.OrderBookSnapshotScheduler = scheduler.NewOrderBookSnapshotScheduler(c.MatchingEngine, 300*time.Millisecond)
	c.LQDTScheduler = scheduler.NewLQDTScheduler(c.AmmExFuncProxy, c.UserService, 2*time.Minute)
}

func (c *Container) initProxy() {
	c.AmmExFuncProxy = amm.NewAmmExchangeFuncProxyImpl(
		c.OrderBookService, c.BalanceService, c.OrderService, c.UserService,
		&http.Client{
			Timeout: 30 * time.Second,
		})
}
