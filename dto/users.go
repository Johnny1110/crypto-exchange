package dto

type User struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"-"`
	VipLevel     int     `json:"vip_level"`
	MakerFee     float64 `json:"maker_fee"`
	TakerFee     float64 `json:"taker_fee"`
}
