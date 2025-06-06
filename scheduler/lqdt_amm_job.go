package scheduler

import (
	"context"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/service/impl/amm"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"time"
)

type LQDTScheduler struct {
	ammExgFuncProxy amm.IAmmExchangeFuncProxy
	duration        time.Duration
	ammUser         dto.User
}

func NewLQDTScheduler(ammExgFuncProxy amm.IAmmExchangeFuncProxy, duration time.Duration) Scheduler {
	return &LQDTScheduler{
		ammExgFuncProxy: ammExgFuncProxy,
		duration:        duration,
		// TODO: remove this
		ammUser: dto.User{ID: "MID250606CXAZ1199", Username: "market_maker", MakerFee: 0.0001, TakerFee: 0.002},
	}
}

func (L LQDTScheduler) Start() error {
	ticker := time.NewTicker(L.duration)
	log.Info("[LQDTScheduler] start")

	ctx := context.Background()
	stg, _ := amm.GetStrategy(amm.PROVIDE_LIQUIDITY, L.ammExgFuncProxy, L.ammUser)

	go func() {
		for range ticker.C {
			for _, marketInfo := range settings.ALL_MARKETS {
				stg.MakeMarket(ctx, *marketInfo)
			}
		}
	}()

	return nil
}

func (L LQDTScheduler) Stop() error {
	//TODO implement me
	panic("implement me")
}
