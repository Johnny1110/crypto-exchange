package service

import (
	"context"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/ohlcv"
)

type ICacheService interface {
	Update(key string, data interface{})
	Get(key string) (interface{}, bool)
}

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
	GetBaseQuoteAssets(ctx context.Context, market string) (string, string, error)
}

type IBalanceService interface {
	GetBalances(ctx context.Context, userId string) ([]*dto.Balance, error)
}

type IOrderService interface {
	PlaceOrder(ctx context.Context, market string, user *dto.User, req *dto.OrderReq) (*dto.PlaceOrderResult, error)
	QueryOrder(ctx context.Context, userId string, isOpenOrder bool) ([]*dto.Order, error)
	CancelOrder(ctx context.Context, userID, orderID string) (*dto.Order, error)
	QueryOrdersByMarketAndStatuses(ctx context.Context, market string, statuses []model.OrderStatus) ([]*dto.Order, error)
	PaginationQuery(ctx context.Context, query *dto.GetOrdersQueryReq) (*dto.PaginationResp[*dto.Order], error)
	QueryOrderByMarket(ctx context.Context, userID string, market string, isOpenOrder bool) ([]*dto.Order, error)
	CountOpenOrders(ctx context.Context, marketName string) (int64, error)
}

type IAdminService interface {
	Settlement(ctx context.Context, req dto.SettlementReq) error
	TestAutoMakeMarket(ctx context.Context) error
}

// Auto Market Maker (AMM) etc. >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type IAutoMarketMakerService interface {
	BootUp(ctx context.Context, markets []market.MarketInfo)
}

type IPriceIndexService interface {
}

// Markets
type IMarketDataService interface {
	CalculateMarketData(ctx context.Context, market string) (*dto.MarketData, error)
	GetAllMarketData() ([]dto.MarketData, error)
	GetMarketData(market string) (dto.MarketData, error)
	GetOHLCVHistory(ctx context.Context, req *ohlcv.GetOhlcvDataReq) (*ohlcv.OHLCV, error)
}
