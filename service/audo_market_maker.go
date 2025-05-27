package service

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/labstack/gommon/log"
)

type AutoMakerService struct {
	MakerName string
	MakerUID  string
}

func (s AutoMakerService) MakeMarket(os *OrderService) {
	orderReqList := make([]PlaceOrderRequest, 10)
	market := "ETH-USDT"
	// make ETH buy order * 5
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.BID, 2500, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.BID, 2550, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.BID, 2600, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.BID, 2650, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.BID, 2700, 10, book.LIMIT, model.MAKER,
	})

	// make ETH sell order * 5
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.ASK, 2500, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.ASK, 2550, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.ASK, 2600, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.ASK, 2650, 10, book.LIMIT, model.MAKER,
	})
	orderReqList = append(orderReqList, PlaceOrderRequest{
		s.MakerUID, market, model.ASK, 2700, 10, book.LIMIT, model.MAKER,
	})

	for _, req := range orderReqList {
		log.Infof("[AutoMaker] PlaceOrder, req:[%s]", req)
		_, err := os.PlaceOrder(req)
		if err != nil {
			_ = fmt.Errorf("falied to auto make market %s", err.Error())
		}
	}
}
