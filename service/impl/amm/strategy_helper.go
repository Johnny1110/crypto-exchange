package amm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/labstack/gommon/log"
	"io"
	"net/http"
)

type Balance struct {
	baseAsset      string  // ex: BTC, ETH
	quoteAsset     string  // ex: USDT, USDC
	baseAvailable  float64 // able to place order
	baseLocked     float64 // locked by opening order
	quoteAvailable float64 // able to place order
	quoteLocked    float64 // locked by opening order
}

func NewBalance(baseAsset, quoteAsset string, baseAva, baseLocked, quoteAva, quoteLocked float64) *Balance {
	return &Balance{
		baseAsset:      baseAsset,
		quoteAsset:     quoteAsset,
		baseAvailable:  baseAva,
		baseLocked:     baseLocked,
		quoteAvailable: quoteAva,
		quoteLocked:    quoteLocked,
	}
}

type IAmmExchangeFuncProxy interface {
	GetBalance(ctx context.Context, ammUID string, marketName string) (Balance, error)
	GetIndexPrice(ctx context.Context, symbol string) (float64, error)
	GetOrderBookSnapshot(ctx context.Context, marketName string) (book.BookSnapshot, error)
	GetOpenOrders(ctx context.Context, ammUID string, marketName string) ([]*dto.Order, error)
	PlaceOrder(ctx context.Context, user dto.User, marketName string, placeOrderReq *dto.OrderReq) error
	CancelOrder(ctx context.Context, ammUID string, orderId string) (*dto.Order, error)
}

type AmmExchangeFuncProxyImpl struct {
	orderBookService service.IOrderBookService
	balanceService   service.IBalanceService
	orderService     service.IOrderService
	userService      service.IUserService
	httpClient       *http.Client
}

func NewAmmExchangeFuncProxyImpl(orderBookService service.IOrderBookService,
	balanceService service.IBalanceService,
	orderService service.IOrderService,
	userService service.IUserService,
	httpClient *http.Client) IAmmExchangeFuncProxy {
	return &AmmExchangeFuncProxyImpl{
		orderBookService: orderBookService,
		balanceService:   balanceService,
		orderService:     orderService,
		userService:      userService,
		httpClient:       httpClient,
	}
}

func (a AmmExchangeFuncProxyImpl) GetBalance(ctx context.Context, ammUID string, marketName string) (Balance, error) {
	balances, err := a.balanceService.GetBalances(ctx, ammUID)
	if err != nil {
		return Balance{}, err
	}
	base, quote, err := a.orderBookService.GetBaseQuoteAssets(ctx, marketName)
	if err != nil {
		return Balance{}, err
	}

	var baseAva, baseLocked, quoteAva, quoteLocked float64
	for _, balance := range balances {
		if balance.Asset == base {
			baseAva = balance.Available
			baseLocked = balance.Locked
		}
		if balance.Asset == quote {
			quoteAva = balance.Available
			quoteLocked = balance.Locked
		}
	}

	return Balance{
		base,
		quote,
		baseAva,
		baseLocked,
		quoteAva,
		quoteLocked}, nil
}

// MarketSummary represents the structure of market data from BTSE API
type MarketSummary struct {
	Symbol              string      `json:"symbol"`
	Last                float64     `json:"last"`
	LowestAsk           float64     `json:"lowestAsk"`
	HighestBid          float64     `json:"highestBid"`
	PercentageChange    float64     `json:"percentageChange"`
	Volume              float64     `json:"volume"`
	High24Hr            float64     `json:"high24Hr"`
	Low24Hr             float64     `json:"low24Hr"`
	Base                string      `json:"base"`
	Quote               string      `json:"quote"`
	Active              bool        `json:"active"`
	Size                float64     `json:"size"`
	MinValidPrice       float64     `json:"minValidPrice"`
	MinPriceIncrement   float64     `json:"minPriceIncrement"`
	MinOrderSize        float64     `json:"minOrderSize"`
	MaxOrderSize        float64     `json:"maxOrderSize"`
	MinSizeIncrement    float64     `json:"minSizeIncrement"`
	OpenInterest        float64     `json:"openInterest"`
	OpenInterestUSD     float64     `json:"openInterestUSD"`
	ContractStart       int64       `json:"contractStart"`
	ContractEnd         int64       `json:"contractEnd"`
	TimeBasedContract   bool        `json:"timeBasedContract"`
	OpenTime            int64       `json:"openTime"`
	CloseTime           int64       `json:"closeTime"`
	StartMatching       int64       `json:"startMatching"`
	InactiveTime        int64       `json:"inactiveTime"`
	FundingRate         float64     `json:"fundingRate"`
	ContractSize        float64     `json:"contractSize"`
	MaxPosition         float64     `json:"maxPosition"`
	MinRiskLimit        float64     `json:"minRiskLimit"`
	MaxRiskLimit        float64     `json:"maxRiskLimit"`
	AvailableSettlement interface{} `json:"availableSettlement"`
	Futures             bool        `json:"futures"`
	IsMarketOpenToOtc   bool        `json:"isMarketOpenToOtc"`
	IsMarketOpenToSpot  bool        `json:"isMarketOpenToSpot"`
}

// GetIndexPrice retrieves the latest price for a given symbol from BTSE API
func (a *AmmExchangeFuncProxyImpl) GetIndexPrice(ctx context.Context, symbol string) (float64, error) {
	const apiURL = "https://api.btse.com/spot/api/v3.2/market_summary"

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-HTTP-Client/1.1")

	// Make the HTTP request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var marketSummaries []MarketSummary
	if err := json.Unmarshal(body, &marketSummaries); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Find the symbol in the response
	for _, market := range marketSummaries {
		if market.Symbol == symbol {
			if !market.Active {
				return 0, fmt.Errorf("market for symbol %s is not active", symbol)
			}
			return market.Last, nil
		}
	}

	log.Warnf("[GetIndexPrcie] not found by BTSE API, using Default")

	return a.orderBookService.GetLatestPrice(ctx, symbol)
}

func (a AmmExchangeFuncProxyImpl) GetOrderBookSnapshot(ctx context.Context, marketName string) (book.BookSnapshot, error) {
	snapshot, err := a.orderBookService.GetSnapshot(ctx, marketName)
	return *snapshot, err

}

func (a AmmExchangeFuncProxyImpl) GetOpenOrders(ctx context.Context, ammUID string, marketName string) ([]*dto.Order, error) {
	return a.orderService.QueryOrderByMarket(ctx, ammUID, marketName, true)
}

func (a AmmExchangeFuncProxyImpl) PlaceOrder(ctx context.Context, user dto.User, marketName string, placeOrderReq *dto.OrderReq) error {
	_, err := a.orderService.PlaceOrder(ctx, marketName, &user, placeOrderReq)
	return err
}

func (a AmmExchangeFuncProxyImpl) CancelOrder(ctx context.Context, ammUID string, orderId string) (*dto.Order, error) {
	return a.orderService.CancelOrder(ctx, ammUID, orderId)
}
