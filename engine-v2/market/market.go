package market

import (
	"fmt"
)

type MarketInfo struct {
	Name       string // e.g."BTC/USDT"
	BaseAsset  string // e.g. "BTC"
	QuoteAsset string // e.g. "USDT"
}

func NewMarketInfo(name string, baseAsset, quoteAsset string) *MarketInfo {
	return &MarketInfo{
		Name:       name,
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
	}
}

type MarketManager struct {
	markets map[string]*MarketInfo
}

func NewManager(mkList []MarketInfo) *MarketManager {
	mgr := &MarketManager{
		markets: make(map[string]*MarketInfo),
	}
	for _, mi := range mkList {
		mgr.markets[mi.Name] = &mi
	}
	return mgr
}

// List returns the names of all registered market
func (mgr *MarketManager) List() []string {
	names := make([]string, 0, len(mgr.markets))
	for name := range mgr.markets {
		names = append(names, name)
	}
	return names
}

// Get retrieves MarketInfo by market name
func (mgr *MarketManager) Get(market string) (*MarketInfo, error) {
	if mi, ok := mgr.markets[market]; ok {
		return mi, nil
	}
	return nil, fmt.Errorf("market %s not found", market)
}

// GetAssets input market and return (base, quote) assets.html
func (mgr *MarketManager) GetAssets(market string) (string, string, error) {
	mi, err := mgr.Get(market)
	if err != nil {
		return "", "", err
	}
	return mi.BaseAsset, mi.QuoteAsset, nil
}
