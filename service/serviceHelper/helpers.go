package serviceHelper

import (
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"time"
)

func ParseMarket(engine *core.MatchingEngine, market string) (base, quote string, err error) {
	ob, err := engine.GetOrderBook(market)
	if err != nil {
		return "", "", err
	}
	info := ob.MarketInfo()
	return info.BaseAsset, info.QuoteAsset, nil
}

func DetermineFreezeValue(req *dto.OrderReq, base string, quote string) (freezeAsset string, freezeAmt float64) {
	if req.Side == model.BID {
		freezeAsset = quote
		switch req.OrderType {
		case book.LIMIT:
			// limit buy order, freeze price*size
			freezeAmt = req.Price * req.Size
			break
		case book.MARKET:
			// market order freeze quoteAmt
			freezeAmt = req.QuoteAmount
		}
	} else {
		// all ask order just freeze base asset size
		freezeAsset = base
		freezeAmt = req.Size
	}
	return freezeAsset, freezeAmt
}

func NewOrderDtoByOrderReq(market, userID string, req *dto.OrderReq) *dto.Order {
	return &dto.Order{
		ID:            uuid.NewString(),
		UserID:        userID,
		Market:        market,
		Side:          req.Side,
		Price:         req.Price,
		OriginalSize:  req.Size,
		RemainingSize: req.Size,
		QuoteAmount:   req.QuoteAmount,
		Type:          req.OrderType,
		Mode:          req.Mode,
		Status:        model.ORDER_STATUS_NEW,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func NewEngineOrderByOrderDto(orderDto *dto.Order) *model.Order {
	return model.NewOrder(
		orderDto.ID,
		orderDto.UserID,
		orderDto.Side,
		orderDto.Price,
		orderDto.RemainingSize,
		orderDto.QuoteAmount,
		orderDto.Mode)
}

type DealtOrderUpdateData struct {
	OrderID                 string
	RemainingSizeDecreasing float64
}

type DealtOrderSettlementData struct {
	BaseAssetAvailable  float64
	BaseAssetLocked     float64
	QuoteAssetAvailable float64
	QuoteAssetLocked    float64
}

func TidyUpTradesData(baseAsset, quoteAsset string, freezeAsset string, freezeAmt float64, eatenOrder *dto.Order, trades []book.Trade) ([]*DealtOrderUpdateData, map[string]*DealtOrderSettlementData, error) {
	// count of eaten order + matching orders = len(trades)+1
	orderUpdates := make([]*DealtOrderUpdateData, 0, len(trades)+1)
	// uid: DealtOrderSettlementData
	userSettlements := make(map[string]*DealtOrderSettlementData)

	// Add eaten order first.
	orderUpdates = append(orderUpdates, &DealtOrderUpdateData{eatenOrder.ID, eatenOrder.OriginalSize - eatenOrder.RemainingSize})

	isEatenOrderLimitBuy := eatenOrder.Side == model.BID && eatenOrder.Type == book.LIMIT

	userIds := make(map[string]bool)
	for _, trade := range trades {
		userIds[trade.BidUserID] = true
		userIds[trade.AskUserID] = true
	}
	for uid := range userIds {
		userSettlements[uid] = &DealtOrderSettlementData{}
	}

	// loop trades
	for _, trade := range trades {
		bidUid := trade.BidUserID
		askUid := trade.AskUserID
		bidOrderId := trade.BidOrderID
		askOrderId := trade.AskOrderID
		tradePrice := trade.Price
		tradeSize := trade.Size

		askUserSettlementData, _ := userSettlements[askUid]
		bidUserSettlementData, _ := userSettlements[bidUid]

		// 1-1. Process bid user locked quote balance.
		if isEatenOrderLimitBuy {
			// if eaten order is a limit buy, unfreezeQuoteAmt -= eatenOrder.Price * trade.Size
			bidUserSettlementData.QuoteAssetLocked -= eatenOrder.Price * tradeSize
		} else {
			bidUserSettlementData.QuoteAssetLocked -= tradePrice * tradeSize
		}

		// 1-2. Process bid user available base balance.
		bidUserSettlementData.BaseAssetAvailable += tradeSize

		// 2-1. Process ask user locked base balance.
		askUserSettlementData.BaseAssetLocked -= tradeSize

		// 2-2. Process ask user available quote balance.
		askUserSettlementData.QuoteAssetAvailable += tradePrice * tradeSize

		// 3. collect opposite order update info
		if eatenOrder.Side == model.BID {
			orderUpdates = append(orderUpdates, &DealtOrderUpdateData{askOrderId, tradeSize})
		} else {
			orderUpdates = append(orderUpdates, &DealtOrderUpdateData{bidOrderId, tradeSize})
		}
	}

	return orderUpdates, userSettlements, nil
}
