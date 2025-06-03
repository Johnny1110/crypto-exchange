package serviceImpl

import (
	"context"
	"database/sql"
	"errors"
	"github.com/johnny1110/crypto-exchange/dto"
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
	if req.Secret != "frizo" {
		return errors.New("secret invalid")
	}

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
	// TODO: make some testing maker
	return nil
}
