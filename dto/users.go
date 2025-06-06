package dto

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	VipLevel     int       `json:"vip_level"`
	MakerFee     float64   `json:"maker_fee"`
	TakerFee     float64   `json:"taker_fee"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		CreatedAt int64 `json:"created_at"`
	}{
		Alias:     (*Alias)(&u),
		CreatedAt: u.CreatedAt.UnixMilli(),
	})
}
