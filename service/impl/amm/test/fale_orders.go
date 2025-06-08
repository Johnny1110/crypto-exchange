package test

import (
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"time"
)

func fake_orders(userId string, market string) []*dto.Order {
	orders := make([]*dto.Order, 0, 5)
	order1 := &dto.Order{
		ID:            "1",
		UserID:        userId,
		Market:        market,
		Side:          model.BID,
		Price:         2999,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "ETH",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order2 := &dto.Order{
		ID:            "2",
		UserID:        userId,
		Market:        market,
		Side:          model.BID,
		Price:         2998,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "ETH",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order3 := &dto.Order{
		ID:            "3",
		UserID:        userId,
		Market:        market,
		Side:          model.BID,
		Price:         2997,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "ETH",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order4 := &dto.Order{
		ID:            "4",
		UserID:        userId,
		Market:        market,
		Side:          model.BID,
		Price:         2996,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order5 := &dto.Order{
		ID:            "5",
		UserID:        userId,
		Market:        market,
		Side:          model.BID,
		Price:         2995,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Sell ____________________________________________________________

	order6 := &dto.Order{
		ID:            "6",
		UserID:        userId,
		Market:        market,
		Side:          model.ASK,
		Price:         3001,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order7 := &dto.Order{
		ID:            "7",
		UserID:        userId,
		Market:        market,
		Side:          model.ASK,
		Price:         3002,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order8 := &dto.Order{
		ID:            "8",
		UserID:        userId,
		Market:        market,
		Side:          model.ASK,
		Price:         3003,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order9 := &dto.Order{
		ID:            "9",
		UserID:        userId,
		Market:        market,
		Side:          model.ASK,
		Price:         3004,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order10 := &dto.Order{
		ID:            "10",
		UserID:        userId,
		Market:        market,
		Side:          model.ASK,
		Price:         3005,
		OriginalSize:  0.1,
		RemainingSize: 0.1,
		QuoteAmount:   0,
		AvgDealtPrice: 0.0,
		Type:          model.LIMIT,
		Mode:          model.MAKER,
		Status:        model.ORDER_STATUS_NEW,
		FeeRate:       0.0001,
		Fees:          0.0,
		FeeAsset:      "USDT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orders = append(orders, order1)
	orders = append(orders, order2)
	orders = append(orders, order3)
	orders = append(orders, order4)
	orders = append(orders, order5)
	orders = append(orders, order6)
	orders = append(orders, order7)
	orders = append(orders, order8)
	orders = append(orders, order9)
	orders = append(orders, order10)

	return orders
}
