package repositoryImpl

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/repository"
	"strings"
	"time"
)

type orderRepository struct {
}

func NewOrderRepository() repository.IOrderRepository {
	return &orderRepository{}
}

func (o orderRepository) Insert(ctx context.Context, db repository.DBExecutor, order *dto.Order) error {
	query := `INSERT INTO orders (
		id, user_id, market, side, price, original_size, remaining_size, 
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.Market,
		order.Side,
		order.Price,
		order.OriginalSize,
		order.RemainingSize,
		order.QuoteAmount,
		order.AvgDealtPrice,
		order.Type,
		order.Mode,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

func (o orderRepository) Update(ctx context.Context, db repository.DBExecutor, order *dto.Order) error {
	query := `UPDATE orders SET 
		remaining_size = ?, status = ?, updated_at = ?
		WHERE id = ?`

	result, err := db.ExecContext(ctx, query,
		order.RemainingSize,
		order.Status,
		time.Now(),
		order.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order with id %s not found", order.ID)
	}

	return nil
}

func (o orderRepository) GetOrderByOrderId(ctx context.Context, db repository.DBExecutor, orderId string) (*dto.Order, error) {
	query := `SELECT id, user_id, market, side, price, original_size, remaining_size, 
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at 
		FROM orders WHERE id = ?`

	var order dto.Order

	err := db.QueryRowContext(ctx, query, orderId).Scan(
		&order.ID,
		&order.UserID,
		&order.Market,
		&order.Side,
		&order.Price,
		&order.OriginalSize,
		&order.RemainingSize,
		&order.QuoteAmount,
		&order.AvgDealtPrice,
		&order.Type,
		&order.Mode,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order with id %s not found", orderId)
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	return &order, nil
}

func (o orderRepository) GetOrdersByUserIdAndStatus(ctx context.Context, db repository.DBExecutor, userId string, status model.OrderStatus) ([]*dto.Order, error) {
	query := `SELECT id, user_id, market, side, price, original_size, remaining_size, 
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at 
		FROM orders WHERE user_id = ? AND status = ? 
		ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, userId, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*dto.Order
	for rows.Next() {
		order := &dto.Order{}

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Market,
			&order.Side,
			&order.Price,
			&order.OriginalSize,
			&order.RemainingSize,
			&order.QuoteAmount,
			&order.AvgDealtPrice,
			&order.Type,
			&order.Mode,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return orders, nil
}

func (o orderRepository) GetOrdersByUserIdAndStatuses(ctx context.Context, db repository.DBExecutor, id string, statuses []model.OrderStatus) ([]*dto.Order, error) {
	if len(statuses) == 0 {
		return []*dto.Order{}, nil
	}

	// create IN prepare statement
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses)+1)
	args[0] = id // user_id

	for i, status := range statuses {
		placeholders[i] = "?"
		args[i+1] = string(status)
	}

	query := fmt.Sprintf(`SELECT id, user_id, market, side, price, original_size, remaining_size, 
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at 
		FROM orders WHERE user_id = ? AND status IN (%s) 
		ORDER BY created_at DESC`, strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*dto.Order
	for rows.Next() {
		order := &dto.Order{}

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Market,
			&order.Side,
			&order.Price,
			&order.OriginalSize,
			&order.RemainingSize,
			&order.QuoteAmount,
			&order.AvgDealtPrice,
			&order.Type,
			&order.Mode,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return orders, nil
}

func (o orderRepository) DecreaseRemainingSize(ctx context.Context, db repository.DBExecutor, orderId string, decreasingSize float64) error {
	query := `UPDATE orders SET 
		remaining_size = remaining_size-?, 
		status = CASE           
		    					WHEN original_size = 0 THEN ?
			                  	WHEN remaining_size - ? = 0 THEN ?
								WHEN remaining_size - ? < original_size THEN ?
								ELSE status END
                , updated_at = ?
		WHERE id = ?`

	result, err := db.ExecContext(ctx, query,
		decreasingSize,
		model.ORDER_STATUS_FILLED,
		decreasingSize,
		model.ORDER_STATUS_FILLED,
		decreasingSize,
		model.ORDER_STATUS_PARTIAL,
		time.Now(),
		orderId,
	)

	if err != nil {
		return fmt.Errorf("failed to DecreaseRemainingSize order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order with id %s not found", orderId)
	}
	return nil
}

func (o orderRepository) CancelOrder(ctx context.Context, db repository.DBExecutor, orderId string, remainingSize float64) error {
	query := `UPDATE orders SET 
		remaining_size = ?, status = ?, updated_at = ?
		WHERE id = ?`

	result, err := db.ExecContext(ctx, query,
		remainingSize,
		model.ORDER_STATUS_CANCELED,
		time.Now(),
		orderId,
	)

	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order with id %s not found", orderId)
	}

	return nil
}
