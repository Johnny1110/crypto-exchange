package repository

import (
	"context"
	"database/sql"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"time"
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
	UpdatePwd(ctx context.Context, db DBExecutor, user *dto.User) error
	UpdateVipLevel(ctx context.Context, db DBExecutor, user *dto.User) error
}

type IBalanceRepository interface {
	// GetBalancesByUserId get balance by userId
	GetBalancesByUserId(ctx context.Context, db DBExecutor, userId string) ([]*dto.Balance, error)
	// ModifyAvailableByUserIdAndAsset modify asset balance available amount if sign==true (+), sign==false (-), if available not enough return error.
	ModifyAvailableByUserIdAndAsset(ctx context.Context, db DBExecutor, userID, asset string, sign bool, amount float64) error
	// ModifyLockedByUserIdAndAsset modify asset balance locked amount if sign==true (+), sign==false (-), if locked not enough return error.
	ModifyLockedByUserIdAndAsset(ctx context.Context, db DBExecutor, userID, asset string, sign bool, amount float64) error
	// LockedByUserIdAndAsset lock user asset available amount (decrease) and add locked amount, if available not enough return error.
	LockedByUserIdAndAsset(ctx context.Context, db DBExecutor, userID, asset string, amount float64) error
	// UnlockedByUserIdAndAsset unlock user asset locked amount (decrease) and add available amount, if locked not enough return error.
	UnlockedByUserIdAndAsset(ctx context.Context, db DBExecutor, userID, asset string, amount float64) error
	// BatchCreate batch insert by userId and assets.html slice. available and locked default = 0.0
	BatchCreate(ctx context.Context, db DBExecutor, userId string, assets []string) error
	UpdateAsset(ctx context.Context, db DBExecutor, userId string, asset string, availableChanging float64, lockedChanging float64) error
}

type IOrderRepository interface {
	Insert(ctx context.Context, db DBExecutor, order *dto.Order) error
	Update(ctx context.Context, db DBExecutor, order *dto.Order) error
	GetOrderByOrderId(ctx context.Context, db DBExecutor, orderId string) (*dto.Order, error)
	GetOrdersByUserIdAndStatus(ctx context.Context, db DBExecutor, userId string, status model.OrderStatus) ([]*dto.Order, error)
	GetOrdersByUserIdAndStatuses(ctx context.Context, db DBExecutor, id string, statuses []model.OrderStatus) ([]*dto.Order, error)
	SyncTradeMatchingResult(ctx context.Context, db DBExecutor, orderId string, decreasingSize, dealtQuoteAmount float64, fees float64) error
	CancelOrder(ctx context.Context, db DBExecutor, orderId string, remainingSize float64) error
	UpdateOriginalSize(ctx context.Context, db DBExecutor, orderId string, originalSize float64) error
	GetOrdersByMarketAndStatuses(ctx context.Context, db DBExecutor, market string, statuses []model.OrderStatus) ([]*dto.Order, error)
	PaginationQuery(ctx context.Context, db DBExecutor, query *dto.GetOrdersQueryReq, statuses []model.OrderStatus, endTime time.Time) (*dto.PaginationResp[*dto.Order], error)
	GetOrdersByUserIdAndMarketAndStatuses(ctx context.Context, b DBExecutor, userId string, market string, statuses []model.OrderStatus) ([]*dto.Order, error)
	CountOpenOrders(ctx context.Context, db *sql.DB, marketName string) (int64, error)
}

type ITradeRepository interface {
	BatchInsert(ctx context.Context, db DBExecutor, trades []book.Trade) error
	GetMarketLatestPrice(ctx context.Context, db DBExecutor, market string) (float64, error)
	GetMarketPriceTimesAgo(ctx context.Context, db DBExecutor, market string, timeAgo time.Time) (float64, error)
	GetMarketVolumeByTimeRange(ctx context.Context, db DBExecutor, market string, startTime time.Time, endTime time.Time) (float64, error)
}
