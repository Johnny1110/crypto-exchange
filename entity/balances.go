package entity

type Balance struct {
	Asset     string  `json:"asset"`
	Available float64 `json:"available"`
	Locked    float64 `json:"locked"`
}
