package service

import (
	"context"
	"database/sql"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
)

type balanceService struct {
	db          *sql.DB
	userRepo    repository.IUserRepository
	balanceRepo repository.IBalanceRepository
}

func NewIBalanceService(db *sql.DB,
	userRepo repository.IUserRepository,
	balanceRepo repository.IBalanceRepository) service.IBalanceService {
	return &balanceService{
		db:          db,
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
	}
}

func (bs *balanceService) GetBalances(ctx context.Context, userId string) ([]*dto.Balance, error) {
	return bs.balanceRepo.GetBalancesByUserId(ctx, bs.db, userId)
}
