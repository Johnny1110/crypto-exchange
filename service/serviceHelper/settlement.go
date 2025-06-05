package serviceHelper

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

// OrderUpdateData represents data needed to update a dealt order
type OrderUpdateData struct {
	OrderID                    string
	RemainingSizeDecreasing    float64
	DealtQuoteAmountIncreasing float64
}

// UserSettlementData represents settlement data for a user's assets
type UserSettlementData struct {
	BaseAssetAvailable  float64
	BaseAssetLocked     float64
	QuoteAssetAvailable float64
	QuoteAssetLocked    float64
}

// TradeSettlementResult encapsulates the result of trade settlement processing
type TradeSettlementResult struct {
	OrderUpdates    []*OrderUpdateData
	UserSettlements map[string]*UserSettlementData
	TotalDealtAmt   float64
	TotalDealtSize  float64
}

// ProcessTradeSettlement handles the core logic for processing trades and updating balances
func ProcessTradeSettlement(eatenOrder *dto.Order, trades []book.Trade) (*TradeSettlementResult, error) {
	if eatenOrder == nil {
		return nil, fmt.Errorf("eaten order cannot be nil")
	}

	result := &TradeSettlementResult{
		OrderUpdates:    make([]*OrderUpdateData, 0, len(trades)+1),
		UserSettlements: initializeUserSettlements(trades),
	}

	// Process each trade
	for _, trade := range trades {
		result.processIndividualTrade(trade, eatenOrder)
	}

	// Add the eaten order to updates
	result.addEatenOrderUpdate(eatenOrder)

	// Update eaten order statistics
	if result.TotalDealtSize > 0 {
		eatenOrder.AvgDealtPrice = result.TotalDealtAmt / result.TotalDealtSize
		eatenOrder.QuoteAmount = result.TotalDealtAmt
	}

	return result, nil
}

// initializeUserSettlements creates settlement data for all users involved in trades
func initializeUserSettlements(trades []book.Trade) map[string]*UserSettlementData {
	userIds := extractUniqueUserIds(trades)
	settlements := make(map[string]*UserSettlementData, len(userIds))

	for uid := range userIds {
		settlements[uid] = &UserSettlementData{}
	}

	return settlements
}

// extractUniqueUserIds gets all unique user IDs from trades
func extractUniqueUserIds(trades []book.Trade) map[string]bool {
	userIds := make(map[string]bool, len(trades)*2) // Preallocate for efficiency

	for _, trade := range trades {
		userIds[trade.BidUserID] = true
		userIds[trade.AskUserID] = true
	}

	return userIds
}

// processIndividualTrade handles the settlement logic for a single trade
func (r *TradeSettlementResult) processIndividualTrade(trade book.Trade, eatenOrder *dto.Order) {
	tradeQuoteAmount := trade.Price * trade.Size
	r.TotalDealtAmt += tradeQuoteAmount
	r.TotalDealtSize += trade.Size

	bidSettlement := r.UserSettlements[trade.BidUserID]
	askSettlement := r.UserSettlements[trade.AskUserID]

	// Process bid user balances
	r.processBidUserBalances(bidSettlement, trade, tradeQuoteAmount, eatenOrder)

	// Process ask user balances
	r.processAskUserBalances(askSettlement, trade, tradeQuoteAmount)

	// Add opposite order update
	r.addOppositeOrderUpdate(trade, eatenOrder, tradeQuoteAmount)
}

// processBidUserBalances handles bid user's balance updates
func (r *TradeSettlementResult) processBidUserBalances(bidSettlement *UserSettlementData, trade book.Trade, tradeQuoteAmount float64, eatenOrder *dto.Order) {
	// Handle quote asset (what bid user pays)
	if eatenOrder.Type == model.LIMIT && eatenOrder.Side == model.BID {
		// If processing bid is incoming eatenOrder.
		// For limit buy orders, unlock at order price and refund difference
		unlockAmount := eatenOrder.Price * trade.Size
		bidSettlement.QuoteAssetLocked -= unlockAmount
		bidSettlement.QuoteAssetAvailable += unlockAmount - tradeQuoteAmount // Refund overpayment
	} else {
		// For other orders, unlock exact trade amount
		bidSettlement.QuoteAssetLocked -= tradeQuoteAmount
	}

	// Add base asset received
	bidSettlement.BaseAssetAvailable += trade.Size
}

// processAskUserBalances handles ask user's balance updates
func (r *TradeSettlementResult) processAskUserBalances(settlement *UserSettlementData, trade book.Trade, tradeQuoteAmount float64) {
	// Remove locked base asset (what ask user sells)
	settlement.BaseAssetLocked -= trade.Size

	// Add quote asset received
	settlement.QuoteAssetAvailable += tradeQuoteAmount
}

// addOppositeOrderUpdate adds update data for the order opposite to the eaten order
func (r *TradeSettlementResult) addOppositeOrderUpdate(trade book.Trade, eatenOrder *dto.Order, tradeQuoteAmount float64) {
	var oppositeOrderId string
	if eatenOrder.Side == model.BID {
		oppositeOrderId = trade.AskOrderID
	} else {
		oppositeOrderId = trade.BidOrderID
	}

	r.OrderUpdates = append(r.OrderUpdates, &OrderUpdateData{
		OrderID:                    oppositeOrderId,
		RemainingSizeDecreasing:    trade.Size,
		DealtQuoteAmountIncreasing: tradeQuoteAmount,
	})
}

// addEatenOrderUpdate adds the eaten order to the updates list
func (r *TradeSettlementResult) addEatenOrderUpdate(eatenOrder *dto.Order) {
	var update *OrderUpdateData

	if eatenOrder.Type == model.MARKET && eatenOrder.Side == model.BID {
		// Market bid orders don't need size/amount updates as they're already processed
		update = &OrderUpdateData{
			OrderID:                    eatenOrder.ID,
			RemainingSizeDecreasing:    0.0,
			DealtQuoteAmountIncreasing: 0.0,
		}
	} else {
		// Limit orders and market sell orders need full updates
		update = &OrderUpdateData{
			OrderID:                    eatenOrder.ID,
			RemainingSizeDecreasing:    r.TotalDealtSize,
			DealtQuoteAmountIncreasing: r.TotalDealtAmt,
		}
	}

	r.OrderUpdates = append(r.OrderUpdates, update)
}
