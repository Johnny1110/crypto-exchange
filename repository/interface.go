package repository

import (
	"context"
	"database/sql"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type IUserRepository interface {
	GetUserById(ctx context.Context, db DBExecutor, userId string) (*dto.User, error)
	GetUserByUsername(ctx context.Context, db DBExecutor, username string) (*dto.User, error)
	Insert(ctx context.Context, db DBExecutor, user *dto.User) error
	Update(ctx context.Context, db DBExecutor, user *dto.User) error
}

type IBalanceRepository interface {
	GetBalancesByUserId(ctx context.Context, db DBExecutor, userId string) ([]*dto.Balance, error)
	ModifyAvailableByUserIdAndAsset(ctx context.Context, db DBExecutor, userID, asset string, sign bool, amount float64) error
	ModifyLockedByUserIdAndAsset(ctx context.Context, db DBExecutor, userID, asset string, sign bool, amount float64) error
	BatchCreate(ctx context.Context, db DBExecutor, userId string, assets []string) error
}

type IOrderRepository interface {
	Insert(ctx context.Context, db DBExecutor, order *dto.Order) error
	Update(ctx context.Context, db DBExecutor, order *dto.Order) error
	GetOrderByOrderId(ctx context.Context, db DBExecutor, orderId string) (*dto.Order, error)
	GetOrdersByUserIdAndStatus(ctx context.Context, db DBExecutor, userId string, status model.OrderStatus) ([]*dto.Order, error)
	GetOrdersByUserIdAndStatuses(ctx context.Context, db *sql.DB, id string, statuses []model.OrderStatus) ([]*dto.Order, error)
}

type ITradeRepository interface {
	BatchInsert(ctx context.Context, db DBExecutor, trades []*book.Trade) error
}
