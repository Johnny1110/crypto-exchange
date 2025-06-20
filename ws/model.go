package ws

import (
	"encoding/json"
	"fmt"
)

type WSChannel string

const OHLCV = WSChannel("ohlcv")
const ORDERBOOK = WSChannel("orderbook")

type WSAction string

const (
	SUBSCRIBE   WSAction = "subscribe"
	UNSUBSCRIBE WSAction = "unsubscribe"
)

// WSReq WebSocket request
type WSReq struct {
	Action  WSAction    `json:"action"` // subscribe/unsubscribe
	Channel WSChannel   `json:"channel"`
	Params  interface{} `json:"params"`
}

// OHLCV params
type OHLCVReqParams struct {
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
}

// OrderBook params
type OrderBookReqParams struct {
	Market string `json:"market"`
}

// WSResp WebSocket response
type WSResp struct {
	Channel   WSChannel   `json:"channel"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type SubscriptionKey struct {
	Channel WSChannel
	Params  interface{}
}

func BuildSubscriptionKey(req WSReq) (SubscriptionKey, error) {
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return SubscriptionKey{}, err
	}

	switch req.Channel {
	case OHLCV:
		var params OHLCVReqParams
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return SubscriptionKey{}, err
		}
		return SubscriptionKey{
			Channel: req.Channel,
			Params:  params,
		}, nil

	case ORDERBOOK:
		var params OrderBookReqParams
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return SubscriptionKey{}, err
		}
		return SubscriptionKey{
			Channel: req.Channel,
			Params:  params,
		}, nil

	default:
		return SubscriptionKey{}, fmt.Errorf("Unsupport Channel: %s", req.Channel)
	}
}

type WSFeedPackage struct {
	Key  SubscriptionKey
	Data interface{}
}
