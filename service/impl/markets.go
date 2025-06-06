package serviceImpl

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"time"
)

type MarketDataService struct {
	db        *sql.DB
	tradeRepo repository.ITradeRepository
	cache     service.ICacheService
}

func (d *MarketDataService) GetAllMarketData() ([]dto.MarketData, error) {
	allMarketInfos := settings.ALL_MARKETS
	allMarketData := make([]dto.MarketData, 0, len(allMarketInfos))

	for _, marketInfo := range allMarketInfos {
		val, found := d.cache.Get(settings.MARKET_DATA_CACHE.Apply(marketInfo.Name))
		if found {
			if marketData, ok := val.(*dto.MarketData); ok {
				allMarketData = append(allMarketData, *marketData)
			}
		}
	}

	return allMarketData, nil
}

func (d *MarketDataService) GetMarketData(market string) (dto.MarketData, error) {
	val, found := d.cache.Get(settings.MARKET_DATA_CACHE.Apply(market))
	if found {
		if marketData, ok := val.(*dto.MarketData); ok {
			return *marketData, nil
		} else {
			log.Errorf("[GetMarketData] Error market: %s", market)
		}
	}
	return dto.MarketData{}, fmt.Errorf("market %s data not found", market)
}

func NewMarketDataService(db *sql.DB, tradeRepo repository.ITradeRepository, cache service.ICacheService) service.IMarketDataService {
	return &MarketDataService{db: db, tradeRepo: tradeRepo, cache: cache}
}

func (d *MarketDataService) CalculateMarketData(ctx context.Context, market string) (*dto.MarketData, error) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	latestPrice, _ := d.getLatestPrice(ctx, market)
	price24hAgo, _ := d.getPrice24HoursAgo(ctx, market, yesterday)
	volume24h, _ := d.getVolume24Hours(ctx, market, yesterday, now)

	// calculate changing
	priceChange := 0.0
	if price24hAgo == 0 {
		price24hAgo = latestPrice
	} else {
		priceChange = (latestPrice - price24hAgo) / price24hAgo
	}

	return &dto.MarketData{
		MarketName:     market,
		LatestPrice:    latestPrice,
		PriceChange24H: priceChange,
		TotalVolume24H: volume24h,
	}, nil
}

func (d *MarketDataService) getLatestPrice(ctx context.Context, market string) (float64, error) {
	return d.tradeRepo.GetMarketLatestPrice(ctx, d.db, market)
}

func (d *MarketDataService) getPrice24HoursAgo(ctx context.Context, market string, yesterday time.Time) (float64, error) {
	return d.tradeRepo.GetMarketPriceTimesAgo(ctx, d.db, market, yesterday)
}

func (d *MarketDataService) getVolume24Hours(ctx context.Context, market string, yesterday time.Time, now time.Time) (float64, error) {
	return d.tradeRepo.GetMarketVolumeByTimeRange(ctx, d.db, market, yesterday, now)
}
