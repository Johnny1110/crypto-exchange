package serviceImpl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/service/serviceHelper"
	"github.com/labstack/gommon/log"
)

type orderService struct {
	db          *sql.DB
	engine      *core.MatchingEngine
	orderRepo   repository.IOrderRepository
	tradeRepo   repository.ITradeRepository
	balanceRepo repository.IBalanceRepository
}

func NewIOrderService(
	db *sql.DB,
	engine *core.MatchingEngine,
	orderRepo repository.IOrderRepository,
	tradeRepo repository.ITradeRepository,
	balanceRepo repository.IBalanceRepository) service.IOrderService {
	return &orderService{
		db:          db,
		engine:      engine,
		orderRepo:   orderRepo,
		tradeRepo:   tradeRepo,
		balanceRepo: balanceRepo,
	}
}

func (s *orderService) PlaceOrder(ctx context.Context, market, userID string, req *dto.OrderReq) (*dto.PlaceOrderResult, error) {
	err := validatePlacingOrderReq(userID, market, req)
	if err != nil {
		log.Errorf("[PlaceOrder] validatePlacingOrderReq err: %v", err)
		return nil, err
	}

	baseAsset, quoteAsset, err := serviceHelper.ParseMarket(s.engine, market)
	if err != nil {
		log.Errorf("[PlaceOrder] ParseMarket err: %v", err)
		return nil, err
	}
	freezeAsset, freezeAmt := serviceHelper.DetermineFreezeValue(req, baseAsset, quoteAsset)

	assetDetails := &orderAssetDetails{
		freezeAmount: freezeAmt,
		baseAsset:    baseAsset,
		quoteAsset:   quoteAsset,
		freezeAsset:  freezeAsset,
	}

	// call placing limit order | market order:
	var orderDto *dto.Order
	var trades []book.Trade
	switch req.OrderType {
	case book.LIMIT:
		orderDto, trades, err = s.placingLimitOrder(ctx, market, userID, req, assetDetails)
		break
	case book.MARKET:
		orderDto, trades, err = s.placingMarketOrder(ctx, market, userID, req, assetDetails)
		break
	}

	if err != nil {
		log.Errorf("[PlaceOrder] failed to placing type order, err: %v", err)
		return nil, err
	}

	return serviceHelper.WrapPlaceOrderResult(orderDto, trades), err
}

// Core Logic (Placing 2 types of Order) >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type orderAssetDetails struct {
	baseAsset, quoteAsset, freezeAsset string
	freezeAmount                       float64
}

func (s *orderService) placingLimitOrder(ctx context.Context, market, userID string, req *dto.OrderReq, assetDetails *orderAssetDetails) (*dto.Order, []book.Trade, error) {
	orderDto := serviceHelper.NewLimitOrderDtoByOrderReq(market, userID, req)

	log.Infof("[placingLimitOrder] orderDto: %v", orderDto)

	var err error
	var trades []book.Trade

	freezeAsset := assetDetails.freezeAsset // The Asset that will be locked by incoming order's userId
	freezeAmt := assetDetails.freezeAmount  // The Asset's Amount that will be locked by incoming order's userId
	baseAsset := assetDetails.baseAsset     // Market base asset (ex: ETH, BTC)
	quoteAsset := assetDetails.quoteAsset   // Market quote asset (ex: USDT, USDC)

	// Txn-1: process placing order flow.
	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		// 1. Freeze funds based on market and side.
		err = s.balanceRepo.LockedByUserIdAndAsset(ctx, tx, userID, freezeAsset, freezeAmt)
		if err != nil {
			log.Warnf("[placingLimitOrder] BalanceRepo.LockedByUserIdAndAsset err: %v", err)
			return err
		}

		// 2. Save new orderDto into DB.
		err = s.orderRepo.Insert(ctx, tx, orderDto)
		if err != nil {
			log.Errorf("[placingLimitOrder] OrderRepo.Insert err: %v", err)
			return err
		}

		// 3. Placing Order into engine
		engineOrder := serviceHelper.NewEngineOrderByOrderDto(orderDto)
		trades, err = s.engine.PlaceOrder(market, req.OrderType, engineOrder)
		if err != nil {
			log.Errorf("[placingLimitOrder] engine.PlaceOrder err: %v", err)
			return err
		}
		// dump engineOrder size&status to orderDto
		orderDto.RemainingSize = engineOrder.RemainingSize
		orderDto.Status = engineOrder.GetStatus()

		// 4. Save all matching trade details
		err = s.tradeRepo.BatchInsert(ctx, tx, trades)
		if err != nil {
			log.Errorf("[placingLimitOrder] tradeRepo.BatchInsert err: %v", err)
			return err
		}

		return err
	})

	// IF Txn-1 Got Error:
	if err != nil {
		log.Errorf("[placingLimitOrder] PlaceOrder Txn-1 process has err: %v", err)
		return nil, trades, err
	}

	// Collect trades data to updateOrderDataList and settlementList
	settlementResult, err := serviceHelper.ProcessTradeSettlement(orderDto, trades)
	if err != nil {
		log.Errorf("[placingLimitOrder] TidyUpTradesData err: %v", err)
		return nil, trades, err
	}

	// Txn-2: handle matching trades flow and update orders.
	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		for _, ou := range settlementResult.OrderUpdates {
			// decreasing order remainingSize and update quoteAmt, avgDealtAmt
			err = s.orderRepo.SyncTradeMatchingResult(ctx, tx, ou.OrderID, ou.RemainingSizeDecreasing, ou.DealtQuoteAmountIncreasing)
			if err != nil {
				log.Errorf("[placingLimitOrder] SyncTradeMatchingResult err: %v", err)
				return err
			}
		}

		for userId, us := range settlementResult.UserSettlements {
			err = s.balanceRepo.UpdateAsset(ctx, tx, userId, baseAsset, us.BaseAssetAvailable, us.BaseAssetLocked)
			if err != nil {
				log.Errorf("[placingLimitOrder] Update Base Asset err: %v", err)
				return err
			}
			err = s.balanceRepo.UpdateAsset(ctx, tx, userId, quoteAsset, us.QuoteAssetAvailable, us.QuoteAssetLocked)
			if err != nil {
				log.Errorf("[placingLimitOrder] Update Quote Asset err: %v", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Errorf("[placingLimitOrder] PlaceOrder Txn-2 process has err: %v", err)
		return nil, trades, err
	}

	return orderDto, trades, nil
}

func (s *orderService) placingMarketOrder(ctx context.Context, market, userID string, req *dto.OrderReq, assetDetails *orderAssetDetails) (*dto.Order, []book.Trade, error) {
	orderDto := serviceHelper.NewMarketOrderDtoByOrderReq(market, userID, req)
	log.Infof("[placingMarketOrder] orderDto: %v", orderDto)

	var err error
	var trades []book.Trade

	freezeAsset := assetDetails.freezeAsset // The Asset that will be locked by incoming order's userId
	freezeAmt := assetDetails.freezeAmount  // The Asset's Amount that will be locked by incoming order's userId
	baseAsset := assetDetails.baseAsset     // Market base asset (ex: ETH, BTC)
	quoteAsset := assetDetails.quoteAsset   // Market quote asset (ex: USDT, USDC)

	// Txn-1: process placing order flow.
	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		// 1. Freeze funds based on market and side.
		err = s.balanceRepo.LockedByUserIdAndAsset(ctx, tx, userID, freezeAsset, freezeAmt)
		if err != nil {
			log.Warnf("[placingMarketOrder] BalanceRepo.LockedByUserIdAndAsset err: %v", err)
			return err
		}

		// 2. Save new orderDto into DB.
		err = s.orderRepo.Insert(ctx, tx, orderDto)
		if err != nil {
			log.Errorf("[placingMarketOrder] OrderRepo.Insert err: %v", err)
			return err
		}

		// 3. Placing Order into engine
		engineOrder := serviceHelper.NewEngineOrderByOrderDto(orderDto)
		trades, err = s.engine.PlaceOrder(market, req.OrderType, engineOrder)
		if err != nil {
			log.Errorf("[placingMarketOrder] engine.PlaceOrder err: %v", err)
			return err
		}

		// dump engineOrder size&status to orderDto
		orderDto.RemainingSize = engineOrder.RemainingSize
		orderDto.Status = engineOrder.GetStatus()

		// 4. Save all matching trade details
		err = s.tradeRepo.BatchInsert(ctx, tx, trades)
		if err != nil {
			log.Errorf("[placingMarketOrderplacingMarketOrder] tradeRepo.BatchInsert err: %v", err)
			return err
		}

		// 5. MarketBidOrder need update: order.OriginalSize, because we can only know how much qty has been trade by MarketBidOrder.
		if engineOrder.Side == model.BID {
			err = s.orderRepo.UpdateOriginalSize(ctx, tx, engineOrder.ID, engineOrder.OriginalSize)
			orderDto.OriginalSize = engineOrder.OriginalSize
			if err != nil {
				log.Errorf("[placingMarketOrder] orderRepo.UpdateOriginalSize err: %v", err)
				return err
			}
		}
		return err
	})

	// IF Txn-1 Got Error:
	if err != nil {
		log.Errorf("[placingMarketOrder] PlaceOrder Txn-1 process has err: %v", err)
		return nil, trades, err
	}

	// Collect trades data to updateOrderDataList and settlementList
	settlementResult, err := serviceHelper.ProcessTradeSettlement(orderDto, trades)
	if err != nil {
		log.Errorf("[placingMarketOrder] TidyUpTradesData err: %v", err)
		return nil, trades, err
	}

	// Txn-2: handle matching trades flow and update orders.
	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		for _, ou := range settlementResult.OrderUpdates {
			// decreasing order remainingSize and update quoteAmt, avgDealtAmt
			err = s.orderRepo.SyncTradeMatchingResult(ctx, tx, ou.OrderID, ou.RemainingSizeDecreasing, ou.DealtQuoteAmountIncreasing)
			if err != nil {
				log.Errorf("[placingMarketOrder] SyncTradeMatchingResult err: %v", err)
				return err
			}
		}

		for userId, us := range settlementResult.UserSettlements {
			err = s.balanceRepo.UpdateAsset(ctx, tx, userId, baseAsset, us.BaseAssetAvailable, us.BaseAssetLocked)
			if err != nil {
				log.Errorf("[placingMarketOrder] Update Base Asset err: %v", err)
				return err
			}
			err = s.balanceRepo.UpdateAsset(ctx, tx, userId, quoteAsset, us.QuoteAssetAvailable, us.QuoteAssetLocked)
			if err != nil {
				log.Errorf("[placingMarketOrder] Update Quote Asset err: %v", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Errorf("[placingMarketOrder] PlaceOrder Txn-2 process has err: %v", err)
		return nil, trades, err
	}

	return orderDto, trades, nil
}

// Core Logic (Placing 2 types of Order) <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func (s *orderService) CancelOrder(ctx context.Context, market, userID, orderID string) (*dto.Order, error) {
	if market == "" || userID == "" || orderID == "" {
		return nil, fmt.Errorf("invalid input")
	}

	orderDto, err := s.orderRepo.GetOrderByOrderId(ctx, s.db, orderID)
	if err != nil {
		log.Warnf("[CancelOrder] orderRepo.GetOrderByOrderId err: %v", err)
		return nil, err
	}

	if orderDto.UserID != userID {
		log.Warnf("[CancelOrder] failed, order not belongs to user: %v", err)
		return nil, fmt.Errorf("order not belongs to user")
	}

	engineOrder, err := s.engine.CancelOrder(market, orderID)
	if err != nil {
		log.Errorf("[CancelOrder] engine.CancelOrder err: %v", err)
		return nil, err
	}

	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		// update order
		orderDto.RemainingSize = engineOrder.RemainingSize
		orderDto.Status = model.ORDER_STATUS_CANCELED
		err = s.orderRepo.Update(ctx, tx, orderDto)
		if err != nil {
			log.Errorf("[CancelOrder] orderRepo.CancelOrder err: %v", err)
			return err
		}
		// refund user balances
		unlockAsset, unlockAmount, err := serviceHelper.CalculateRefund(s.engine, market, engineOrder)
		if err != nil {
			log.Errorf("[CancelOrder] calculate refund error: %v", err)
		}
		err = s.balanceRepo.UnlockedByUserIdAndAsset(ctx, tx, userID, unlockAsset, unlockAmount)
		if err != nil {
			log.Errorf("[CancelOrder] balanceRepo.UnlockedByUserIdAndAsset err: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Errorf("[CancelOrder] WithTx err: %v", err)
		return nil, err
	}

	return orderDto, nil
}

func (s *orderService) QueryOrder(ctx context.Context, userID string, isOpenOrder bool) ([]*dto.Order, error) {
	statuses := orderStatusesByOpenFlag(isOpenOrder)
	return s.orderRepo.GetOrdersByUserIdAndStatuses(ctx, s.db, userID, statuses)
}

func orderStatusesByOpenFlag(isOpen bool) []model.OrderStatus {
	if isOpen {
		return []model.OrderStatus{
			model.ORDER_STATUS_NEW,
			model.ORDER_STATUS_PARTIAL,
		}
	}
	return []model.OrderStatus{
		model.ORDER_STATUS_CANCELED,
		model.ORDER_STATUS_FILLED,
	}
}

func validatePlacingOrderReq(userId, market string, req *dto.OrderReq) error {
	if userId == "" {
		return errors.New("user id is required")
	}
	if market == "" {
		return errors.New("market is required")
	}

	if req.Side == model.ASK {
		if req.Size <= 0 {
			return errors.New("ask order size must be greater than zero")
		}
	}

	if req.Side == model.BID {
		if req.OrderType == book.MARKET && req.QuoteAmount <= 0 {
			return errors.New("bid order quote amount must be greater than zero")
		}
	}

	if req.OrderType == book.LIMIT && (req.Price <= 0 || req.Size <= 0) {
		return errors.New("limit order price and size must be greater than zero")
	}
	return nil
}
