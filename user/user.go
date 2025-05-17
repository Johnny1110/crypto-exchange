package user

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
)

type User struct {
	Username   string
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(username, hotWalletPrivateKey string) *User {
	edcdsaPkey, err := crypto.HexToECDSA(hotWalletPrivateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		Username:   username,
		PrivateKey: edcdsaPkey,
	}
}
