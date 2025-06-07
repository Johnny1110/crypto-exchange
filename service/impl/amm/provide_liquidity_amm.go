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
type provideLiquidityStrategy struct {
	exchangeFuncProxy IAmmExchangeFuncProxy
	ammUID            string
	ammUser           dto.User
}

func NewProvideLiquidityStrategy(exchangeFuncProxy IAmmExchangeFuncProxy, ammUser dto.User) AutoMarketStrategy {
	return &provideLiquidityStrategy{
		exchangeFuncProxy: exchangeFuncProxy,
		ammUID:            ammUser.ID,
		ammUser:           ammUser,
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

// MakeMarket provideLiquidityStrategy can access exchangeFuncProxy to make market
func (p *provideLiquidityStrategy) MakeMarket(ctx context.Context, marketInfo market.MarketInfo, maxQuoteAmtPerLevel float64) {
	marketName := marketInfo.Name

	// 1. 獲取指數價格作為中間價
	indexPrice, err := p.exchangeFuncProxy.GetIndexPrice(ctx, marketName)
	if err != nil {
		log.Printf("Failed to get index price for %s: %v", marketName, err)
		return
	}

	// 2. 獲取當前餘額
	balance, err := p.exchangeFuncProxy.GetBalance(ctx, p.ammUID, marketName)
	if err != nil {
		log.Printf("Failed to get balance for %s: %v", marketName, err)
		return
	}

	// 3. 獲取當前開放訂單
	openOrders, err := p.exchangeFuncProxy.GetOpenOrders(ctx, p.ammUID, marketName)
	if err != nil {
		log.Printf("Failed to get open orders for %s: %v", marketName, err)
		return
	}

	// 4. 計算理想的價格檔位
	idealBidLevels := p.calculateIdealPriceLevels(indexPrice, model.BID, balance, maxQuoteAmtPerLevel)
	idealAskLevels := p.calculateIdealPriceLevels(indexPrice, model.ASK, balance, maxQuoteAmtPerLevel)

	// 5. 分析現有訂單並進行調整
	p.adjustOrders(ctx, marketName, openOrders, idealBidLevels, idealAskLevels)
}

// calculateIdealPriceLevels 計算理想的價格檔位
func (p *provideLiquidityStrategy) calculateIdealPriceLevels(indexPrice float64, side model.Side, balance Balance, maxQuoteAmtPerLevel float64) []PriceLevel {
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
				Price:  p.roundPrice(price),
				Volume: p.roundVolume(maxVolume),
			})
		}
	}

	return levels
}

// adjustOrders 調整訂單以符合理想檔位
func (p *provideLiquidityStrategy) adjustOrders(ctx context.Context, marketName string,
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
	p.adjustOrdersForSide(ctx, marketName, existingBids, idealBidLevels, model.BID)

	// 調整 Ask 訂單
	p.adjustOrdersForSide(ctx, marketName, existingAsks, idealAskLevels, model.ASK)
}

// adjustOrdersForSide 調整某一邊的訂單
func (p *provideLiquidityStrategy) adjustOrdersForSide(ctx context.Context, marketName string,
	existingOrders map[float64]*dto.Order, idealLevels []PriceLevel, side model.Side) {

	// 創建理想價格的 map 以便快速查找
	idealPrices := make(map[float64]PriceLevel)
	for _, level := range idealLevels {
		idealPrices[level.Price] = level
	}

	// 1. 檢查需要取消的訂單（價格不在理想檔位中，或數量需要調整）
	var ordersToCancel []*dto.Order
	for price, order := range existingOrders {
		if idealLevel, exists := idealPrices[price]; !exists {
			// 價格不在理想檔位中，需要取消
			ordersToCancel = append(ordersToCancel, order)
		} else if math.Abs(order.RemainingSize-idealLevel.Volume) > idealLevel.Volume*0.1 {
			// 數量差異超過 10%，需要重新下單
			ordersToCancel = append(ordersToCancel, order)
		}
	}

	// 2. 取消不合適的訂單
	for _, order := range ordersToCancel {
		_, err := p.exchangeFuncProxy.CancelOrder(ctx, p.ammUID, order.ID)
		if err != nil {
			log.Printf("Failed to cancel order %s: %v", order.ID, err)
			continue
		}
		delete(existingOrders, order.Price)
		log.Printf("Canceled order: %s at price %f", order.ID, order.Price)
	}

	// 3. 檢查需要新增的訂單
	for _, idealLevel := range idealLevels {
		if _, exists := existingOrders[idealLevel.Price]; !exists {
			// 這個價格檔位沒有訂單，需要新增
			err := p.placeNewOrder(ctx, marketName, idealLevel.Price, idealLevel.Volume, side)
			if err != nil {
				log.Printf("Failed to place new order at price %f: %v", idealLevel.Price, err)
			}
		}
	}
}

// placeNewOrder 下新訂單
func (p *provideLiquidityStrategy) placeNewOrder(ctx context.Context, marketName string,
	price, volume float64, side model.Side) error {

	orderReq := &dto.OrderReq{
		Side:      side,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     price,
		Size:      volume,
	}

	err := p.exchangeFuncProxy.PlaceOrder(ctx, p.ammUser, marketName, orderReq)
	if err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}

	log.Printf("Placed new %s order: price=%f, volume=%f",
		map[model.Side]string{model.BID: "BUY", model.ASK: "SELL"}[side], price, volume)

	return nil
}

// roundPrice 價格精度處理
func (p *provideLiquidityStrategy) roundPrice(price float64) float64 {
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
func (p *provideLiquidityStrategy) roundVolume(volume float64) float64 {
	if volume >= 1 {
		return math.Round(volume*1000) / 1000 // 3 位小數
	} else {
		return math.Round(volume*10000) / 10000 // 4 位小數
	}
}
