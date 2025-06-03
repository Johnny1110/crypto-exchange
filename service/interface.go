package service

import (
	"context"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
)

type IUserService interface {
	GetUser(ctx context.Context, userId string) (*dto.User, error)
	Register(ctx context.Context, req *dto.RegisterReq) (string, error)
	// Login return token
	Login(ctx context.Context, req *dto.LoginReq) (string, error)
	Logout(ctx context.Context, token string) error
}

type IOrderBookService interface {
	GetSnapshot(ctx context.Context, market string) (*book.BookSnapshot, error)
	GetLatestPrice(ctx context.Context, market string) (float64, error)
}

type IBalanceService interface {
	GetBalances(ctx context.Context, userId string) ([]*dto.Balance, error)
}

type IOrderService interface {
	PlaceOrder(ctx context.Context, market, userID string, req *dto.OrderReq) (*dto.PlaceOrderResult, error)
	QueryOrder(ctx context.Context, userId string, isOpenOrder bool) ([]*dto.Order, error)
	CancelOrder(ctx context.Context, market, userID, orderID string) (*dto.Order, error)
}

type IAdminService interface {
	Settlement(ctx context.Context, req dto.SettlementReq) error
}
