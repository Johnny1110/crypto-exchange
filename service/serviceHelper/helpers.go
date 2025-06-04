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

func NewLimitOrderDtoByOrderReq(market, userID string, req *dto.OrderReq) *dto.Order {
	return &dto.Order{
		ID:            uuid.NewString(),
		UserID:        userID,
		Market:        market,
		Side:          req.Side,
		Price:         req.Price,
		OriginalSize:  req.Size,
		RemainingSize: req.Size,
		QuoteAmount:   0.0,
		AvgDealtPrice: 0.0,
		Type:          book.LIMIT,
		Mode:          req.Mode,
		Status:        model.ORDER_STATUS_NEW,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func NewMarketOrderDtoByOrderReq(market, userID string, req *dto.OrderReq) *dto.Order {
	size := 0.0
	quoteAmt := 0.0
	if req.Side == model.BID {
		quoteAmt = req.QuoteAmount
	} else {
		size = req.Size
	}

	return &dto.Order{
		ID:            uuid.NewString(),
		UserID:        userID,
		Market:        market,
		Side:          req.Side,
		Price:         -1,
		OriginalSize:  size,
		RemainingSize: size,
		QuoteAmount:   quoteAmt,
		AvgDealtPrice: 0.0,
		Type:          book.MARKET,
		Mode:          model.TAKER,
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
