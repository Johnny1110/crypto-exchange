package scheduler

import (
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

type orderBookSnapshotScheduler struct {
	engine   *core.MatchingEngine
	duration time.Duration

	runTimes int64
	mu       sync.RWMutex //RW mutex
}

func NewOrderBookSnapshotScheduler(engine *core.MatchingEngine, duration time.Duration) Scheduler {
	return &orderBookSnapshotScheduler{
		engine:   engine,
		duration: duration,
	}
}

func (o *orderBookSnapshotScheduler) Name() string {
	return "orderBookSnapshot"
}

func (o *orderBookSnapshotScheduler) RunTimes() int64 {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.runTimes
}

func (o *orderBookSnapshotScheduler) countRunTime() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.runTimes += 1
}

func (o *orderBookSnapshotScheduler) Start() error {
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
