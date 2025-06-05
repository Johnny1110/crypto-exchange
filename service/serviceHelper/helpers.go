package serviceHelper

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

// ParseMarket extracts base and quote assets from market
func ParseMarket(engine *core.MatchingEngine, market string) (string, string, error) {
	if engine == nil {
		return "", "", fmt.Errorf("engine cannot be nil")
	}
	if market == "" {
		return "", "", fmt.Errorf("market cannot be empty")
	}

	orderBook, err := engine.GetOrderBook(market)
	if err != nil {
		return "", "", fmt.Errorf("failed to get order book for market %s: %w", market, err)
	}

	marketInfo := orderBook.MarketInfo()
	if marketInfo.BaseAsset == "" || marketInfo.QuoteAsset == "" {
		return "", "", fmt.Errorf("invalid market info for market %s", market)
	}

	return marketInfo.BaseAsset, marketInfo.QuoteAsset, nil
}

// DetermineFreezeValue calculates which asset and amount to freeze
func DetermineFreezeValue(req *dto.OrderReq, baseAsset, quoteAsset string) (string, float64) {
	if req == nil {
		return "", 0
	}

	switch req.Side {
	case model.BID:
		return calculateBidFreezeValue(req, quoteAsset)
	case model.ASK:
		return baseAsset, req.Size
	default:
		return "", 0
	}
}

func calculateBidFreezeValue(req *dto.OrderReq, quoteAsset string) (string, float64) {
	switch req.OrderType {
	case model.LIMIT:
		return quoteAsset, req.Price * req.Size
	case model.MARKET:
		return quoteAsset, req.QuoteAmount
	default:
		return quoteAsset, 0
	}
}

func NewLimitOrderDtoByOrderReq(market, userID string, req *dto.OrderReq) *dto.Order {
	return dto.NewOrderBuilder().
		WithMarket(market).
		WithUser(userID).
		WithSide(req.Side).
		WithType(model.LIMIT).
		WithMode(req.Mode).
		WithPrice(req.Price).
		WithSize(req.Size).
		Build()
}

func NewMarketOrderDtoByOrderReq(market, userID string, req *dto.OrderReq) *dto.Order {
	builder := dto.NewOrderBuilder().
		WithMarket(market).
		WithUser(userID).
		WithSide(req.Side).
		WithType(model.MARKET).
		WithMode(model.TAKER).
		WithPrice(-1) // Market orders don't have a specific price

	if req.Side == model.BID {
		builder.WithQuoteAmount(req.QuoteAmount)
	} else {
		builder.WithSize(req.Size)
	}

	return builder.Build()
}

func NewEngineOrderByOrderDto(orderDto *dto.Order) *model.Order {
	if orderDto == nil {
		return nil
	}

	return model.NewOrder(
		orderDto.ID,
		orderDto.UserID,
		orderDto.Side,
		orderDto.Price,
		orderDto.RemainingSize,
		orderDto.QuoteAmount,
		orderDto.Mode,
	)
}

// CalculateRefund calculates refund amount for cancelled orders
func CalculateRefund(engine *core.MatchingEngine, market string, engineOrder *model.Order) (unlockAsset string, unlockAmount float64, err error) {
	if engine == nil || engineOrder == nil {
		return "", 0, fmt.Errorf("engine and engineOrder cannot be nil")
	}

	baseAsset, quoteAsset, err := ParseMarket(engine, market)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse market: %w", err)
	}

	switch engineOrder.Side {
	case model.BID:
		return quoteAsset, engineOrder.Price * engineOrder.RemainingSize, nil
	case model.ASK:
		return baseAsset, engineOrder.RemainingSize, nil
	default:
		return "", 0, fmt.Errorf("unknown order side: %v", engineOrder.Side)
	}
}

func WrapPlaceOrderResult(orderDto *dto.Order, trades []book.Trade) *dto.PlaceOrderResult {
	if orderDto == nil {
		return nil
	}

	matches := make([]*dto.Match, 0, len(trades))
	for _, trade := range trades {
		matches = append(matches, &dto.Match{
			Price:     trade.Price,
			Size:      trade.Size,
			Timestamp: trade.Timestamp,
		})
	}

	return &dto.PlaceOrderResult{
		Order:   *orderDto,
		Matches: matches,
	}
}
