package dto

type MarketData struct {
	MarketName     string  `json:"market_name"`
	LatestPrice    float64 `json:"latest_price"`
	PriceChange24H float64 `json:"price_change_24h"`
	TotalVolume24H float64 `json:"total_volume_24h"`
}
