package settings

import "github.com/johnny1110/crypto-exchange/engine-v2/market"

// All supported Tokens
func GetAllAssets() []string {
	return []string{"USDT", "BTC", "ETH", "DOT", "ASTR", "HDX", "BTSE", "SOL", "LINK", "ADA", "BNB", "AVAX", "DOGE"}
}

// All supported Markets
var ALL_MARKETS = []*market.MarketInfo{
	{Name: "BTC-USDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	{Name: "ETH-USDT", BaseAsset: "ETH", QuoteAsset: "USDT"},
	{Name: "DOT-USDT", BaseAsset: "DOT", QuoteAsset: "USDT"},
	{Name: "SOL-USDT", BaseAsset: "SOL", QuoteAsset: "USDT"},
	{Name: "LINK-USDT", BaseAsset: "LINK", QuoteAsset: "USDT"},
	{Name: "ADA-USDT", BaseAsset: "ADA", QuoteAsset: "USDT"},
	{Name: "BNB-USDT", BaseAsset: "BNB", QuoteAsset: "USDT"},
	{Name: "AVAX-USDT", BaseAsset: "AVAX", QuoteAsset: "USDT"},
	{Name: "DOGE-USDT", BaseAsset: "DOGE", QuoteAsset: "USDT"},
	{Name: "BTSE-USDT", BaseAsset: "BTSE", QuoteAsset: "USDT"},
}

// AMM Price Level settings
var MAX_QUOTE_AMT_PER_LEVEL_MAP = map[string]float64{
	"ETH-USDT":  4000,
	"BTC-USDT":  20000,
	"DOT-USDT":  1500,
	"SOL-USDT":  3888,
	"LINK-USDT": 1188,
	"ADA-USDT":  1188,
	"BNB-USDT":  1288,
	"AVAX-USDT": 1388,
	"DOGE-USDT": 1488,
	"BTSE-USDT": 1588,
}

const MARGIN_ACCOUNT_ID = "0"
const INTERNAL_AMM_ACCOUNT_ID = "MID250606CXAZ1199"
