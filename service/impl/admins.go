package serviceImpl

import (
	"context"
	"database/sql"
	"errors"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
)

type adminService struct {
	db           *sql.DB
	userRepo     repository.IUserRepository
	balanceRepo  repository.IBalanceRepository
	orderService service.IOrderService
}

func NewIAdminService(db *sql.DB,
	userRepo repository.IUserRepository,
	balanceRepo repository.IBalanceRepository,
	orderService service.IOrderService) service.IAdminService {
	return &adminService{
		db:           db,
		userRepo:     userRepo,
		balanceRepo:  balanceRepo,
		orderService: orderService,
	}
}

func (as adminService) Settlement(ctx context.Context, req dto.SettlementReq) error {
	err := WithTx(ctx, as.db, func(tx *sql.Tx) error {
		user, err := as.userRepo.GetUserByUsername(ctx, tx, req.Username)
		if err != nil {
			return err
		}
		if user == nil {
			return errors.New("user not found by username")
		}

		err = as.balanceRepo.ModifyAvailableByUserIdAndAsset(ctx, tx, user.ID, req.Asset, true, req.Amount)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (as adminService) TestAutoMakeMarket(ctx context.Context) error {
	// make some testing maker
	user := &dto.User{
		Username: "market_maker",
		ID:       "1",
		VipLevel: 7,
		MakerFee: 0.0001,
		TakerFee: 0.002,
	}
	market := "ETH-USDT"
	// make 5 bid orders
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3000,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2900,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2800,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2700,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2600,
		Size:      10,
	})

	// make 5 ask orders
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3500,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3400,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3300,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3200,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3100,
		Size:      10,
	})

	return nil
}
