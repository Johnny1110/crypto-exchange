package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/entity"
	"log"
	"net/http"
	"time"
)

// PlaceOrderHandler order entry
func PlaceOrderHandler(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	engine := c.MustGet("engine").(*core.MatchingEngine)
	userID := c.MustGet("userID").(string)
	market := c.Param("market") // router is /:market/order

	var err error
	ob, err := engine.GetOrderBook(market)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	var req orderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// 1. freeze user asset balance
	err = tryFreezeAmount(db, ob, req, userID)
	if err != nil {
		log.Println("[PlaceOrderHandler][tryFreezeAmount] error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "insufficient balance"})
	}

	// 2. create order into DB
	orderId, err = persistOrder(db, market, req, userID)

	// 3. put order into match-engine
	trades, err := matchingOrder(ob, req, userID)
}

func matchingOrder(ob *book.OrderBook, req orderReq, id string) ([]book.Trade, error) {
	// TODO
}

func persistOrder(db *sql.DB, market string, req orderReq, userID string) (string, error) {
	orderID := uuid.NewString()
	now := time.Now()

	_, err := db.Exec(`
        INSERT INTO orders
          (id, user_id, market, side, price, original_size, remaining_size, type, status, created_at, updated_at)
        VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		orderID, userID, market, req.Side, req.Price, req.Size, req.Size, req.Mode, entity.ORDER_NEW, now, now,
	)
	if err != nil {
		return "", fmt.Errorf("insert order failed: %w", err)
	}
	return orderID, nil
}

func tryFreezeAmount(db *sql.DB, orderBook *book.OrderBook, req orderReq, userID string) error {
	baseAsset, quoteAsset := orderBook.GetAssets()
	var freezeAsset string
	var freezeAmt float64

	if req.Side == model.BID {
		freezeAsset = quoteAsset
		freezeAmt = req.Price * req.Size
	} else {
		freezeAsset = baseAsset
		freezeAmt = req.Size
	}

	res, err := db.Exec(`
        UPDATE balances
        SET available = available - ?, locked = locked + ?
        WHERE user_id = ? AND asset = ? AND available >= ?`,
		freezeAmt, freezeAmt, userID, freezeAsset, freezeAmt,
	)

	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("insufficient balance")
	}

	return nil
}
