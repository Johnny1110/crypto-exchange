package scheduler

import (
	"context"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"time"
)

type MarketDataScheduler struct {
	dataService  service.IMarketDataService
	cacheService service.ICacheService
	ticker       *time.Ticker
	stopCh       chan struct{}
	markets      []string
	duration     time.Duration
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
	}
}

func (s *MarketDataScheduler) Start() error {
	log.Printf("[MarketDataScheduler] Starting scheduler for markets: %v", s.markets)
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
	log.Info("[MarketDataScheduler] Updating market data...")

	for _, market := range s.markets {
		marketData, err := s.dataService.CalculateMarketData(ctx, market)
		if err != nil {
			log.Printf("Error calculating data for market %s: %v", market, err)
			continue
		}
		cacheKey := settings.MARKET_DATA_CACHE.Apply(market)
		s.cacheService.Update(cacheKey, marketData)
		log.Printf("Updated data for market: %s, price: %.4f, change: %.4f, volume: %.2f",
			market, marketData.LatestPrice, marketData.PriceChange24H, marketData.TotalVolume24H)
	}
}
