package repositoryImpl

import (
	"context"
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/repository"
	"strings"
)

type tradeRepository struct {
}

func NewTradeRepository() repository.ITradeRepository {
	return &tradeRepository{}
}

func (t tradeRepository) BatchInsert(ctx context.Context, db repository.DBExecutor, trades []book.Trade) error {
	if len(trades) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(trades))
	valueArgs := make([]interface{}, 0, len(trades)*8) // 8 columns

	for _, trade := range trades {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			trade.Market,
			trade.AskOrderID,
			trade.BidOrderID,
			trade.AskFeeRate,
			trade.BidFeeRate,
			trade.Price,
			trade.Size,
			trade.Timestamp,
		)
	}

	query := fmt.Sprintf("INSERT INTO trades (market, ask_order_id, bid_order_id, ask_fee_rate, bid_fee_rate, price, size, timestamp) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert trades: %w", err)
	}

	return nil
}

func (t tradeRepository) GetMarketLatestPrice(ctx context.Context, db repository.DBExecutor, marketName string) (float64, error) {
	query := `SELECT price
		FROM trades WHERE market = ? 
		ORDER BY timestamp DESC LIMIT 1`

	rows, err := db.QueryContext(ctx, query, marketName)
	if err != nil {
		return 0.0, fmt.Errorf("failed to query latest price: %w", err)
	}
	defer rows.Close()

	var price float64
	if rows.Next() {
		if err := rows.Scan(&price); err != nil {
			return 0.0, fmt.Errorf("failed to scan price: %w", err)
		}
		return price, nil
	}

	return 0.0, fmt.Errorf("no price found for market: %s", marketName)
}
