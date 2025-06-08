package dto

type Balance struct {
	Asset     string  `json:"asset"`
	Available float64 `json:"available"`
	Locked    float64 `json:"locked"`
	Total     float64 `json:"total"`

	// for API
	AssetValuation    float64 `json:"asset_valuation"`    // size * latestPrice
	ValuationCurrency string  `json:"valuation_currency"` //xxx USDT
}
