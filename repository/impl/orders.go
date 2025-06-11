package repositoryImpl

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/utils"
	"math"
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
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at, fee_asset, fee_rate, fees
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ,?)`

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
		order.FeeAsset,
		order.FeeRate,
		order.Fees,
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
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at, fee_asset, fee_rate, fees
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
		&order.FeeAsset,
		&order.FeeRate,
		&order.Fees,
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
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at, fee_asset, fee_rate, fees
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
			&order.FeeAsset,
			&order.FeeRate,
			&order.Fees,
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
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at, fee_asset, fee_rate, fees
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
			&order.FeeAsset,
			&order.FeeRate,
			&order.Fees,
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

func (o orderRepository) SyncTradeMatchingResult(ctx context.Context, db repository.DBExecutor, orderId string, decreasingSize, dealtQuoteAmount float64, fees float64) error {
	query := `UPDATE orders SET 
		remaining_size = remaining_size - ?, 
		quote_amount = quote_amount + ?,
		avg_dealt_price = (quote_amount + ?) / (original_size - remaining_size + ?),
		status = CASE           
		    					WHEN original_size = 0 THEN ?
			                  	WHEN remaining_size - ? < ? THEN ?
								WHEN remaining_size - ? < original_size THEN ?
								ELSE status END
                   , fees = fees + ?
                   , updated_at = ?
		WHERE id = ?`

	result, err := db.ExecContext(ctx, query,
		decreasingSize,
		dealtQuoteAmount,
		dealtQuoteAmount,
		decreasingSize,
		model.ORDER_STATUS_FILLED,
		decreasingSize,
		utils.Scale,
		model.ORDER_STATUS_FILLED,
		decreasingSize,
		model.ORDER_STATUS_PARTIAL,
		fees,
		time.Now(),
		orderId,
	)

	if err != nil {
		return fmt.Errorf("failed to SyncTradeMatchingResult order: %w", err)
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

func (o orderRepository) UpdateOriginalSize(ctx context.Context, db repository.DBExecutor, orderId string, originalSize float64) error {
	query := `UPDATE orders SET 
		original_size = ?, updated_at = ?
		WHERE id = ?`

	result, err := db.ExecContext(ctx, query,
		originalSize,
		time.Now(),
		orderId,
	)

	if err != nil {
		return fmt.Errorf("failed to update order original size: %w", err)
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

func (o orderRepository) GetOrdersByMarketAndStatuses(ctx context.Context, db repository.DBExecutor, market string, statuses []model.OrderStatus) ([]*dto.Order, error) {
	if len(statuses) == 0 {
		return []*dto.Order{}, nil
	}

	// create IN prepare statement
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses)+1)
	args[0] = market

	for i, status := range statuses {
		placeholders[i] = "?"
		args[i+1] = string(status)
	}

	query := fmt.Sprintf(`SELECT id, user_id, market, side, price, original_size, remaining_size, 
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at, fee_asset, fee_rate, fees
		FROM orders WHERE market = ? AND status IN (%s) 
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
			&order.FeeAsset,
			&order.FeeRate,
			&order.Fees,
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

func (o orderRepository) GetOrdersByUserIdAndMarketAndStatuses(ctx context.Context, db repository.DBExecutor, userId string, market string, statuses []model.OrderStatus) ([]*dto.Order, error) {
	if len(statuses) == 0 {
		return []*dto.Order{}, nil
	}

	// create IN prepare statement
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses)+2)
	args[0] = userId
	args[1] = market

	for i, status := range statuses {
		placeholders[i] = "?"
		args[i+2] = string(status)
	}

	query := fmt.Sprintf(`SELECT id, user_id, market, side, price, original_size, remaining_size, 
		quote_amount, avg_dealt_price, type, mode, status, created_at, updated_at, fee_asset, fee_rate, fees
		FROM orders WHERE user_id = ? AND market = ? AND status IN (%s) 
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
			&order.FeeAsset,
			&order.FeeRate,
			&order.Fees,
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

func (o orderRepository) PaginationQuery(ctx context.Context, db repository.DBExecutor, query *dto.GetOrdersQueryReq, statuses []model.OrderStatus, endTime time.Time) (*dto.PaginationResp[*dto.Order], error) {
	if query == nil || len(statuses) == 0 {
		return nil, fmt.Errorf("invalid query parameters")
	}

	// Build WHERE conditions
	var conditions []string
	var args []interface{}

	// user_id is required
	conditions = append(conditions, "user_id = ?")
	args = append(args, query.UserID)

	// status filter (multiple values)
	statusPlaceholders := make([]string, len(statuses))
	for i, status := range statuses {
		statusPlaceholders[i] = "?"
		args = append(args, string(status))
	}
	conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusPlaceholders, ",")))

	// market filter (optional)
	if query.Market != "" {
		conditions = append(conditions, "market = ?")
		args = append(args, query.Market)
	}

	// side filter (optional)
	if query.Side != 0 { // assuming 0 is not a valid side value
		conditions = append(conditions, "side = ?")
		args = append(args, int(query.Side))
	}

	// time filter for closed orders (if endTime is set)
	if !endTime.IsZero() {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, endTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total records
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM orders WHERE %s", whereClause)
	var total int64
	if err := db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Calculate pagination
	offset := (query.CurrentPage - 1) * query.PageSize
	totalPages := int64(math.Ceil(float64(total) / float64(query.PageSize)))

	// Query data with pagination
	dataSQL := fmt.Sprintf(`
        SELECT id, user_id, market, side, price, original_size, remaining_size, 
               quote_amount, avg_dealt_price, type, mode, status, fee_rate, 
               fees, fee_asset, created_at, updated_at 
        FROM orders 
        WHERE %s 
        ORDER BY created_at DESC 
        LIMIT ? OFFSET ?`, whereClause)

	args = append(args, query.PageSize, offset)

	rows, err := db.QueryContext(ctx, dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*dto.Order
	for rows.Next() {
		var order dto.Order
		var feeAsset sql.NullString

		err = rows.Scan(
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
			&order.FeeRate,
			&order.Fees,
			&feeAsset,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Handle nullable fee_asset
		if feeAsset.Valid {
			order.FeeAsset = feeAsset.String
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Build response
	response := &dto.PaginationResp[*dto.Order]{
		Result:      orders,
		Total:       total,
		CurrentPage: query.CurrentPage,
		PageSize:    query.PageSize,
		TotalPages:  totalPages,
		HasNext:     query.CurrentPage < totalPages,
		HasPrev:     query.CurrentPage > 1,
	}

	return response, nil
}

// CountOpenOrders open orders status in  ('NEW', 'PARTIAL')
func (o orderRepository) CountOpenOrders(ctx context.Context, db *sql.DB, marketName string) (int64, error) {
	query := `
		SELECT COUNT(id) 
		FROM orders 
		WHERE market = ? 
		  AND status IN ('NEW', 'PARTIAL')
	`

	var count int64
	err := db.QueryRowContext(ctx, query, marketName).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count open orders for market %s: %w", marketName, err)
	}

	return count, nil
}
