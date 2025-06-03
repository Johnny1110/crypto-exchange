package repositoryImpl

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/repository"
)

type userRepository struct {
}

func NewUserRepository() repository.IUserRepository {
	return &userRepository{}
}

func (u userRepository) GetUserById(ctx context.Context, db repository.DBExecutor, userId string) (*dto.User, error) {
	query := `SELECT id, username, password_hash FROM users WHERE id = ?`

	var user dto.User

	err := db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %s not found", userId)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (u userRepository) GetUserByUsername(ctx context.Context, db repository.DBExecutor, username string) (*dto.User, error) {
	query := `SELECT id, username, password_hash FROM users WHERE username = ?`

	var user dto.User

	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with username %s not found", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (u userRepository) Insert(ctx context.Context, db repository.DBExecutor, user *dto.User) error {
	query := `INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)`

	_, err := db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (u userRepository) Update(ctx context.Context, db repository.DBExecutor, user *dto.User) error {
	query := `UPDATE users SET username = ?, password_hash = ? WHERE id = ?`

	result, err := db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %s not found", user.ID)
	}

	return nil
}
