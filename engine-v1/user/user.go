package user

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"math/rand"
)

type User struct {
	UserID     int64
	Username   string
	Address    string
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(username, address, hotWalletPrivateKey string) *User {
	edcdsaPkey, err := crypto.HexToECDSA(hotWalletPrivateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		UserID:     int64(rand.Intn(100000000)),
		Username:   username,
		Address:    address,
		PrivateKey: edcdsaPkey,
	}
}
