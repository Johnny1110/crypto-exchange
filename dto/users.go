package dto

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string
}
