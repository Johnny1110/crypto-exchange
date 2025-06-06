package amm

import (
	"context"
	"database/sql"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/labstack/gommon/log"
)

// Auto Market Maker (AMM)

type autoMarketMakerService struct {
	db                *sql.DB
	orderService      service.IOrderService
	balanceService    service.IBalanceService
	orderBookService  service.IOrderBookService
	priceIndexService service.IPriceIndexService
}

func NewAutoMarketMakerService() service.IAutoMarketMakerService {
	return &autoMarketMakerService{}
}

func (a autoMarketMakerService) BootUp(ctx context.Context, markets []market.MarketInfo) {

	for _, m := range markets {
		log.Infof("[AMM] Booting up market %s", m.Name)
	}

	//TODO implement me
	panic("implement me")
}
