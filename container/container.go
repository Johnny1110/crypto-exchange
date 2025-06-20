package container

import (
	"context"
	"database/sql"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/ohlcv"
	"github.com/johnny1110/crypto-exchange/repository"
	repositoryImpl "github.com/johnny1110/crypto-exchange/repository/impl"
	"github.com/johnny1110/crypto-exchange/scheduler"
	"github.com/johnny1110/crypto-exchange/security"
	"github.com/johnny1110/crypto-exchange/service"
	serviceImpl "github.com/johnny1110/crypto-exchange/service/impl"
	"github.com/johnny1110/crypto-exchange/service/impl/amm"
	"github.com/johnny1110/crypto-exchange/service/impl/metrics"
	"github.com/johnny1110/crypto-exchange/settings"
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
	OHLCVRepo   ohlcv.OHLCVRepository

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
	SchedulerReporter          *scheduler.SchedulerReporter
	MarketDataScheduler        scheduler.Scheduler
	OrderBookSnapshotScheduler scheduler.Scheduler
	LQDTScheduler              scheduler.Scheduler

	// Metrics
	MetricsService *metrics.MetricService

	// Proxy
	AmmExFuncProxy amm.IAmmExchangeFuncProxy

	// OHLCV
	OHLCVAggregator       *ohlcv.OHLCVAggregator
	OHLCVTradeStream      ohlcv.TradeStream
	OHLCVAggregatorConfig *ohlcv.AggregatorConfig
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

	// init kline module
	c.initOHLCVAgg()

	// init services
	c.initServices()

	// init proxy()
	c.initProxy()

	// init Scheduler
	c.initScheduler()

	// init Metrics
	c.initMetrics()

	return c
}

func (c *Container) initRepositories() {
	c.UserRepo = repositoryImpl.NewUserRepository()
	c.BalanceRepo = repositoryImpl.NewBalanceRepository()
	c.OrderRepo = repositoryImpl.NewOrderRepository()
	c.TradeRepo = repositoryImpl.NewTradeRepository()
	c.OHLCVRepo = ohlcv.NewSQLiteOHLCVRepository(c.DB)
}

func (c *Container) initServices() {
	c.UserService = serviceImpl.NewIUserService(c.DB, c.UserRepo, c.BalanceRepo, c.CredentialCache)
	c.OrderService = serviceImpl.NewIOrderService(c.DB, c.MatchingEngine, c.OrderRepo, c.TradeRepo, c.BalanceRepo, c.OHLCVTradeStream)
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
	c.LQDTScheduler = scheduler.NewLQDTScheduler(c.AmmExFuncProxy, c.UserService, 5*time.Minute)

	schedulers := make([]scheduler.Scheduler, 0, 3)
	schedulers = append(schedulers, c.MarketDataScheduler)
	schedulers = append(schedulers, c.OrderBookSnapshotScheduler)
	schedulers = append(schedulers, c.LQDTScheduler)

	c.SchedulerReporter = scheduler.NewSchedulerReporter(schedulers)
}

func (c *Container) initProxy() {
	c.AmmExFuncProxy = amm.NewAmmExchangeFuncProxyImpl(
		c.OrderBookService, c.BalanceService, c.OrderService, c.UserService,
		&http.Client{
			Timeout: 30 * time.Second,
		})
}

func (c *Container) initMetrics() {
	c.MetricsService = metrics.NewMetricService(c.OrderBookService, c.OrderService, c.SchedulerReporter)
}

func (c *Container) initOHLCVAgg() {
	c.OHLCVTradeStream = ohlcv.NewSimpleTradeStream(1000)

	c.OHLCVAggregatorConfig = &ohlcv.AggregatorConfig{
		BatchSize:      100,
		FlushInterval:  1 * time.Second,
		ChannelSize:    1000,
		MaxConcurrency: 10,
		EnableMetrics:  false,
	}
	if agg, err := ohlcv.NewOHLCVAggregator(c.OHLCVRepo, c.OHLCVTradeStream, c.OHLCVAggregatorConfig); err != nil {
		log.Fatalf("[OHLCVAggregator] initOHLCVAgg faialed, err: %v", err)
	} else {
		c.OHLCVAggregator = agg
	}

	ctx := context.Background()

	allSymbolNames := make([]string, 0, len(settings.ALL_MARKETS))
	for _, symbol := range settings.ALL_MARKETS {
		initPrice := 0.001
		if ob, err := c.MatchingEngine.GetOrderBook(symbol.Name); err == nil {
			initPrice = max(ob.LatestPrice(), initPrice)
		}

		err := c.OHLCVAggregator.AddSymbol(
			symbol.Name,
			initPrice,
			ohlcv.SupportedIntervals,
		)
		if err != nil {
			log.Fatalf("[OHLCVAggregator] initOHLCVAgg add symbol err: %v", err)
			return
		}

		allSymbolNames = append(allSymbolNames, symbol.Name)
	}

	err := c.OHLCVAggregator.Start(ctx, allSymbolNames)
	if err != nil {
		log.Fatalf("[OHLCVAggregator] initOHLCVAgg start err: %v", err)
		return
	}
}
