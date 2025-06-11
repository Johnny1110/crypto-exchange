package metrics

import (
	"context"
	"github.com/johnny1110/crypto-exchange/scheduler"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics define
var (
	bidTotalVolume = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "exchange_bid_total_volume",
			Help: "Total volume of bid orders by market",
		},
		[]string{"market"},
	)

	askTotalVolume = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "exchange_ask_total_volume",
			Help: "Total volume of ask orders by market",
		},
		[]string{"market"},
	)

	latestDealtPrice = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "exchange_market_latest_price",
			Help: "latest dealt price by market",
		},
		[]string{"market"},
	)

	openOrdersCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "exchange_open_orders_count",
			Help: "Number of open orders by market",
		},
		[]string{"market"},
	)

	schedulerExecTimes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "exchange_scheduler_exec_times",
			Help: "Number of scheduler execute times by job name",
		},
		[]string{"jobName"},
	)
)

func init() {
	// register metrics
	prometheus.MustRegister(bidTotalVolume)
	prometheus.MustRegister(askTotalVolume)
	prometheus.MustRegister(latestDealtPrice)
	prometheus.MustRegister(openOrdersCount)
	prometheus.MustRegister(schedulerExecTimes)
}

type MetricService struct {
	bookService       service.IOrderBookService
	orderService      service.IOrderService
	schedulerReporter *scheduler.SchedulerReporter
}

func NewMetricService(bookService service.IOrderBookService,
	orderService service.IOrderService,
	schedulerReporter *scheduler.SchedulerReporter) *MetricService {
	return &MetricService{
		bookService:       bookService,
		orderService:      orderService,
		schedulerReporter: schedulerReporter,
	}
}

func (m *MetricService) UpdateMetrics(ctx context.Context) {
	m.updateOrderBookMetrics(ctx)
	m.updateOpenOrdersMetrics(ctx)
	m.updateSchedulerExecTimesMetrics(ctx)
}

func (m *MetricService) updateOrderBookMetrics(ctx context.Context) {
	for _, market := range settings.ALL_MARKETS {
		snapshot, err := m.bookService.GetSnapshot(ctx, market.Name)
		if err != nil {
			log.Errorf("[Metrics] updateOrderBookMetrics error: %v", err)
			continue
		}
		bidTotalVolume.WithLabelValues(market.Name).Set(snapshot.TotalBidSize)
		askTotalVolume.WithLabelValues(market.Name).Set(snapshot.TotalAskSize)
		latestDealtPrice.WithLabelValues(market.Name).Set(snapshot.LatestPrice)
	}
}

func (m *MetricService) updateOpenOrdersMetrics(ctx context.Context) {
	for _, market := range settings.ALL_MARKETS {
		count, err := m.orderService.CountOpenOrders(ctx, market.Name)
		if err != nil {
			log.Errorf("[Metrics] updateOpenOrdersMetrics error: %v", err)
			continue
		}
		openOrdersCount.WithLabelValues(market.Name).Set(float64(count))
	}
}

func (m *MetricService) updateSchedulerExecTimesMetrics(ctx context.Context) {
	reports := m.schedulerReporter.Report()
	for _, report := range reports {
		schedulerExecTimes.WithLabelValues(report.JobName).Set(float64(report.Times))
	}
}
