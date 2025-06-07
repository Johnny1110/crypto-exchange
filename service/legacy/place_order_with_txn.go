package legacy

//
//import (
//	"database/sql"
//	"fmt"
//	"github.com/johnny1110/crypto-exchange/engine-v2/model"
//	"github.com/johnny1110/crypto-exchange/service/serviceHelper"
//	"github.com/labstack/gommon/log"
//)
//
//func (s *orderService) executeOrderPlacementPhase(ctx context.Context, orderCtx *dto.PlaceOrderContext, isMarketOrder bool) error {
//	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
//		// 1. Freeze user funds
//		if err := s.balanceRepo.LockedByUserIdAndAsset(ctx, tx, orderCtx.UserID, orderCtx.Assets.FreezeAsset, orderCtx.Assets.FreezeAmt); err != nil {
//			log.Warnf("[executeOrderPlacementPhase] failed to lock user balance, %v", err)
//			return ErrInsufficientBalance
//		}
//
//		// 2. Insert order to database
//		if err := s.orderRepo.Insert(ctx, tx, orderCtx.OrderDTO); err != nil {
//			log.Errorf("[executeOrderPlacementPhase] Insert Order error : %v", err)
//			return UnknownError
//		}
//
//		// 3. Place order in matching engine
//		engineOrder := serviceHelper.NewEngineOrderByOrderDto(orderCtx.OrderDTO)
//		trades, err := s.engine.PlaceOrder(orderCtx.Market, orderCtx.Request.OrderType, engineOrder)
//		if err != nil {
//			log.Warnf("[executeOrderPlacementPhase] Engine warning : %v", err)
//			return err
//		}
//
//		// 4. Update order status from engine result
//		orderCtx.SyncTradeResult(engineOrder, trades)
//
//		// 5. Save trade records
//		if len(orderCtx.Trades) > 0 {
//			if err := s.tradeRepo.BatchInsert(ctx, tx, trades); err != nil {
//				log.Errorf("[executeOrderPlacementPhase] BatchInsert Trades error : %v", err)
//				return UnknownError
//			}
//		}
//
//		// 6. Handle market bid order special case
//		if isMarketOrder && orderCtx.Request.Side == model.BID {
//			if err := s.orderRepo.UpdateOriginalSize(ctx, tx, engineOrder.ID, engineOrder.OriginalSize); err != nil {
//				log.Errorf("[executeOrderPlacementPhase] Handle market bid order special case error : %v", err)
//				return UnknownError
//			}
//			orderCtx.OrderDTO.OriginalSize = engineOrder.OriginalSize
//		}
//
//		return nil
//	})
//}
//
//func (s *orderService) executeTradeSettlementPhase(ctx context.Context, orderCtx *dto.PlaceOrderContext) error {
//	if len(orderCtx.Trades) == 0 {
//		return nil // No trades to settle
//	}
//
//	settlementResult, err := serviceHelper.ProcessTradeSettlement(orderCtx)
//	if err != nil {
//		return fmt.Errorf("failed to process trade settlement: %w", err)
//	}
//
//	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
//		// Update orders
//		for _, orderUpdate := range settlementResult.OrderUpdates {
//			if err := s.orderRepo.SyncTradeMatchingResult(ctx, tx, orderUpdate.OrderID, orderUpdate.RemainingSizeDecreasing, orderUpdate.DealtQuoteAmountIncreasing, orderUpdate.FeesIncreasing); err != nil {
//				return fmt.Errorf("failed to sync trade matching result for order %s: %w", orderUpdate.OrderID, err)
//			}
//		}
//
//		// Update user balances
//		for userID, settlement := range settlementResult.UserSettlements {
//			if err := s.updateUserAssets(ctx, tx, userID, orderCtx.Assets, settlement); err != nil {
//				log.Errorf("updateUserAssets error: %v", err)
//				return err
//			}
//		}
//
//		// settle Fees Revenue to exchange's margin account
//		if err := s.settleFeesRevenue(ctx, tx, settlementResult); err != nil {
//			return err
//		}
//
//		return nil
//	})
//}
