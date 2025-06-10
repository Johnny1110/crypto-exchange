package serviceHelper

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/utils"
)

// OrderUpdateData represents data needed to update a dealt order
type OrderUpdateData struct {
	OrderID                    string
	RemainingSizeDecreasing    float64
	DealtQuoteAmountIncreasing float64
	FeesIncreasing             float64
}

// UserSettlementData represents settlement data for a user's assets.html
type UserSettlementData struct {
	BaseAssetAvailable  float64
	BaseAssetLocked     float64
	QuoteAssetAvailable float64
	QuoteAssetLocked    float64
}

// TradeSettlementResult encapsulates the result of trade settlement processing
type TradeSettlementResult struct {
	BaseAsset       string
	QuoteAsset      string
	OrderUpdates    []*OrderUpdateData
	UserSettlements map[string]*UserSettlementData
	TotalDealtAmt   float64
	TotalDealtSize  float64
	TotalBaseFees   float64 // add to settings margin account balances
	TotalQuoteFees  float64 // add to settings margin account balances
}

// ProcessTradeSettlement handles the core logic for processing trades and updating balances
func ProcessTradeSettlement(ctx *dto.PlaceOrderContext) (*TradeSettlementResult, error) {
	eatenOrder := ctx.OrderDTO
	trades := ctx.Trades

	if eatenOrder == nil {
		return nil, fmt.Errorf("eaten order cannot be nil")
	}

	result := &TradeSettlementResult{
		OrderUpdates:    make([]*OrderUpdateData, 0, len(trades)+1),
		UserSettlements: initializeUserSettlements(trades),
		BaseAsset:       ctx.Assets.BaseAsset,
		QuoteAsset:      ctx.Assets.QuoteAsset,
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
	bidFees := r.processBidUserBalances(bidSettlement, trade, tradeQuoteAmount, eatenOrder)

	// Process ask user balances
	askFees := r.processAskUserBalances(askSettlement, trade, tradeQuoteAmount)

	// Add opposite order update
	r.addOppositeOrderUpdate(trade, eatenOrder, tradeQuoteAmount, bidFees, askFees)
}

// processBidUserBalances handles bid user's balance updates and return bid fees (Base Asset)
func (r *TradeSettlementResult) processBidUserBalances(bidSettlement *UserSettlementData, trade book.Trade, tradeQuoteAmount float64, eatenOrder *dto.Order) (fees float64) {
	// Handle quote asset (what bid user pays)
	if eatenOrder.Type == model.LIMIT && eatenOrder.Side == model.BID {
		// If processing bid is incoming eatenOrder.
		// For limit buy orders, unlock at order price and refund difference
		unlockAmount := eatenOrder.Price * trade.Size
		bidSettlement.QuoteAssetLocked -= unlockAmount
		bidSettlement.QuoteAssetLocked = utils.RoundFloat(bidSettlement.QuoteAssetLocked)
		bidSettlement.QuoteAssetAvailable += unlockAmount - tradeQuoteAmount // Refund overpayment
		bidSettlement.QuoteAssetAvailable = utils.RoundFloat(bidSettlement.QuoteAssetAvailable)
	} else {
		// For other orders, unlock exact trade amount
		bidSettlement.QuoteAssetLocked -= tradeQuoteAmount
		bidSettlement.QuoteAssetLocked = utils.RoundFloat(bidSettlement.QuoteAssetLocked)
	}

	// Calculate fees and accumulate to sum.
	bidFees := trade.Size * trade.BidFeeRate
	r.TotalBaseFees += bidFees

	// Add base asset received (deduct fees)
	bidSettlement.BaseAssetAvailable += trade.Size - bidFees
	bidSettlement.BaseAssetAvailable = utils.RoundFloat(bidSettlement.BaseAssetAvailable)

	return bidFees
}

// processAskUserBalances handles ask user's balance updates and return ask fees (Quote Asset)
func (r *TradeSettlementResult) processAskUserBalances(askSettlement *UserSettlementData, trade book.Trade, tradeQuoteAmount float64) (fees float64) {
	// Remove locked base asset (what ask user sells)
	askSettlement.BaseAssetLocked -= trade.Size
	askSettlement.BaseAssetLocked = utils.RoundFloat(askSettlement.BaseAssetLocked)

	// Calculate fees and accumulate to sum.
	askFees := tradeQuoteAmount * trade.AskFeeRate
	r.TotalQuoteFees += askFees

	// Add quote asset received
	askSettlement.QuoteAssetAvailable += tradeQuoteAmount - askFees
	askSettlement.QuoteAssetAvailable = utils.RoundFloat(askSettlement.QuoteAssetAvailable)

	return askFees
}

// addOppositeOrderUpdate adds update data for the order opposite to the eaten order
func (r *TradeSettlementResult) addOppositeOrderUpdate(trade book.Trade, eatenOrder *dto.Order, tradeQuoteAmount, bidFees, askFees float64) {
	var oppositeOrderId string
	var feeIncreasing float64
	if eatenOrder.Side == model.BID {
		oppositeOrderId = trade.AskOrderID
		feeIncreasing = askFees
	} else {
		oppositeOrderId = trade.BidOrderID
		feeIncreasing = bidFees
	}

	r.OrderUpdates = append(r.OrderUpdates, &OrderUpdateData{
		OrderID:                    oppositeOrderId,
		RemainingSizeDecreasing:    utils.RoundFloat(trade.Size),
		DealtQuoteAmountIncreasing: utils.RoundFloat(tradeQuoteAmount),
		FeesIncreasing:             feeIncreasing,
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
			FeesIncreasing:             r.TotalBaseFees,
		}
	} else {
		// Limit orders and market sell orders need full updates
		var fees float64
		if eatenOrder.Side == model.BID {
			fees = r.TotalBaseFees
		} else {
			fees = r.TotalQuoteFees
		}

		update = &OrderUpdateData{
			OrderID:                    eatenOrder.ID,
			RemainingSizeDecreasing:    utils.RoundFloat(r.TotalDealtSize),
			DealtQuoteAmountIncreasing: utils.RoundFloat(r.TotalDealtAmt),
			FeesIncreasing:             fees,
		}
	}
	eatenOrder.Fees += update.FeesIncreasing

	r.OrderUpdates = append(r.OrderUpdates, update)
}
