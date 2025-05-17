package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"net/http"

	"github.com/johnny1110/crypto-exchange/exchange"
	"github.com/labstack/echo/v4"
)

const (
	hotWalletAddress    = "0x775066a589F42731f097e2bc6165677A65466038"
	hotWalletPrivateKey = "285c16b558b840cb3f6de3701f2e7d75708910c617f335c7454ac5fece5cd499"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	ex, err := exchange.NewExchange(hotWalletAddress, hotWalletPrivateKey)

	if err != nil {
		log.Fatal(err)
	}

	ex.InitOrderbooks()
	e.GET("/healthcheck", func(c echo.Context) error { return c.String(http.StatusOK, "OK") })
	e.POST("/order", ex.HandlePlaceOrder)
	e.DELETE("/orderbook/:market/order/:id", ex.HandleDeleteOrder)
	e.GET("/orderbook/:market", ex.HandleGetOrderBook)
	e.GET("/orderbook/:market/orderIds", ex.HandleGetOrderIds)

	//test_gnache()

	e.Start(":3000")
}

func test_gnache() {
	// testing ganache with a hot wallet balance check.
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to the Ethereum client:", err)
	}

	ctx := context.Background()
	address := common.HexToAddress("0x775066a589F42731f097e2bc6165677A65466038")
	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}

	fmt.Println("Address:", address.Hex())
	// balance divide 18
	fmt.Println("Balance:", balance)

	privateKey, err := crypto.HexToECDSA("285c16b558b840cb3f6de3701f2e7d75708910c617f335c7454ac5fece5cd499")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1000000000000000000) // in wei (1 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xE0d9b33e91d8c17d2B41ffBe6567c997747a816F")
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s \n", signedTx.Hash().Hex())

	balanceAfter, _ := client.BalanceAt(ctx, address, nil)
	fmt.Println("After Txn Balance:", balanceAfter)
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}
