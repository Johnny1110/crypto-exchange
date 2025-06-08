package settings

import "github.com/johnny1110/crypto-exchange/engine-v2/market"

func GetAllAssets() []string {
	return []string{"USDT", "BTC", "ETH", "DOT", "ASTR", "HDX"}
}

var ALL_MARKETS = []*market.MarketInfo{
	{Name: "BTC-USDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	{Name: "ETH-USDT", BaseAsset: "ETH", QuoteAsset: "USDT"},
	{Name: "DOT-USDT", BaseAsset: "DOT", QuoteAsset: "USDT"},
}

const MARGIN_ACCOUNT_ID = "0"
const INTERNAL_AMM_ACCOUNT_ID = "MID250606CXAZ1199"
