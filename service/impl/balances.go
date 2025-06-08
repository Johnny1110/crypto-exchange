package serviceImpl

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/labstack/gommon/log"
)

type balanceService struct {
	db                *sql.DB
	userRepo          repository.IUserRepository
	balanceRepo       repository.IBalanceRepository
	marketDataService service.IMarketDataService
}

func NewIBalanceService(db *sql.DB,
	userRepo repository.IUserRepository,
	balanceRepo repository.IBalanceRepository,
	marketDataService service.IMarketDataService) service.IBalanceService {
	return &balanceService{
		db:                db,
		userRepo:          userRepo,
		balanceRepo:       balanceRepo,
		marketDataService: marketDataService,
	}
}

func (bs *balanceService) GetBalances(ctx context.Context, userId string) ([]*dto.Balance, error) {
	balances, err := bs.balanceRepo.GetBalancesByUserId(ctx, bs.db, userId)

	if err != nil {
		return nil, err
	}

	for _, balance := range balances {
		balance.ValuationCurrency = "USDT"

		if balance.Asset == "USDT" {
			balance.AssetValuation = balance.Total
			continue
		}

		if balance.Total > 0 {
			data, err := bs.marketDataService.GetMarketData(fmt.Sprintf("%v-USDT", balance.Asset))
			if err != nil {
				log.Warnf("Get balances market data err: %v", err)
			}
			balance.AssetValuation = data.LatestPrice * balance.Total
		} else {
			balance.AssetValuation = 0.0
		}
	}

	return balances, nil
}
