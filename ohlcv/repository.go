package ohlcv

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3"
)

type SQLiteOHLCVRepository struct {
	db *sql.DB
}

func NewSQLiteOHLCVRepository(db *sql.DB) *SQLiteOHLCVRepository {
	return &SQLiteOHLCVRepository{
		db: db,
	}
}

// SaveOHLCVBar
func (r *SQLiteOHLCVRepository) SaveOHLCVBar(ctx context.Context, bar *OHLCVBar, interval OHLCV_INTERVAL) error {
	if bar == nil {
		return fmt.Errorf("bar cannot be nil")
	}

	tableName, err := r.getTableName(interval)
	if err != nil {
		return fmt.Errorf("unsupported interval: %s", interval)
	}

	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s 
		(symbol, open_price, high_price, low_price, close_price, volume, quote_volume, 
		 open_time, close_time, trade_count, is_closed, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, tableName)

	_, err = r.db.ExecContext(ctx, query,
		bar.Symbol,
		bar.OpenPrice,
		bar.HighPrice,
		bar.LowPrice,
		bar.ClosePrice,
		bar.Volume,
		bar.QuoteVolume,
		bar.OpenTime,
		bar.CloseTime,
		bar.TradeCount,
		r.boolToInt(bar.IsClosed),
	)

	return err
}

// GetOHLCVData
func (r *SQLiteOHLCVRepository) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	tableName, err := r.getTableName(req.Interval)
	if err != nil {
		return nil, fmt.Errorf("unsupported interval: %s", req.Interval)
	}

	whereConditions := []string{"symbol = ?"}
	args := []interface{}{req.Symbol}

	if !req.StartTime.IsZero() {
		whereConditions = append(whereConditions, "open_time >= ?")
		args = append(args, req.StartTime.Unix())
	}

	if !req.EndTime.IsZero() {
		whereConditions = append(whereConditions, "close_time <= ?")
		args = append(args, req.EndTime.Unix())
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 500
	}
	if limit > 1000 {
		limit = 1000
	}

	query := fmt.Sprintf(`
		SELECT open_time, open_price, high_price, low_price, close_price, volume
		FROM %s 
		WHERE %s AND is_closed = 1
		ORDER BY open_time ASC
		LIMIT ?
	`, tableName, strings.Join(whereConditions, " AND "))

	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ohlcv := &OHLCV{
		S: "ok",
		T: make([]int64, 0),
		O: make([]float64, 0),
		H: make([]float64, 0),
		L: make([]float64, 0),
		C: make([]float64, 0),
		V: make([]float64, 0),
	}

	for rows.Next() {
		var timestamp int64
		var open, high, low, closed, volume float64

		err := rows.Scan(&timestamp, &open, &high, &low, &closed, &volume)
		if err != nil {
			return nil, err
		}

		ohlcv.T = append(ohlcv.T, timestamp)
		ohlcv.O = append(ohlcv.O, open)
		ohlcv.H = append(ohlcv.H, high)
		ohlcv.L = append(ohlcv.L, low)
		ohlcv.C = append(ohlcv.C, closed)
		ohlcv.V = append(ohlcv.V, volume)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ohlcv, nil
}

// UpdateRealtimeOHLCV
func (r *SQLiteOHLCVRepository) UpdateRealtimeOHLCV(ctx context.Context, bar OHLCVBar, interval OHLCV_INTERVAL) error {
	query := `
		INSERT OR REPLACE INTO ohlcv_realtime 
		(symbol, interval_type, open_price, high_price, low_price, close_price, 
		 volume, quote_volume, open_time, close_time, trade_count, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query,
		bar.Symbol,
		string(interval),
		bar.OpenPrice,
		bar.HighPrice,
		bar.LowPrice,
		bar.ClosePrice,
		bar.Volume,
		bar.QuoteVolume,
		bar.OpenTime,
		bar.CloseTime,
		bar.TradeCount,
	)

	return err
}

// GetRealtimeOHLCV
func (r *SQLiteOHLCVRepository) GetRealtimeOHLCV(ctx context.Context, symbol, interval OHLCV_INTERVAL, openTime int64) (*OHLCVBar, error) {
	query := `
		SELECT symbol, open_price, high_price, low_price, close_price, 
		       volume, quote_volume, open_time, close_time, trade_count
		FROM ohlcv_realtime 
		WHERE symbol = ? AND interval_type = ? AND open_time = ?
	`

	row := r.db.QueryRowContext(ctx, query, symbol, interval, openTime)

	bar := &OHLCVBar{}
	err := row.Scan(
		&bar.Symbol,
		&bar.OpenPrice,
		&bar.HighPrice,
		&bar.LowPrice,
		&bar.ClosePrice,
		&bar.Volume,
		&bar.QuoteVolume,
		&bar.OpenTime,
		&bar.CloseTime,
		&bar.TradeCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// setup Duration (based on interval)
	duration, err := r.parseDuration(interval)
	if err != nil {
		return nil, err
	}
	bar.Duration = duration
	bar.IsClosed = false

	return bar, nil
}

// UpdateStatistics
func (r *SQLiteOHLCVRepository) UpdateStatistics(ctx context.Context, symbol, interval OHLCV_INTERVAL, date time.Time, stats *ohlcvStatistics) error {
	if stats == nil {
		return fmt.Errorf("stats cannot be nil")
	}

	dateKey := date.Format("2006-01-02")

	query := `
		INSERT OR REPLACE INTO ohlcv_statistics 
		(symbol, interval_type, date_key, record_count, min_open_time, max_close_time, 
		 avg_volume, total_volume, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query,
		symbol,
		interval,
		dateKey,
		stats.RecordCount,
		stats.MinOpenTime,
		stats.MaxCloseTime,
		stats.AvgVolume,
		stats.TotalVolume,
	)

	return err
}

// UpsertOHLCVBars Batch insert or update OHLCV data
// if exists same (symbol, open_time)，accumulate data; insert otherwise.
func (r *SQLiteOHLCVRepository) UpsertOHLCVBars(ctx context.Context, ohlcvBars []OHLCVBar, interval OHLCV_INTERVAL) error {
	if len(ohlcvBars) == 0 {
		return nil
	}

	tableName, err := r.getTableName(interval)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	selectQuery := fmt.Sprintf(`
		SELECT open_price, high_price, low_price, close_price, volume, quote_volume, trade_count
		FROM %s 
		WHERE symbol = ? AND open_time = ?
	`, tableName)

	selectStmt, err := tx.PrepareContext(ctx, selectQuery)
	if err != nil {
		return err
	}
	defer selectStmt.Close()

	insertQuery := fmt.Sprintf(`
		INSERT INTO %s 
		(symbol, open_price, high_price, low_price, close_price, volume, quote_volume, 
		 open_time, close_time, trade_count, is_closed, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, tableName)

	insertStmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	updateQuery := fmt.Sprintf(`
		UPDATE %s 
		SET high_price = ?, 
		    low_price = ?, 
		    close_price = ?, 
		    volume = ?, 
		    quote_volume = ?, 
		    trade_count = ?,
		    is_closed = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE symbol = ? AND open_time = ?
	`, tableName)

	updateStmt, err := tx.PrepareContext(ctx, updateQuery)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	for _, bar := range ohlcvBars {
		var existingOpenPrice, existingHighPrice, existingLowPrice, existingClosePrice float64
		var existingVolume, existingQuoteVolume float64
		var existingTradeCount int64

		row := selectStmt.QueryRowContext(ctx, bar.Symbol, bar.OpenTime)
		err := row.Scan(&existingOpenPrice, &existingHighPrice, &existingLowPrice,
			&existingClosePrice, &existingVolume, &existingQuoteVolume, &existingTradeCount)

		if err == sql.ErrNoRows {
			// insert directly
			_, err = insertStmt.ExecContext(ctx,
				bar.Symbol,
				bar.OpenPrice,
				bar.HighPrice,
				bar.LowPrice,
				bar.ClosePrice,
				bar.Volume,
				bar.QuoteVolume,
				bar.OpenTime,
				bar.CloseTime,
				bar.TradeCount,
				r.boolToInt(bar.IsClosed),
			)
			if err != nil {
				return fmt.Errorf("failed to insert new bar: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query existing bar: %w", err)
		} else {
			// exist, accumulate data
			//accumulate rule:
			// - open_price keep the same.
			// - high_price max
			// - low_price min
			// - close_price update
			// - volume & quote_volume accumulate
			// - trade_count accumulate
			newHighPrice := max(existingHighPrice, bar.HighPrice)
			newLowPrice := min(existingLowPrice, bar.LowPrice)
			newVolume := existingVolume + bar.Volume
			newQuoteVolume := existingQuoteVolume + bar.QuoteVolume
			newTradeCount := existingTradeCount + bar.TradeCount

			_, err = updateStmt.ExecContext(ctx,
				newHighPrice,
				newLowPrice,
				bar.ClosePrice,
				newVolume,
				newQuoteVolume,
				newTradeCount,
				r.boolToInt(bar.IsClosed),
				bar.Symbol,
				bar.OpenTime,
			)
			if err != nil {
				return fmt.Errorf("failed to update existing bar: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *SQLiteOHLCVRepository) getTableName(interval OHLCV_INTERVAL) (string, error) {
	config, ok := SupportedIntervals[interval]
	if !ok {
		return "", fmt.Errorf("unsupported interval: %s", interval)
	}
	return config.Table, nil
}

func (r *SQLiteOHLCVRepository) parseDuration(interval OHLCV_INTERVAL) (time.Duration, error) {
	config, ok := SupportedIntervals[interval]
	if !ok {
		return 0, fmt.Errorf("unsupported interval: %s", interval)
	}
	return config.Duration, nil
}

func (r *SQLiteOHLCVRepository) boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// CloseRealtimeOHLCV
func (r *SQLiteOHLCVRepository) CloseRealtimeOHLCV(ctx context.Context, symbol string, interval OHLCV_INTERVAL, openTime int64) error {
	// 開始事務
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 從實時表獲取數據
	selectQuery := `
		SELECT symbol, open_price, high_price, low_price, close_price, 
		       volume, quote_volume, open_time, close_time, trade_count
		FROM ohlcv_realtime 
		WHERE symbol = ? AND interval_type = ? AND open_time = ?
	`

	row := tx.QueryRowContext(ctx, selectQuery, symbol, string(interval), openTime)

	var bar OHLCVBar
	err = row.Scan(
		&bar.Symbol,
		&bar.OpenPrice,
		&bar.HighPrice,
		&bar.LowPrice,
		&bar.ClosePrice,
		&bar.Volume,
		&bar.QuoteVolume,
		&bar.OpenTime,
		&bar.CloseTime,
		&bar.TradeCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil // 沒有找到實時數據，直接返回
		}
		return err
	}

	// 2. 插入到歷史表
	tableName, err := r.getTableName(interval)
	if err != nil {
		return err
	}

	if tableName != "" {
		insertQuery := fmt.Sprintf(`
			INSERT OR REPLACE INTO %s 
			(symbol, open_price, high_price, low_price, close_price, volume, quote_volume, 
			 open_time, close_time, trade_count, is_closed, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP)
		`, tableName)

		_, err = tx.ExecContext(ctx, insertQuery,
			bar.Symbol,
			bar.OpenPrice,
			bar.HighPrice,
			bar.LowPrice,
			bar.ClosePrice,
			bar.Volume,
			bar.QuoteVolume,
			bar.OpenTime,
			bar.CloseTime,
			bar.TradeCount,
		)
		if err != nil {
			return err
		}
	}

	// 3. 從實時表刪除
	deleteQuery := `
		DELETE FROM ohlcv_realtime 
		WHERE symbol = ? AND interval_type = ? AND open_time = ?
	`

	_, err = tx.ExecContext(ctx, deleteQuery, symbol, string(interval), openTime)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetOHLCVStatistics 獲取統計數據
func (r *SQLiteOHLCVRepository) GetOHLCVStatistics(ctx context.Context, symbol, interval string, date time.Time) (*ohlcvStatistics, error) {
	dateKey := date.Format("2006-01-02")

	query := `
		SELECT record_count, min_open_time, max_close_time, avg_volume, total_volume
		FROM ohlcv_statistics 
		WHERE symbol = ? AND interval_type = ? AND date_key = ?
	`

	row := r.db.QueryRowContext(ctx, query, symbol, interval, dateKey)

	stats := &ohlcvStatistics{}
	err := row.Scan(
		&stats.RecordCount,
		&stats.MinOpenTime,
		&stats.MaxCloseTime,
		&stats.AvgVolume,
		&stats.TotalVolume,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return stats, nil
}
