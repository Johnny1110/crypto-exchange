package external

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"io"
	"net/http"
	"time"
)

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

func GetIndexPrice(ctx context.Context, symbol string) (float64, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

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
	resp, err := httpClient.Do(req)
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
	return 0.01, nil
}
