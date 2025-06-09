package repositoryImpl

import (
	"context"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/repository"
	"strings"
)

type balanceRepository struct {
}

func NewBalanceRepository() repository.IBalanceRepository {
	return &balanceRepository{}
}

// GetBalancesByUserId get balance by userId
func (b balanceRepository) GetBalancesByUserId(ctx context.Context, db repository.DBExecutor, userId string) ([]*dto.Balance, error) {
	query := `SELECT asset, available, locked FROM balances WHERE user_id = ?`

	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}
	defer rows.Close()

	var balances []*dto.Balance
	for rows.Next() {
		balance := &dto.Balance{}
		err := rows.Scan(&balance.Asset, &balance.Available, &balance.Locked)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance: %w", err)
		}
		balance.Total = balance.Available + balance.Locked
		balances = append(balances, balance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return balances, nil

}

// ModifyAvailableByUserIdAndAsset modify asset balance available amount if sign==true (+), sign==false (-), if available not enough return error.
func (b balanceRepository) ModifyAvailableByUserIdAndAsset(ctx context.Context, db repository.DBExecutor, userID, asset string, sign bool, amount float64) error {
	if sign {
		// add available amt.
		query := `UPDATE balances SET available = available + ? WHERE user_id = ? AND asset = ?`
		result, err := db.ExecContext(ctx, query, amount, userID, asset)
		if err != nil {
			return fmt.Errorf("failed to increase available balance: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("balance not found for user %s and asset %s", userID, asset)
		}
	} else {
		// decrease available
		query := `UPDATE balances SET available = available - ? WHERE user_id = ? AND asset = ? AND available >= ?`
		result, err := db.ExecContext(ctx, query, amount, userID, asset, amount)
		if err != nil {
			return fmt.Errorf("failed to decrease available balance: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("insufficient available balance or balance not found for user %s and asset %s", userID, asset)
		}
	}

	return nil
}

// ModifyLockedByUserIdAndAsset modify asset balance locked amount if sign==true (+), sign==false (-), if locked not enough return error.
func (b balanceRepository) ModifyLockedByUserIdAndAsset(ctx context.Context, db repository.DBExecutor, userID, asset string, sign bool, amount float64) error {
	if sign {
		query := `UPDATE balances SET locked = locked + ? WHERE user_id = ? AND asset = ?`
		result, err := db.ExecContext(ctx, query, amount, userID, asset)
		if err != nil {
			return fmt.Errorf("failed to increase locked balance: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("balance not found for user %s and asset %s", userID, asset)
		}
	} else {
		query := `UPDATE balances SET locked = locked - ? WHERE user_id = ? AND asset = ? AND locked >= ?`
		result, err := db.ExecContext(ctx, query, amount, userID, asset, amount)
		if err != nil {
			return fmt.Errorf("failed to decrease locked balance: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("insufficient locked balance or balance not found for user %s and asset %s", userID, asset)
		}
	}

	return nil
}

// LockedByUserIdAndAsset lock user asset available amount (decrease) and add locked amount, if available not enough return error.
func (b balanceRepository) LockedByUserIdAndAsset(ctx context.Context, db repository.DBExecutor, userID, asset string, amount float64) error {
	// atomic update：decrease available and increase locked
	query := `UPDATE balances 
			  SET available = MAX(0, available - ?), locked = locked + ? 
			  WHERE user_id = ? AND asset = ? AND available >= ?`

	result, err := db.ExecContext(ctx, query, amount, amount, userID, asset, amount)
	if err != nil {
		return fmt.Errorf("failed to lock balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient available balance, asset %s", asset)
	}

	return nil
}

// UnlockedByUserIdAndAsset unlock user asset locked amount (decrease) and add available amount, if locked not enough return error.
func (b balanceRepository) UnlockedByUserIdAndAsset(ctx context.Context, db repository.DBExecutor, userID, asset string, amount float64) error {
	// atomic update access
	query := `UPDATE balances 
			  SET locked = MAX(0, locked - ?), available = available + ? 
			  WHERE user_id = ? AND asset = ?`

	result, err := db.ExecContext(ctx, query, amount, amount, userID, asset)
	if err != nil {
		return fmt.Errorf("failed to unlock balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insufficient locked balance or balance not found for user %s and asset %s", userID, asset)
	}

	return nil
}

// BatchCreate batch insert by userId and assets.html slice. available and locked default = 0.0
func (b balanceRepository) BatchCreate(ctx context.Context, db repository.DBExecutor, userId string, assets []string) error {
	if len(assets) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(assets))
	valueArgs := make([]interface{}, 0, len(assets)*4)

	for _, asset := range assets {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs, userId, asset, 0.0, 0.0)
	}

	query := fmt.Sprintf("INSERT INTO balances (user_id, asset, available, locked) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch create balances: %w", err)
	}

	return nil
}

func (b balanceRepository) UpdateAsset(ctx context.Context, db repository.DBExecutor, userId string, asset string, availableChanging float64, lockedChanging float64) error {
	query := `UPDATE balances 
			  SET available = available + ?, locked = locked + ? 
			  WHERE user_id = ? AND asset = ?`

	result, err := db.ExecContext(ctx, query, availableChanging, lockedChanging, userId, asset)
	if err != nil {
		return fmt.Errorf("failed to UpdateAsset balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get UpdateAsset affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("userId %s and asset %s not found", userId, asset)
	}

	return nil
}
