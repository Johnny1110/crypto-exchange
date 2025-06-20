package scheduler

import (
	"context"
	"github.com/johnny1110/crypto-exchange/ohlcv"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/ws"
	"github.com/labstack/gommon/log"
	"time"
)

type WSDataFeederJob struct {
	wsHub            *ws.Hub
	ohlcvAgg         *ohlcv.OHLCVAggregator
	orderbookService service.IOrderBookService
}

func NewWSDataFeederJob(
	wsHub *ws.Hub,
	ohlcvAgg *ohlcv.OHLCVAggregator,
	orderbookService service.IOrderBookService) Scheduler {
	return &WSDataFeederJob{
		wsHub:            wsHub,
		ohlcvAgg:         ohlcvAgg,
		orderbookService: orderbookService,
	}
}

func (W *WSDataFeederJob) Start() error {
	log.Infof("[WSDataFeederJob] start to feeding data to ws.")
	go func() {

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				keys := W.wsHub.GetSubscriptionKeys()
				for _, key := range keys {
					go W.collectAndSend(key)
				}
			}
		}

	}()

	return nil
}

func (W *WSDataFeederJob) collectAndSend(key ws.SubscriptionKey) {
	ctx := context.Background()

	switch key.Channel {
	case ws.OHLCV:
		ohlcvWsParam := key.Params.(*ws.OHLCVReqParams)
		symbol := ohlcvWsParam.Symbol
		interval := ohlcv.OHLCV_INTERVAL(ohlcvWsParam.Interval)

		realtimeOHLCV, err := W.ohlcvAgg.GetRealtimeOHLCV(ctx, symbol, interval)
		if err != nil {
			log.Warnf("[WSDataFeederJob] get realtime ohlcv error: %v", err)
			return
		}
		W.wsHub.BroadcastToSubscribers(key, realtimeOHLCV)
		return
	case ws.ORDERBOOK:
		obWsParam := key.Params.(*ws.OrderBookReqParams)
		market := obWsParam.Market
		snapshot, err := W.orderbookService.GetSnapshot(ctx, market)
		if err != nil {
			log.Warnf("[WSDataFeederJob] get orderbook snapshot error: %v", err)
			return
		}
		W.wsHub.BroadcastToSubscribers(key, snapshot)
		return
	default:
		return
	}
}

func (W *WSDataFeederJob) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (W *WSDataFeederJob) Name() string {
	return "WSDataFeederJob"
}

func (W *WSDataFeederJob) RunTimes() int64 {
	return 0
}
