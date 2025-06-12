package ohlcv

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLiteOHLCVRepository struct {
	db *sql.DB
}

func NewSQLiteOHLCVRepository(dbPath string) (OHLCVRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	return &SQLiteOHLCVRepository{db: db}, nil
}

func (r *SQLiteOHLCVRepository) SaveOHLCVBar(ctx context.Context, bar *ohlcvBar, interval string) error {
	config, exists := SupportedIntervals[interval]
	if !exists {
		return fmt.Errorf("unsupported interval: %s", interval)
	}

	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (
			symbol, open_price, high_price, low_price, close_price, volume, quote_volume,
			open_time, close_time, trade_count,
			is_closed, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, config.Table)

	_, err := r.db.ExecContext(ctx, query,
		bar.Symbol, bar.OpenPrice, bar.HighPrice, bar.LowPrice, bar.ClosePrice,
		bar.Volume, bar.QuoteVolume, bar.OpenTime, bar.CloseTime, bar.TradeCount,
		1, // is_closed = 1
	)

	return err
}

func (r *SQLiteOHLCVRepository) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
	var query string
	var args []interface{}

	// Build query based on interval
	config, exists := SupportedIntervals[req.Interval]
	if !exists {
		return nil, fmt.Errorf("unsupported interval: %s", req.Interval)
	}

	query = fmt.Sprintf(`
		SELECT open_time, open_price, high_price, low_price, close_price, volume
		FROM %s
		WHERE symbol = ? AND is_closed = 1
	`, config.Table)
	args = append(args, req.Symbol)

	// Add time filters
	if !req.StartTime.IsZero() {
		query += " AND open_time >= ?"
		args = append(args, req.StartTime.Unix())
	}
	if !req.EndTime.IsZero() {
		query += " AND open_time <= ?"
		args = append(args, req.EndTime.Unix())
	}

	query += " ORDER BY open_time DESC LIMIT ?"
	args = append(args, req.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &OHLCV{
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
		var openP, highP, lowP, closeP, volume float64

		if err := rows.Scan(&timestamp, &openP, &highP, &lowP, &closeP, &volume); err != nil {
			return nil, err
		}

		result.T = append(result.T, timestamp)
		result.O = append(result.O, openP)
		result.H = append(result.H, highP)
		result.L = append(result.L, lowP)
		result.C = append(result.C, closeP)
		result.V = append(result.V, volume)
	}

	// Reverse to get ascending order
	r.reverseOHLCVArrays(result)

	return result, nil
}

func (r *SQLiteOHLCVRepository) reverseOHLCVArrays(ohlcv *OHLCV) {
	n := len(ohlcv.T)
	for i := 0; i < n/2; i++ {
		j := n - 1 - i
		ohlcv.T[i], ohlcv.T[j] = ohlcv.T[j], ohlcv.T[i]
		ohlcv.O[i], ohlcv.O[j] = ohlcv.O[j], ohlcv.O[i]
		ohlcv.H[i], ohlcv.H[j] = ohlcv.H[j], ohlcv.H[i]
		ohlcv.L[i], ohlcv.L[j] = ohlcv.L[j], ohlcv.L[i]
		ohlcv.C[i], ohlcv.C[j] = ohlcv.C[j], ohlcv.C[i]
		ohlcv.V[i], ohlcv.V[j] = ohlcv.V[j], ohlcv.V[i]
	}
}

func (r *SQLiteOHLCVRepository) UpdateRealtimeOHLCV(ctx context.Context, bar *ohlcvBar, interval string) error {
	query := `
		INSERT OR REPLACE INTO ohlcv_realtime (
			symbol, interval_type, open_price, high_price, low_price, close_price,
			volume, quote_volume, open_time, close_time, trade_count,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query,
		bar.Symbol, interval, bar.OpenPrice, bar.HighPrice, bar.LowPrice, bar.ClosePrice,
		bar.Volume, bar.QuoteVolume, bar.OpenTime, bar.CloseTime, bar.TradeCount,
	)

	return err
}

func (r *SQLiteOHLCVRepository) GetRealtimeOHLCV(ctx context.Context, symbol, interval string, openTime int64) (*ohlcvBar, error) {
	query := `
		SELECT symbol, open_price, high_price, low_price, close_price, volume, quote_volume,
			   open_time, close_time, trade_count
		FROM ohlcv_realtime
		WHERE symbol = ? AND interval_type = ? AND open_time = ?
	`

	row := r.db.QueryRowContext(ctx, query, symbol, interval, openTime)

	bar := &ohlcvBar{}
	err := row.Scan(
		&bar.Symbol, &bar.OpenPrice, &bar.HighPrice, &bar.LowPrice, &bar.ClosePrice,
		&bar.Volume, &bar.QuoteVolume, &bar.OpenTime, &bar.CloseTime, &bar.TradeCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return bar, nil
}

func (r *SQLiteOHLCVRepository) UpdateStatistics(ctx context.Context, symbol, interval string, date time.Time, stats *ohlcvStatistics) error {
	query := `
		INSERT OR REPLACE INTO ohlcv_statistics (
			symbol, interval_type, date_key, record_count, min_open_time, max_close_time,
			avg_volume, total_volume, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	dateKey := date.Format("2006-01-02")
	_, err := r.db.ExecContext(ctx, query,
		symbol, interval, dateKey, stats.RecordCount, stats.MinOpenTime, stats.MaxCloseTime,
		stats.AvgVolume, stats.TotalVolume,
	)

	return err
}

func (r *SQLiteOHLCVRepository) Close() error {
	return r.db.Close()
}
