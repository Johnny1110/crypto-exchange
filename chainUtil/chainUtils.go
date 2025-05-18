package chainUtil

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math"
	"math/big"
)

// standard ERC-20 ABI snippet with only balanceOf
const erc20ABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`

// tokenAddresses maps symbol to its mainnet contract address
var tokenAddresses = map[string]string{
	"USDT": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
	// add more tokens here as needed...
}

func TransferToken(client *ethclient.Client, symbol string, amount float64, to string, privateKey ecdsa.PrivateKey) {
	if !checkSymbol(symbol) {
		panic("symbol check failed")
	}

	ctx := context.Background()
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := common.HexToAddress(to)

	switch symbol {
	case "ETH":
		transferETH(client, fromAddress, toAddress, &privateKey, amount, &ctx)
		break
	default:
		transferERC20(client, fromAddress, toAddress, &privateKey, amount, &ctx)
	}

}

func transferETH(client *ethclient.Client, fromAddress, toAddress common.Address, privateKey *ecdsa.PrivateKey, amount float64, ctx *context.Context) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}

	// Convert amount from ETH to Wei
	ethToWei := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(math.Pow10(18)))
	value := new(big.Int)
	ethToWei.Int(value) // convert *big.Float to *big.Int (truncating)

	// Set gas parameters
	gasLimit := uint64(21000) // standard for ETH transfer
	gasPrice, err := client.SuggestGasPrice(*ctx)
	if err != nil {
		panic(err)
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	if err != nil {
		panic(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		panic(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("tx sent: %s \n", signedTx.Hash().Hex())

	fromBalanceAfter, _ := client.BalanceAt(*ctx, fromAddress, nil)
	toBalanceAfter, _ := client.BalanceAt(*ctx, toAddress, nil)
	fmt.Println("From Address After Txn Balance:", fromBalanceAfter)
	fmt.Println("To Address After Txn Balance:", toBalanceAfter)
	// TODO: seems like toWalletAddress didn't received the ETH, idk why, maybe let find it out next week.
}

func transferERC20(client *ethclient.Client, address common.Address, address2 common.Address, e *ecdsa.PrivateKey, amount float64, ctx *context.Context) {
	// TODO
}

// CheckBalance return address balance is greater than amount
func CheckBalance(client *ethclient.Client, symbol string, addressStr string, amount float64) bool {
	if !checkSymbol(symbol) {
		panic("symbol check failed")
	}

	ctx := context.Background()
	address := common.HexToAddress(addressStr)

	switch symbol {
	case "ETH":
		// native ETH balance at latest block
		rawBalance, err := client.BalanceAt(ctx, address, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println("Address: %s, ETH Balance: %s", address.Hex(), rawBalance)

		// convert Wei -> Ether
		balFloat := new(big.Float).Quo(
			new(big.Float).SetInt(rawBalance),
			big.NewFloat(math.Pow10(18)),
		)
		return balFloat.Cmp(big.NewFloat(amount)) >= 0

	default:
		// TODO: implement check ERC-20 token balance
		return true
	}
}

func QueryBalance(client *ethclient.Client, symbol string, addressStr string) (float64, error) {
	if !checkSymbol(symbol) {
		panic("symbol check failed")
	}

	ctx := context.Background()
	address := common.HexToAddress(addressStr)

	switch symbol {
	case "ETH":
		rawBalance, err := client.BalanceAt(ctx, address, nil)
		if err != nil {
			return 0, err
		}
		// convert Wei -> Ether
		balFloat := new(big.Float).Quo(
			new(big.Float).SetInt(rawBalance),
			big.NewFloat(math.Pow10(18)),
		)

		f64, _ := balFloat.Float64() // convert *big.Float to float64 (lossy but fine for display)
		return f64, nil

	default:
		// TODO: Add ERC-20 support
		return 0, errors.New("token not supported yet")
	}
}

// checkSymbol Only support ETH USDT now.
func checkSymbol(symbol string) bool {
	return symbol == "ETH" || symbol == "USDT"
}
