package service

import (
	"context"
	"database/sql"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
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

func (s *orderService) PlaceOrder(ctx context.Context, req *dto.OrderReq) (*dto.PlaceOrderResult, error) {
	//TODO implement me
	panic("implement me")
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
