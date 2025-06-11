package scheduler

import (
	"context"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

type MarketDataScheduler struct {
	dataService  service.IMarketDataService
	cacheService service.ICacheService
	ticker       *time.Ticker
	stopCh       chan struct{}
	markets      []string
	duration     time.Duration

	runTimes int64
	mu       sync.RWMutex //RW mutex
}

func NewMarketDataScheduler(dataService service.IMarketDataService, cache service.ICacheService, duration time.Duration) Scheduler {
	markets := make([]string, 0, len(settings.ALL_MARKETS))
	for _, info := range settings.ALL_MARKETS {
		markets = append(markets, info.Name)
	}

	return &MarketDataScheduler{
		markets:      markets,
		dataService:  dataService,
		cacheService: cache,
		stopCh:       make(chan struct{}),
		duration:     duration,

		runTimes: 0,
	}
}

func (s *MarketDataScheduler) Name() string {
	return "MarketDataScheduler"
}

func (s *MarketDataScheduler) RunTimes() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	time := s.runTimes
	log.Debugf("[MarketDataScheduler] return run times: %d", time)
	return time
}

func (s *MarketDataScheduler) countRunTime() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runTimes += 1
	log.Debugf("[MarketDataScheduler] run time count: %d]", s.runTimes)
}

func (s *MarketDataScheduler) Start() error {
	log.Debugf("[MarketDataScheduler] Starting scheduler for markets: %v", s.markets)
	ctx := context.Background()
	s.updateMarketData(ctx)

	s.ticker = time.NewTicker(s.duration)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.updateMarketData(ctx)
			case <-s.stopCh:
				return
			}
		}
	}()

	return nil
}

func (s *MarketDataScheduler) Stop() error {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)
	log.Info("[MarketDataScheduler] stopped")
	return nil
}

func (s *MarketDataScheduler) updateMarketData(ctx context.Context) {
	log.Debugf("[MarketDataScheduler] Updating market data...")
	s.countRunTime()

	for _, market := range s.markets {
		marketData, err := s.dataService.CalculateMarketData(ctx, market)
		if err != nil {
			log.Printf("Error calculating data for market %s: %v", market, err)
			continue
		}
		cacheKey := settings.MARKET_DATA_CACHE.Apply(market)
		s.cacheService.Update(cacheKey, marketData)
		log.Debugf("Updated data for market: %s, price: %.4f, change: %.4f, volume: %.2f",
			market, marketData.LatestPrice, marketData.PriceChange24H, marketData.TotalVolume24H)
	}
}
