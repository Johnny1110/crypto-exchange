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

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderNotBelongsToUser = errors.New("order not belongs to user")
	ErrInsufficientBalance   = errors.New("insufficient balance")
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

// PlaceOrderContext encapsulates all order placement context
type PlaceOrderContext struct {
	Market   string
	UserID   string
	Request  *dto.OrderReq
	OrderDTO *dto.Order
	Assets   *AssetDetails
	Trades   []book.Trade
}

func (c *PlaceOrderContext) syncTradeResult(engineOrder *model.Order, trades []book.Trade) {
	c.OrderDTO.RemainingSize = engineOrder.RemainingSize
	c.OrderDTO.Status = engineOrder.GetStatus()
	c.Trades = trades
}

// AssetDetails holds asset-related information
type AssetDetails struct {
	BaseAsset   string
	QuoteAsset  string
	FreezeAsset string
	FreezeAmt   float64
}

func (s *orderService) PlaceOrder(ctx context.Context, market, userID string, req *dto.OrderReq) (*dto.PlaceOrderResult, error) {
	// Initialize order context
	orderCtx, err := s.initializeOrderContext(market, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize order context: %w", err)
	}

	// Execute order placement strategy
	strategy, err := s.getOrderPlacementStrategy(req.OrderType)
	if err != nil {
		return nil, fmt.Errorf("failed to get order placement strategy: %w", err)
	}
	if err := strategy.Execute(ctx, s, orderCtx); err != nil {
		return nil, fmt.Errorf("failed to execute order placement: %w", err)
	}

	return serviceHelper.WrapPlaceOrderResult(orderCtx.OrderDTO, orderCtx.Trades), nil
}

func (s *orderService) initializeOrderContext(market, userID string, req *dto.OrderReq) (*PlaceOrderContext, error) {
	if err := validatePlacingOrderReq(userID, market, req); err != nil {
		return nil, err
	}

	baseAsset, quoteAsset, err := serviceHelper.ParseMarket(s.engine, market)
	if err != nil {
		return nil, fmt.Errorf("failed to parse market: %w", err)
	}

	freezeAsset, freezeAmt := serviceHelper.DetermineFreezeValue(req, baseAsset, quoteAsset)

	return &PlaceOrderContext{
		Market:  market,
		UserID:  userID,
		Request: req,
		Assets: &AssetDetails{
			BaseAsset:   baseAsset,
			QuoteAsset:  quoteAsset,
			FreezeAsset: freezeAsset,
			FreezeAmt:   freezeAmt,
		},
	}, nil
}

// OrderPlacementStrategy >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// OrderPlacementStrategy defines the interface for order placement strategies
type OrderPlacementStrategy interface {
	Execute(ctx context.Context, service *orderService, orderCtx *PlaceOrderContext) error
}

// LimitOrderStrategy implements limit order placement logic
type LimitOrderStrategy struct{}

func (s *LimitOrderStrategy) Execute(ctx context.Context, service *orderService, orderCtx *PlaceOrderContext) error {
	orderCtx.OrderDTO = serviceHelper.NewLimitOrderDtoByOrderReq(orderCtx.Market, orderCtx.UserID, orderCtx.Request)
	return service.executeOrderPlacement(ctx, orderCtx, false)
}

// MarketOrderStrategy implements market order placement logic
type MarketOrderStrategy struct{}

func (s *MarketOrderStrategy) Execute(ctx context.Context, service *orderService, orderCtx *PlaceOrderContext) error {
	orderCtx.OrderDTO = serviceHelper.NewMarketOrderDtoByOrderReq(orderCtx.Market, orderCtx.UserID, orderCtx.Request)
	return service.executeOrderPlacement(ctx, orderCtx, true)
}

func (s *orderService) getOrderPlacementStrategy(orderType book.OrderType) (OrderPlacementStrategy, error) {
	switch orderType {
	case book.LIMIT:
		return &LimitOrderStrategy{}, nil
	case book.MARKET:
		return &MarketOrderStrategy{}, nil
	default:
		log.Errorf("[getOrderPlacementStrategy] failed, unknown order type: %v", orderType)
		return nil, ErrInvalidInput
	}
}

// OrderPlacementStrategy <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func (s *orderService) executeOrderPlacement(ctx context.Context, orderCtx *PlaceOrderContext, isMarketOrder bool) error {
	// Phase 1: Process order placement
	if err := s.executeOrderPlacementPhase(ctx, orderCtx, isMarketOrder); err != nil {
		return fmt.Errorf("order placement phase failed: %w", err)
	}

	log.Infof("[executeOrderPlacement] Phase-1 done: %v", orderCtx)

	// Phase 2: Process trade settlement
	if err := s.executeTradeSettlementPhase(ctx, orderCtx); err != nil {
		return fmt.Errorf("trade settlement phase failed: %w", err)
	}

	return nil
}

func (s *orderService) executeOrderPlacementPhase(ctx context.Context, orderCtx *PlaceOrderContext, isMarketOrder bool) error {
	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
		// 1. Freeze user funds
		if err := s.balanceRepo.LockedByUserIdAndAsset(ctx, tx, orderCtx.UserID, orderCtx.Assets.FreezeAsset, orderCtx.Assets.FreezeAmt); err != nil {
			log.Warnf("[OrderPlacement] failed to lock user balance, %v", err)
			return ErrInsufficientBalance
		}

		// 2. Insert order to database
		if err := s.orderRepo.Insert(ctx, tx, orderCtx.OrderDTO); err != nil {
			return fmt.Errorf("failed to insert order: %w", err)
		}

		// 3. Place order in matching engine
		engineOrder := serviceHelper.NewEngineOrderByOrderDto(orderCtx.OrderDTO)
		trades, err := s.engine.PlaceOrder(orderCtx.Market, orderCtx.Request.OrderType, engineOrder)
		if err != nil {
			return fmt.Errorf("failed to place order in engine: %w", err)
		}

		// Update order status from engine result
		orderCtx.syncTradeResult(engineOrder, trades)

		// 4. Save trade records
		if len(orderCtx.Trades) > 0 {
			if err := s.tradeRepo.BatchInsert(ctx, tx, trades); err != nil {
				return fmt.Errorf("failed to insert trades: %w", err)
			}
		}

		// 5. Handle market bid order special case
		if isMarketOrder && orderCtx.Request.Side == model.BID {
			if err := s.orderRepo.UpdateOriginalSize(ctx, tx, engineOrder.ID, engineOrder.OriginalSize); err != nil {
				return fmt.Errorf("failed to update original size: %w", err)
			}
			orderCtx.OrderDTO.OriginalSize = engineOrder.OriginalSize
		}

		return nil
	})
}

func (s *orderService) executeTradeSettlementPhase(ctx context.Context, orderCtx *PlaceOrderContext) error {
	if len(orderCtx.Trades) == 0 {
		return nil // No trades to settle
	}

	settlementResult, err := serviceHelper.ProcessTradeSettlement(orderCtx.OrderDTO, orderCtx.Trades)
	if err != nil {
		return fmt.Errorf("failed to process trade settlement: %w", err)
	}

	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
		// Update orders
		for _, orderUpdate := range settlementResult.OrderUpdates {
			if err := s.orderRepo.SyncTradeMatchingResult(ctx, tx, orderUpdate.OrderID, orderUpdate.RemainingSizeDecreasing, orderUpdate.DealtQuoteAmountIncreasing); err != nil {
				return fmt.Errorf("failed to sync trade matching result for order %s: %w", orderUpdate.OrderID, err)
			}
		}

		// Update user balances
		for userID, settlement := range settlementResult.UserSettlements {
			if err := s.updateUserAssets(ctx, tx, userID, orderCtx.Assets, settlement); err != nil {
				return fmt.Errorf("failed to update assets for user %s: %w", userID, err)
			}
		}

		return nil
	})
}

// updateUserAssets Update user base and quote assets.
func (s *orderService) updateUserAssets(ctx context.Context, tx *sql.Tx, userID string, assets *AssetDetails, settlement *serviceHelper.UserSettlementData) error {
	// update BASE asset for user.
	if err := s.balanceRepo.UpdateAsset(ctx, tx, userID, assets.BaseAsset, settlement.BaseAssetAvailable, settlement.BaseAssetLocked); err != nil {
		return fmt.Errorf("failed to update base asset: %w", err)
	}
	// update Quote asset for user.
	if err := s.balanceRepo.UpdateAsset(ctx, tx, userID, assets.QuoteAsset, settlement.QuoteAssetAvailable, settlement.QuoteAssetLocked); err != nil {
		return fmt.Errorf("failed to update quote asset: %w", err)
	}

	return nil
}

func (s *orderService) CancelOrder(ctx context.Context, market, userID, orderID string) (*dto.Order, error) {
	if market == "" || userID == "" || orderID == "" {
		return nil, ErrInvalidInput
	}

	orderDto, err := s.orderRepo.GetOrderByOrderId(ctx, s.db, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if orderDto.UserID != userID {
		return nil, ErrOrderNotBelongsToUser
	}

	engineOrder, err := s.engine.CancelOrder(market, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel order in engine: %w", err)
	}

	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		// Update order status
		orderDto.RemainingSize = engineOrder.RemainingSize
		orderDto.Status = model.ORDER_STATUS_CANCELED

		if err := s.orderRepo.Update(ctx, tx, orderDto); err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		// Calculate and process refund
		unlockAsset, unlockAmount, err := serviceHelper.CalculateRefund(s.engine, market, engineOrder)
		if err != nil {
			return fmt.Errorf("failed to calculate refund: %w", err)
		}

		if unlockAmount > 0 {
			if err := s.balanceRepo.UnlockedByUserIdAndAsset(ctx, tx, userID, unlockAsset, unlockAmount); err != nil {
				return fmt.Errorf("failed to unlock balance: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to cancel order transaction: %w", err)
	}

	return orderDto, nil
}

func (s *orderService) QueryOrder(ctx context.Context, userID string, isOpenOrder bool) ([]*dto.Order, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	statuses := getOrderStatusesByOpenFlag(isOpenOrder)
	orders, err := s.orderRepo.GetOrdersByUserIdAndStatuses(ctx, s.db, userID, statuses)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}

	return orders, nil
}

func getOrderStatusesByOpenFlag(isOpen bool) []model.OrderStatus {
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

func validatePlacingOrderReq(userID, market string, req *dto.OrderReq) error {
	switch {
	case userID == "":
		return errors.New("user id is required")
	case market == "":
		return errors.New("market is required")
	case req == nil:
		return errors.New("order request is required")
	}

	// Validate Ask orders
	if req.Side == model.ASK && req.Size <= 0 {
		return errors.New("ask order size must be greater than zero")
	}

	// Validate Bid orders
	if req.Side == model.BID {
		if req.OrderType == book.MARKET && req.QuoteAmount <= 0 {
			return errors.New("bid market order quote amount must be greater than zero")
		}
		if req.OrderType == book.LIMIT && req.Size <= 0 {
			return errors.New("bid limit order size must be greater than zero")
		}
	}

	// Validate Limit orders
	if req.OrderType == book.LIMIT && req.Price <= 0 {
		return errors.New("limit order price must be greater than zero")
	}

	return nil
}
