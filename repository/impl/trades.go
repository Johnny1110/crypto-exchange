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

func (t tradeRepository) BatchInsert(ctx context.Context, db repository.DBExecutor, trades []*book.Trade) error {
	if len(trades) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(trades))
	valueArgs := make([]interface{}, 0, len(trades)*5) // 5 columns：ask_order_id, bid_order_id, price, size, timestamp

	for _, trade := range trades {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			trade.AskOrderID,
			trade.BidOrderID,
			trade.Price,
			trade.Size,
			trade.Timestamp,
		)
	}

	query := fmt.Sprintf("INSERT INTO trades (ask_order_id, bid_order_id, price, size, timestamp) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert trades: %w", err)
	}

	return nil
}
