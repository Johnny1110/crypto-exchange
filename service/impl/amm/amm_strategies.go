package amm

import (
	"context"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
)

type AutoMarketStrategy interface {
	MakeMarket(ctx context.Context, market market.MarketInfo, maxQuoteAmtPerLevel float64)
}

type Strategy int

const (
	PROVIDE_LIQUIDITY = iota
	GRID_TRADING
	ARBITRAGE
)

func GetStrategy(s Strategy, proxy IAmmExchangeFuncProxy, ammUser dto.User) (AutoMarketStrategy, error) {
	switch s {
	case PROVIDE_LIQUIDITY:
		return NewProvideLiquidityStrategy(proxy, ammUser), nil
	default:
		return nil, fmt.Errorf("unknown strategy %d", s)
	}
}
