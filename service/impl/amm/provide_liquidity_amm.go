package amm

import (
	"context"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/labstack/gommon/log"
	"math"
)

// provideLiquidityStrategy provide Liquidity in OrderBook
type ProvideLiquidityStrategy struct {
	ExchangeFuncProxy IAmmExchangeFuncProxy
	AmmUID            string
	AmmUser           dto.User
}

func NewProvideLiquidityStrategy(exchangeFuncProxy IAmmExchangeFuncProxy, ammUser dto.User) AutoMarketStrategy {
	return &ProvideLiquidityStrategy{
		ExchangeFuncProxy: exchangeFuncProxy,
		AmmUID:            ammUser.ID,
		AmmUser:           ammUser,
	}
}

var (
	separate        = 0.001 // 報價間距(Spread)±0.1%
	levelAmtPerSide = 20    // ask, bid 各維持 20 檔位價格
)

// PriceLevel 代表一個價格檔位
type PriceLevel struct {
	Price  float64
	Volume float64
}

// MakeMarket provideLiquidityStrategy can access ExchangeFuncProxy to make market
func (p *ProvideLiquidityStrategy) MakeMarket(ctx context.Context, marketInfo market.MarketInfo, maxQuoteAmtPerLevel float64) {
	marketName := marketInfo.Name

	// 1. 獲取指數價格作為中間價
	indexPrice, err := p.ExchangeFuncProxy.GetIndexPrice(ctx, marketName)
	if err != nil {
		log.Printf("[AMM] Failed to get index price for %s: %v", marketName, err)
		return
	}

	// 2. 獲取當前餘額
	balance, err := p.ExchangeFuncProxy.GetBalance(ctx, p.AmmUID, marketName)
	if err != nil {
		log.Printf("[AMM] Failed to get balance for %s: %v", marketName, err)
		return
	}

	// 3. 獲取當前開放訂單
	openOrders, err := p.ExchangeFuncProxy.GetOpenOrders(ctx, p.AmmUID, marketName)
	if err != nil {
		log.Printf("[AMM] Failed to get open orders for %s: %v", marketName, err)
		return
	}

	// 4. 計算理想的價格檔位
	idealBidLevels := p.CalculateIdealPriceLevels(indexPrice, model.BID, balance, maxQuoteAmtPerLevel)
	idealAskLevels := p.CalculateIdealPriceLevels(indexPrice, model.ASK, balance, maxQuoteAmtPerLevel)

	// 5. 分析現有訂單並進行調整
	p.AdjustOrders(ctx, marketName, openOrders, idealBidLevels, idealAskLevels)
}

// calculateIdealPriceLevels 計算理想的價格檔位
func (p *ProvideLiquidityStrategy) CalculateIdealPriceLevels(indexPrice float64, side model.Side, balance Balance, maxQuoteAmtPerLevel float64) []PriceLevel {
	levels := make([]PriceLevel, 0, levelAmtPerSide)

	for i := 1; i <= levelAmtPerSide; i++ {
		var price float64
		var maxVolume float64

		if side == model.BID {
			// Buy side: 價格遞減
			price = indexPrice * (1 - separate*float64(i))
			// 計算可用的 quote 資產能買多少 base 資產
			availableQuote := balance.quoteAvailable / float64(levelAmtPerSide)
			if availableQuote > maxQuoteAmtPerLevel {
				availableQuote = maxQuoteAmtPerLevel
			}
			maxVolume = availableQuote / price
		} else {
			// Sell side: 價格遞增
			price = indexPrice * (1 + separate*float64(i))
			// 計算可用的 base 資產
			maxVolume = balance.baseAvailable / float64(levelAmtPerSide)
			if maxVolume*price > maxQuoteAmtPerLevel {
				maxVolume = maxQuoteAmtPerLevel / price
			}
		}

		if maxVolume > 0 {
			levels = append(levels, PriceLevel{
				Price:  p.RoundPrice(price),
				Volume: p.RoundVolume(maxVolume),
			})
		}
	}

	return levels
}

// adjustOrders 調整訂單以符合理想檔位
func (p *ProvideLiquidityStrategy) AdjustOrders(ctx context.Context, marketName string,
	openOrders []*dto.Order, idealBidLevels, idealAskLevels []PriceLevel) {

	// 分類現有訂單
	existingBids := make(map[float64]*dto.Order)
	existingAsks := make(map[float64]*dto.Order)

	for _, order := range openOrders {
		if order.Status == model.ORDER_STATUS_NEW || order.Status == model.ORDER_STATUS_PARTIAL {
			if order.Side == model.BID {
				existingBids[order.Price] = order
			} else {
				existingAsks[order.Price] = order
			}
		}
	}

	// 調整 Bid 訂單
	p.AdjustOrdersForSide(ctx, marketName, existingBids, idealBidLevels, model.BID)

	// 調整 Ask 訂單
	p.AdjustOrdersForSide(ctx, marketName, existingAsks, idealAskLevels, model.ASK)
}

// adjustOrdersForSide 調整某一邊的訂單
func (p *ProvideLiquidityStrategy) AdjustOrdersForSide(ctx context.Context, marketName string,
	existingOrders map[float64]*dto.Order, idealLevels []PriceLevel, side model.Side) {

	// 1. 新增訂單
	for _, idealLevel := range idealLevels {

		err := p.PlaceNewOrder(ctx, marketName, idealLevel.Price, idealLevel.Volume, side)
		if err != nil {
			log.Warnf("[AMM] Failed to place new order at price %f: %v", idealLevel.Price, err)
		}
	}

	// 2. 取消舊單
	//log.Infof("[AMM] ready to cancel orders, marketName: %s, count: %v", marketName, len(existingOrders))
	for _, existingOrder := range existingOrders {
		_, err := p.ExchangeFuncProxy.CancelOrder(ctx, existingOrder.UserID, existingOrder.ID)
		if err != nil {
			log.Warnf("[AMM] Falied to cancel order: orderId:%s, err:%v", existingOrder.ID, err)
		}
	}
}

// placeNewOrder 下新訂單
func (p *ProvideLiquidityStrategy) PlaceNewOrder(ctx context.Context, marketName string,
	price, volume float64, side model.Side) error {

	orderReq := &dto.OrderReq{
		Side:      side,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     price,
		Size:      volume,
	}

	err := p.ExchangeFuncProxy.PlaceOrder(ctx, p.AmmUser, marketName, orderReq)
	if err != nil {
		return fmt.Errorf("[AMM] failed to place order: %w", err)
	}

	log.Printf("[AMM] Placed new %s order: price=%f, volume=%f",
		map[model.Side]string{model.BID: "BUY", model.ASK: "SELL"}[side], price, volume)

	return nil
}

// roundPrice 價格精度處理
func (p *ProvideLiquidityStrategy) RoundPrice(price float64) float64 {
	// 根據價格大小選擇適當的精度
	if price >= 1000 {
		return math.Round(price*100) / 100 // 2 位小數
	} else if price >= 10 {
		return math.Round(price*1000) / 1000 // 3 位小數
	} else {
		return math.Round(price*10000) / 10000 // 4 位小數
	}
}

// roundVolume 數量精度處理
func (p *ProvideLiquidityStrategy) RoundVolume(volume float64) float64 {
	if volume >= 1 {
		return math.Round(volume*1000) / 1000 // 3 位小數
	} else {
		return math.Round(volume*10000) / 10000 // 4 位小數
	}
}
