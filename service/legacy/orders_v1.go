package legacy

//import (
//	"database/sql"
//	"errors"
//	"fmt"
//	"github.com/google/uuid"
//	"github.com/johnny1110/crypto-exchange/engine-v2/book"
//	"github.com/johnny1110/crypto-exchange/engine-v2/core"
//	"github.com/johnny1110/crypto-exchange/engine-v2/model"
//	"github.com/labstack/gommon/log"
//	"time"
//)
//
//type OrderService struct {
//	DB     *sql.DB
//	Engine *core.MatchingEngine
//}
//
//type OrderEntity struct {
//	ID            string            `json:"id"`
//	UserID        string            `json:"user_id"`
//	Market        string            `json:"market"`
//	Side          model.Side        `json:"side"`
//	Price         float64           `json:"price"`
//	OriginalSize  float64           `json:"original_size"`
//	RemainingSize float64           `json:"remaining_size"`
//	Type          book.OrderType    `json:"type"`
//	Mode          model.Mode        `json:"mode"`
//	Status        model.OrderStatus `json:"status"`
//	CreatedAt     time.Time         `json:"created_at"`
//	UpdatedAt     time.Time         `json:"updated_at"`
//}
//
//type PlaceOrderRequest struct {
//	UserID      string
//	Market      string
//	Side        model.Side
//	Price       float64
//	Size        float64
//	OrderType   book.OrderType
//	Mode        model.Mode
//	QuoteAmount float64
//}
//
//func (r *PlaceOrderRequest) validate() error {
//	if r.UserID == "" {
//		return errors.New("user id is required")
//	}
//	if r.Market == "" {
//		return errors.New("market is required")
//	}
//
//	if r.Side == model.ASK {
//		if r.Size <= 0 {
//			return errors.New("ask order size must be greater than zero")
//		}
//	}
//
//	if r.Side == model.BID {
//		if r.OrderType == book.MARKET && r.QuoteAmount <= 0 {
//			return errors.New("bid order quote amount must be greater than zero")
//		}
//	}
//
//	if r.OrderType == book.LIMIT && (r.Price <= 0 || r.Size <= 0) {
//		return errors.New("limit order price and size must be greater than zero")
//	}
//	return nil
//}
//
//type PlaceOrderResult struct {
//	OrderID string
//	Status  model.OrderStatus
//	Trades  []book.Trade
//}
//
//func (s *OrderService) PlaceOrder(req PlaceOrderRequest) (res *PlaceOrderResult, err error) {
//	log.Infof("[OrderService] PlaceOrder: %v", req)
//	// create TXN
//	tx, err := s.DB.Begin()
//	if err != nil {
//		return nil, err
//	}
//	defer func() {
//		if err != nil {
//			log.Error("[OrderService] PlaceOrder rollback, err:", err.Error())
//			tx.Rollback()
//		} else {
//			tx.Commit()
//		}
//	}()
//
//	// 0. basic request params check
//	err = req.validate()
//	if err != nil {
//		log.Warn("[OrderService] validate req err:", err.Error())
//		return nil, err
//	}
//
//	// 1. Freeze funds based on market and side
//	base, quote, err := s.ParseMarket(req.Market)
//	if err != nil {
//		log.Error("[OrderService] ParseMarket err:", err.Error())
//		return nil, err
//	}
//	var freezeAsset string
//	var freezeAmt float64
//	if req.Side == model.BID {
//		freezeAsset = quote
//		switch req.OrderType {
//		case book.LIMIT:
//			// limit buy order, freeze price*size
//			freezeAmt = req.Price * req.Size
//			break
//		case book.MARKET:
//			// market order freeze quoteAmt
//			freezeAmt = req.QuoteAmount
//		}
//	} else {
//		// all ask order just freeze base asset size
//		freezeAsset = base
//		freezeAmt = req.Size
//	}
//	// freeze user balances (DB)
//	updateRes, err := tx.Exec(
//		`UPDATE balances SET available=available-?, locked=locked+? WHERE user_id=? AND asset=? AND available>=?`,
//		freezeAmt, freezeAmt, req.UserID, freezeAsset, freezeAmt,
//	)
//	if err != nil {
//		log.Error("[PlaceOrder] UpdateBalances err:", err.Error())
//		return nil, err
//	}
//	if rows, _ := updateRes.RowsAffected(); rows == 0 {
//		log.Warnf("[PlaceOrder] freezeAmt failed, userID: [%s] insufficient balance, \n", req.UserID)
//		return nil, errors.New("insufficient balance")
//	}
//
//	// 2. Persist order
//	orderID := uuid.NewString()
//	now := time.Now()
//	_, err = tx.Exec(
//		`INSERT INTO orders(id,user_id,market,side,price,original_size,remaining_size, quote_amount, type, mode, status,created_at,updated_at)
//         VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
//		orderID, req.UserID, req.Market, req.Side, req.Price, req.Size, req.Size, req.QuoteAmount, req.OrderType, req.Mode, model.ORDER_STATUS_NEW, now, now,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	// 3. Match engine
//	order := &model.Order{
//		ID:            orderID,
//		UserID:        req.UserID,
//		Side:          req.Side,
//		Price:         req.Price,
//		OriginalSize:  req.Size,
//		RemainingSize: req.Size,
//		QuoteAmount:   req.QuoteAmount,
//		Mode:          req.Mode,
//		Timestamp:     now,
//	}
//	trades, err := s.Engine.PlaceOrder(req.Market, req.OrderType, order)
//	if err != nil {
//		return nil, err
//	}
//
//	// 4. Update incoming order once, based on final RemainingSize
//	status := order.GetStatus()
//	_, err = tx.Exec(
//		`UPDATE orders SET remaining_size=?, status=?, updated_at=? WHERE id=?`,
//		order.RemainingSize, status, time.Now(), orderID,
//	)
//	if err != nil {
//		log.Error("[PlaceOrder] update user incoming order failed.", err)
//		return nil, err
//	}
//
//	// 6. Persist trades and update counterparty orders & balance
//	for _, trade := range trades {
//		// 6-1. insert trade record
//		_, err := tx.Exec(
//			`INSERT INTO trades(bid_order_id, ask_order_id, price, size, timestamp)
//             VALUES(?,?,?,?,?)`,
//			trade.BidOrderID, trade.AskOrderID, trade.Price, trade.Size, trade.Timestamp,
//		)
//		if err != nil {
//			log.Error("[PlaceOrder] insert trade record failed.", err)
//			return nil, err
//		}
//
//		// 6-2. update counterparty order
//		counterSide := order.CounterSide()
//		counterOrderID := trade.GetOrderIDBySide(counterSide)
//
//		err = s.updateCounterOrder(tx, &trade, counterOrderID)
//
//		if err != nil {
//			log.Error("[PlaceOrder] update counter order failed.", err)
//			return nil, err
//		}
//
//		// 6-3. Settlement for bid & ask user.
//		err = s.settleTrade(tx, &trade, order, base, quote)
//		if err != nil {
//			log.Error("[PlaceOrder] settle trade failed.", err)
//			return nil, err
//		}
//	}
//
//	return &PlaceOrderResult{OrderID: orderID, Status: status, Trades: trades}, nil
//}
//
//// CancelOrder cancel order based on market userId, orderId
//func (s *OrderService) CancelOrder(market string, userId string, orderId string) error {
//	orderEntity, err := s.GetOrderDetailByID(orderId)
//	if err != nil {
//		return err
//	}
//
//	if (orderEntity.UserID != userId) || (orderEntity.Market != market) {
//		log.Warn("Order access invalid")
//		return errors.New("invalid access")
//	}
//
//	order, err := s.Engine.CancelOrder(market, orderId)
//	if err != nil {
//		log.Error("[PlaceOrder] cancel order failed.", err)
//		return err
//	}
//
//	// update order entity
//	_, err = s.DB.Exec(`
//			UPDATE orders SET status = ?, updated_at = ?
//			WHERE id = ?
//		`, model.ORDER_STATUS_CANCELED, time.Now(), orderId)
//
//	// unlock user's asset balance
//	var unlockAmount float64
//	var unlockAsset string
//
//	baseAsset, quoteAsset, err := s.ParseMarket(orderEntity.Market)
//
//	switch order.Side {
//	case model.BID:
//		unlockAmount = order.Price * order.RemainingSize
//		unlockAsset = quoteAsset
//		break
//	case model.ASK:
//		unlockAmount = order.RemainingSize
//		unlockAsset = baseAsset
//	}
//	log.Infof("[CancelOrder] orderId:[%s], unlock user[%s] [%s] balance: %s \n", orderId, userId, unlockAsset, unlockAmount)
//
//	_, err = s.DB.Exec(`
//			UPDATE balances SET available = available + ?, locked = locked - ? WHERE user_id = ? AND asset = ?
//		`, unlockAmount, unlockAmount, userId, unlockAsset)
//
//	return err
//}
//
//// ParseMarket input market, return base quote assets.html.
//func (s *OrderService) ParseMarket(market string) (string, string, error) {
//	ob, err := s.Engine.GetOrderBook(market)
//	if err != nil {
//		return "", "", err
//	}
//	return ob.MarketInfo().BaseAsset, ob.MarketInfo().QuoteAsset, nil
//}
//
//func (s *OrderService) updateCounterOrder(tx *sql.Tx, trade *book.Trade, counterOrderID string) error {
//	_, err := tx.Exec(`
//			UPDATE orders SET remaining_size = remaining_size - ?,
//			                  status = CASE
//			                  	WHEN remaining_size - ? = 0 THEN ?
//								WHEN remaining_size - ? < original_size THEN ?
//								ELSE status END,
//			                  updated_at = ?
//			WHERE id = ?
//		`, trade.Size, trade.Size, model.ORDER_STATUS_FILLED, trade.Size, model.ORDER_STATUS_PARTIAL, time.Now(), counterOrderID)
//	return err
//}
//
//// settleTrade
//// bid user (+ baseAsset, - quoteAsset)
//// ask user (- baseAsset, + quoteAsset)
//func (s *OrderService) settleTrade(tx *sql.Tx, trade *book.Trade, eatenOrder *model.Order, baseAsset, quoteAsset string) error {
//	var err error
//	// 1. process bid user balance (+)baseAsset
//	_, err = tx.Exec(`
//			UPDATE balances SET available = available + ? WHERE user_id = ? AND asset = ?
//		`, trade.Size, trade.BidUserID, baseAsset)
//	if err != nil {
//		return err
//	}
//	// 2. process bid user balance (-)quoteAsset (locked)
//	var bidSideQuoteAmt = trade.Size * trade.Price
//	if eatenOrder.Side == model.BID {
//		// If incomingOrder is bid, then unfrozen quote amt = order.LimitPrice * trade.Size
//		bidSideQuoteAmt = eatenOrder.Price * trade.Size
//	}
//	_, err = tx.Exec(`
//			UPDATE balances SET locked = locked - ? WHERE user_id = ? AND asset = ?
//		`, bidSideQuoteAmt, trade.BidUserID, quoteAsset)
//	if err != nil {
//		return err
//	}
//	// 3. process ask user (- baseAsset) (locked)
//	_, err = tx.Exec(`
//			UPDATE balances SET locked = locked - ? WHERE user_id = ? AND asset = ?
//		`, trade.Size, trade.AskUserID, baseAsset)
//	if err != nil {
//		return err
//	}
//	// 4. process ask user (+ quoteAsset)
//	_, err = tx.Exec(`
//			UPDATE balances SET available = available + ? WHERE user_id = ? AND asset = ?
//		`, trade.Size*trade.Price, trade.AskUserID, quoteAsset)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// GetOrderDetailByID returns an order's detail and its trades by ID.
//func (s *OrderService) GetOrderDetailByID(orderID string) (*OrderEntity, error) {
//	// 1. Query order row
//	row := s.DB.QueryRow(
//		`SELECT id, user_id, market, side, price, original_size, remaining_size, type, mode, status, created_at, updated_at
//         FROM orders WHERE id = ?`, orderID)
//
//	var e OrderEntity
//	err := row.Scan(
//		&e.ID,
//		&e.UserID,
//		&e.Market,
//		&e.Side,
//		&e.Price,
//		&e.OriginalSize,
//		&e.RemainingSize,
//		&e.Type,
//		&e.Mode,
//		&e.Status,
//		&e.CreatedAt,
//		&e.UpdatedAt,
//	)
//	if err != nil {
//		if err == sql.ErrNoRows {
//			return nil, fmt.Errorf("order %s not found", orderID)
//		}
//		return nil, err
//	}
//	return &e, nil
//}
