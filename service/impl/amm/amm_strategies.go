package amm

import (
	"context"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
)

type AutoMarketStrategy interface {
	MakeMarket(ctx context.Context, market market.MarketInfo)
}
