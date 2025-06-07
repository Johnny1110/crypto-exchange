package serviceImpl

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/security"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type userService struct {
	db              *sql.DB
	userRepo        repository.IUserRepository
	balanceRepo     repository.IBalanceRepository
	credentialCache *security.CredentialCache
}

func NewIUserService(db *sql.DB, userRepo repository.IUserRepository, balanceRepo repository.IBalanceRepository, credentialCache *security.CredentialCache) service.IUserService {
	return &userService{
		db:              db,
		userRepo:        userRepo,
		balanceRepo:     balanceRepo,
		credentialCache: credentialCache,
	}
}

func (s userService) GetUser(ctx context.Context, userId string) (*dto.User, error) {
	return s.userRepo.GetUserById(ctx, s.db, userId)
}

func (s userService) Register(ctx context.Context, req *dto.RegisterReq) (string, error) {
	// gen userId
	userID, err := genUIDSecure()
	if err != nil {
		log.Errorf("[Register] failed to generate user id: %v", err)
		return "", fmt.Errorf("failed to create user")
	}
	// gen hash pwd
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	err = WithTx(ctx, s.db, func(tx *sql.Tx) error {
		user, err := s.userRepo.GetUserByUsername(ctx, tx, req.Username)

		log.Infof("query username:[%s], got user: [%v], err:[%v]", req.Username, user, err)

		if err == nil {
			log.Warn("[Register] user with username already exists")
			return errors.New("username already exists")
		}

		err = s.userRepo.Insert(ctx, tx, &dto.User{
			ID:           userID,
			Username:     req.Username,
			PasswordHash: string(hash),
			VipLevel:     1,
			MakerFee:     0.001,
			TakerFee:     0.002,
		})

		err = s.balanceRepo.BatchCreate(ctx, tx, userID, settings.GetAllAssets())

		return err
	})

	if err != nil {
		return "", err
	}

	return userID, err
}

func (s userService) Login(ctx context.Context, req *dto.LoginReq) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, s.db, req.Username)
	if err != nil {
		return "", errors.New("username not exists")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := uuid.NewString()
	s.credentialCache.Put(token, user)
	return token, nil
}

func (s userService) Logout(ctx context.Context, token string) error {
	s.credentialCache.Delete(token)
	return nil
}

func genUIDSecure() (string, error) {

	now := time.Now()
	dateStr := now.Format("060102")

	// gen 4 letters
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomLetters := make([]byte, 4)
	letterBytes := make([]byte, 4)

	if _, err := rand.Read(letterBytes); err != nil {
		return "", err
	}

	for i := 0; i < 4; i++ {
		randomLetters[i] = letters[letterBytes[i]%26]
	}

	// gen 4 nums
	numberBytes := make([]byte, 2)
	if _, err := rand.Read(numberBytes); err != nil {
		return "", err
	}

	randomNumbers := (int(numberBytes[0])<<8+int(numberBytes[1]))%9999 + 1
	numberStr := fmt.Sprintf("%04d", randomNumbers)

	return fmt.Sprintf("UID%s%s%s", dateStr, string(randomLetters), numberStr), nil
}
