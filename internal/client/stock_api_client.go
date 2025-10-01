package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/company/stock-api/internal/config"
	"github.com/company/stock-api/internal/domain"
)

// StockAPIResponse represents the response from the external stock API
type StockAPIResponse struct {
	Items    []StockAPIItem `json:"items"`
	NextPage string         `json:"next_page"`
}

// StockAPIItem represents a single stock item from the API
type StockAPIItem struct {
	Ticker     string    `json:"ticker"`
	TargetFrom string    `json:"target_from"`
	TargetTo   string    `json:"target_to"`
	Company    string    `json:"company"`
	Action     string    `json:"action"`
	Brokerage  string    `json:"brokerage"`
	RatingFrom string    `json:"rating_from"`
	RatingTo   string    `json:"rating_to"`
	Time       time.Time `json:"time"`
}

// StockAPIClient handles communication with the external stock API
type StockAPIClient struct {
	httpClient *http.Client
	config     *config.StockAPIConfig
}

// NewStockAPIClient creates a new StockAPIClient
func NewStockAPIClient(cfg *config.StockAPIConfig) *StockAPIClient {
	return &StockAPIClient{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
	}
}

// FetchStocks retrieves stocks from the external API
func (c *StockAPIClient) FetchStocks(ctx context.Context, nextPage string) (*StockAPIResponse, error) {
	url := c.config.URL
	if nextPage != "" {
		url = fmt.Sprintf("%s?next_page=%s", url, nextPage)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrExternalAPI, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", domain.ErrExternalAPI, resp.StatusCode, string(body))
	}

	var apiResponse StockAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResponse, nil
}

// FetchAllStocks retrieves all stocks by paginating through the API
func (c *StockAPIClient) FetchAllStocks(ctx context.Context) ([]*domain.Stock, error) {
	var allStocks []*domain.Stock
	nextPage := ""

	for {
		select {
		case <-ctx.Done():
			return nil, domain.ErrTimeout
		default:
		}

		response, err := c.FetchStocks(ctx, nextPage)
		if err != nil {
			return nil, err
		}

		// Convert API items to domain stocks
		for _, item := range response.Items {
			stock := &domain.Stock{
				Ticker:     item.Ticker,
				TargetFrom: item.TargetFrom,
				TargetTo:   item.TargetTo,
				Company:    item.Company,
				Action:     item.Action,
				Brokerage:  item.Brokerage,
				RatingFrom: item.RatingFrom,
				RatingTo:   item.RatingTo,
				Time:       item.Time,
			}
			allStocks = append(allStocks, stock)
		}

		// Check if there's a next page
		if response.NextPage == "" {
			break
		}
		nextPage = response.NextPage
	}

	return allStocks, nil
}
