package test

import (
	"context"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/service/impl/amm"
	"github.com/labstack/gommon/log"
)

type MockExgFuncProxyImpl struct {
}

func (m MockExgFuncProxyImpl) GetBalance(ctx context.Context, ammUID string, marketName string) (amm.Balance, error) {
	balance := amm.NewBalance("ETH", "USDT", 100, 0, 100000, 0)
	return *balance, nil
}

func (m MockExgFuncProxyImpl) GetIndexPrice(ctx context.Context, symbol string) (float64, error) {
	return 3000, nil
}

func (m MockExgFuncProxyImpl) GetOrderBookSnapshot(ctx context.Context, marketName string) (book.BookSnapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockExgFuncProxyImpl) GetOpenOrders(ctx context.Context, ammUID string, marketName string) ([]*dto.Order, error) {
	return fake_orders("U01", "ETH-USDT"), nil
}

func (m MockExgFuncProxyImpl) PlaceOrder(ctx context.Context, user dto.User, marketName string, placeOrderReq *dto.OrderReq) error {
	log.Infof("[PlaceOrder] req: %v", placeOrderReq)
	return nil
}

func (m MockExgFuncProxyImpl) CancelOrder(ctx context.Context, ammUID string, orderId string) (*dto.Order, error) {
	log.Infof("[CancelOrder] orderId: %s", orderId)
	return nil, nil
}

func newMockExgFuncProxyImpl() amm.IAmmExchangeFuncProxy {
	return &MockExgFuncProxyImpl{}
}

func mockAmmUser() dto.User {
	return dto.User{
		Username: "TEST_AMM_U01",
		ID:       "1",
		MakerFee: 0.0001,
	}
}
