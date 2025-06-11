package scheduler

import (
	"context"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/service/impl/amm"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

type LQDTScheduler struct {
	ammExgFuncProxy amm.IAmmExchangeFuncProxy
	duration        time.Duration
	ammUser         dto.User

	runTimes int64
	mu       sync.RWMutex //RW mutex
}

func (L *LQDTScheduler) Name() string {
	return "AMM"
}

func (L *LQDTScheduler) RunTimes() int64 {
	L.mu.RLock()
	defer L.mu.RUnlock()

	return L.runTimes
}

func (L *LQDTScheduler) countRunTime() {
	L.mu.Lock()
	defer L.mu.Unlock()
	L.runTimes += 1
}

func NewLQDTScheduler(ammExgFuncProxy amm.IAmmExchangeFuncProxy, service service.IUserService, duration time.Duration) Scheduler {
	ammAccount, err := service.GetUser(context.Background(), settings.INTERNAL_AMM_ACCOUNT_ID)
	if err != nil {
		log.Fatalf("[NewLQDTScheduler] failed to gat AMM User Data: %v", err)
	}

	return &LQDTScheduler{
		ammExgFuncProxy: ammExgFuncProxy,
		duration:        duration,
		ammUser:         *ammAccount,
	}
}

func (L *LQDTScheduler) Start() error {
	ticker := time.NewTicker(L.duration)
	log.Info("[LQDTScheduler] start")

	ctx := context.Background()
	stg, _ := amm.GetStrategy(amm.PROVIDE_LIQUIDITY, L.ammExgFuncProxy, L.ammUser)

	go func() {
		for range ticker.C {
			L.countRunTime()
			for _, marketInfo := range settings.ALL_MARKETS {
				maxQuoteAmtPerLevel, ok := settings.MAX_QUOTE_AMT_PER_LEVEL_MAP[marketInfo.Name]
				if !ok {
					log.Warnf("[LQDTScheduler] no found maxQuoteAmtPerLevel param for market: %s, using default 1 USDT", marketInfo.Name)
					maxQuoteAmtPerLevel = 1
				}
				stg.MakeMarket(ctx, *marketInfo, maxQuoteAmtPerLevel)
			}
		}
	}()

	return nil
}

func (L *LQDTScheduler) Stop() error {
	//TODO implement me
	panic("implement me")
}
