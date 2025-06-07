package test

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/service/impl/amm"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v, got %v", b, a)
	}
}

func mock_PLD_strategy() *amm.ProvideLiquidityStrategy {
	exFuncProxy := newMockExgFuncProxyImpl()
	s := &amm.ProvideLiquidityStrategy{
		ExchangeFuncProxy: exFuncProxy,
		AmmUID:            "U01",
		AmmUser:           mockAmmUser(),
	}
	return s
}

func Test_AMM_CalculateIdealPriceLevels_BID(t *testing.T) {
	stg := mock_PLD_strategy()
	b := amm.NewBalance("ETH", "USDT", 100, 0, 100000, 0)
	levels := stg.CalculateIdealPriceLevels(3000, model.BID, *b, 100)
	for i, l := range levels {
		fmt.Printf("Level %v, %v\n", i+1, l)
	}
	//assert(t, b, level)
}

func Test_AMM_CalculateIdealPriceLevels_ASK(t *testing.T) {
	stg := mock_PLD_strategy()
	b := amm.NewBalance("ETH", "USDT", 100, 0, 100000, 0)
	levels := stg.CalculateIdealPriceLevels(3000, model.ASK, *b, 100)
	for i, l := range levels {
		fmt.Printf("Level %v, %v\n", i+1, l)
	}
	//assert(t, b, level)
}
