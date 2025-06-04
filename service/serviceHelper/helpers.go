package serviceHelper

import (
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/labstack/gommon/log"
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
		AvgDealtPrice: 0.0,
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
	DealtQuoteAmount        float64 // dealt amt (quote)
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

	isEatenOrderLimitBuy := eatenOrder.Side == model.BID && eatenOrder.Type == book.LIMIT

	userIds := make(map[string]bool)
	for _, trade := range trades {
		userIds[trade.BidUserID] = true
		userIds[trade.AskUserID] = true
	}
	for uid := range userIds {
		userSettlements[uid] = &DealtOrderSettlementData{}
	}

	totalDealtAmt := 0.0
	totalDealtSize := 0.0

	// loop trades
	for _, trade := range trades {
		bidUid := trade.BidUserID
		askUid := trade.AskUserID
		bidOrderId := trade.BidOrderID
		askOrderId := trade.AskOrderID
		tradePrice := trade.Price
		tradeSize := trade.Size

		// actual dealt Trade Quote Amt
		actualDealtQuoteAmt := tradePrice * tradeSize
		totalDealtAmt += actualDealtQuoteAmt
		totalDealtSize += tradeSize

		askUserSettlementData, _ := userSettlements[askUid]
		bidUserSettlementData, _ := userSettlements[bidUid]

		// 1-1. Process bid user locked quote balance.
		if isEatenOrderLimitBuy {
			unlockEatenOrderQuoteAmt := eatenOrder.Price * tradeSize
			// if eaten order is a limit buy, unfreezeQuoteAmt -= eatenOrder.Price * trade.Size
			bidUserSettlementData.QuoteAssetLocked -= unlockEatenOrderQuoteAmt
			// refund over locked quote amount:
			bidUserSettlementData.QuoteAssetAvailable += unlockEatenOrderQuoteAmt - actualDealtQuoteAmt
		} else {
			bidUserSettlementData.QuoteAssetLocked -= actualDealtQuoteAmt
		}

		// 1-2. Process bid user available base balance.
		bidUserSettlementData.BaseAssetAvailable += tradeSize

		// 2-1. Process ask user locked base balance.
		askUserSettlementData.BaseAssetLocked -= tradeSize

		// 2-2. Process ask user available quote balance.
		askUserSettlementData.QuoteAssetAvailable += tradePrice * tradeSize

		// 3. collect opposite order update info
		if eatenOrder.Side == model.BID {
			orderUpdates = append(orderUpdates, &DealtOrderUpdateData{OrderID: askOrderId, RemainingSizeDecreasing: tradeSize, DealtQuoteAmount: actualDealtQuoteAmt})
		} else {
			orderUpdates = append(orderUpdates, &DealtOrderUpdateData{OrderID: bidOrderId, RemainingSizeDecreasing: tradeSize, DealtQuoteAmount: actualDealtQuoteAmt})
		}
	}

	// Add eaten order as last one.
	orderUpdates = append(orderUpdates, &DealtOrderUpdateData{OrderID: eatenOrder.ID, RemainingSizeDecreasing: totalDealtSize, DealtQuoteAmount: totalDealtAmt})
	// update eatenOrder
	eatenOrder.AvgDealtPrice = totalDealtAmt / totalDealtSize
	eatenOrder.QuoteAmount = totalDealtAmt

	return orderUpdates, userSettlements, nil
}

func CalculateRefund(engine *core.MatchingEngine, market string, engineOrder *model.Order) (unlockAsset string, unlockAmount float64, err error) {
	baseAsset, quoteAsset, err := ParseMarket(engine, market)
	if err != nil {
		log.Errorf("[CalculateRefund] ParseMarket err: %v", err)
		return "", 0, err
	}

	switch engineOrder.Side {
	case model.BID:
		unlockAmount = engineOrder.Price * engineOrder.RemainingSize
		unlockAsset = quoteAsset
		break
	case model.ASK:
		unlockAmount = engineOrder.RemainingSize
		unlockAsset = baseAsset
	}
	return unlockAsset, unlockAmount, nil
}

func WrapPlaceOrderResult(orderDto *dto.Order, trades []book.Trade) *dto.PlaceOrderResult {
	matches := make([]*dto.Match, 0, len(trades))

	for _, trade := range trades {
		match := &dto.Match{
			Price:     trade.Price,
			Size:      trade.Size,
			Timestamp: trade.Timestamp,
		}
		matches = append(matches, match)
	}

	res := &dto.PlaceOrderResult{
		Order:   *orderDto,
		Matches: matches,
	}
	return res
}
