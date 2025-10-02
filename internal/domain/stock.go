package domain

import (
	"context"
	"time"
)

// Stock represents a stock rating/target information from brokerages
type Stock struct {
	ID         int64     `json:"id,string" db:"id"`
	Ticker     string    `json:"ticker" db:"ticker" binding:"required"`
	TargetFrom string    `json:"target_from" db:"target_from"`
	TargetTo   string    `json:"target_to" db:"target_to"`
	Company    string    `json:"company" db:"company" binding:"required"`
	Action     string    `json:"action" db:"action"`
	Brokerage  string    `json:"brokerage" db:"brokerage"`
	RatingFrom string    `json:"rating_from" db:"rating_from"`
	RatingTo   string    `json:"rating_to" db:"rating_to"`
	Time       time.Time `json:"time" db:"time"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// StockFilter represents filters for querying stocks
type StockFilter struct {
	Ticker    string
	Company   string
	Brokerage string
	Action    string
	Limit     int
	Offset    int
}

// StockRepository defines the interface for stock data persistence
type StockRepository interface {
	CreateBatch(stocks []*Stock) error
	FindByID(id int64) (*Stock, error)
	FindAll(filter StockFilter) ([]*Stock, error)
	Count(filter StockFilter) (int64, error)
}

// StockAPIClient defines the interface for fetching stocks from external API
type StockAPIClient interface {
	FetchAllStocks(ctx context.Context) ([]*Stock, error)
}
