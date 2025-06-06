package scheduler

import (
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/labstack/gommon/log"
	"time"
)

type orderBookSnapshotScheduler struct {
	engine   *core.MatchingEngine
	duration time.Duration
}

func NewOrderBookSnapshotScheduler(engine *core.MatchingEngine, duration time.Duration) Scheduler {
	return &orderBookSnapshotScheduler{
		engine:   engine,
		duration: duration,
	}
}

func (o orderBookSnapshotScheduler) Start() error {
	ticker := time.NewTicker(o.duration)
	log.Info("[OrderBookSnapshotScheduler] start")

	go func() {
		for range ticker.C {
			for _, market := range o.engine.Markets() {
				ob, err := o.engine.GetOrderBook(market)
				if err != nil {
					log.Errorf("[BookSnapshotScheduler] StartSnapshotRefresher: GetOrderBook err: %v", err)
				} else {
					ob.RefreshSnapshot()
				}
			}
		}
	}()

	return nil
}

func (o orderBookSnapshotScheduler) Stop() error {
	return nil
}
