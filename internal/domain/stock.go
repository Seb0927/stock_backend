package domain

import (
	"context"
	"time"
)

// Stock represents a stock rating/target information from brokerages
type Stock struct {
	ID           int64     `json:"id,string" db:"id"`
	Ticker       string    `json:"ticker" db:"ticker" binding:"required"`
	TargetFrom   string    `json:"target_from" db:"target_from"`
	TargetTo     string    `json:"target_to" db:"target_to"`
	Company      string    `json:"company" db:"company" binding:"required"`
	ActionID     int64     `json:"action_id,string" db:"action_id"`
	BrokerageID  int64     `json:"brokerage_id,string" db:"brokerage_id"`
	RatingFromID int64     `json:"rating_from_id,string" db:"rating_from_id"`
	RatingToID   int64     `json:"rating_to_id,string" db:"rating_to_id"`
	Time         time.Time `json:"time" db:"time"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`

	// Temporary fields for API client - not persisted to database
	Action     string `json:"-" db:"-"`
	Brokerage  string `json:"-" db:"-"`
	RatingFrom string `json:"-" db:"-"`
	RatingTo   string `json:"-" db:"-"`
}

// StockWithDetails represents a stock with joined details from related tables
type StockWithDetails struct {
	ID             int64     `json:"id,string" db:"id"`
	Ticker         string    `json:"ticker" db:"ticker"`
	TargetFrom     string    `json:"target_from" db:"target_from"`
	TargetTo       string    `json:"target_to" db:"target_to"`
	Company        string    `json:"company" db:"company"`
	ActionID       *int64    `json:"action_id,string,omitempty" db:"action_id"`
	ActionName     string    `json:"action,omitempty" db:"action_name"`
	BrokerageID    *int64    `json:"brokerage_id,string,omitempty" db:"brokerage_id"`
	BrokerageName  string    `json:"brokerage,omitempty" db:"brokerage_name"`
	RatingFromID   *int64    `json:"rating_from_id,string,omitempty" db:"rating_from_id"`
	RatingFromTerm string    `json:"rating_from,omitempty" db:"rating_from_term"`
	RatingToID     *int64    `json:"rating_to_id,string,omitempty" db:"rating_to_id"`
	RatingToTerm   string    `json:"rating_to,omitempty" db:"rating_to_term"`
	Time           time.Time `json:"time" db:"time"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// StockRecommendation represents a stock with its recommendation score
type StockRecommendation struct {
	Stock          *StockWithDetails `json:"stock"`
	Score          float64           `json:"score"`
	Reason         string            `json:"reason"`
	TargetIncrease float64           `json:"target_increase_percent,omitempty"`
}

// StockFilter represents filters for querying stocks
type StockFilter struct {
	Ticker     string
	Company    string
	Brokerage  string
	Action     string
	RatingFrom string
	RatingTo   string
	SortBy     string
	SortOrder  string
	Limit      int
	Offset     int
}

// StockRepository defines the interface for stock data persistence
type StockRepository interface {
	CreateBatch(stocks []*Stock) error
	FindByID(id int64) (*StockWithDetails, error)
	FindAll(filter StockFilter) ([]*StockWithDetails, error)
	FindByTicker(ticker string) ([]*StockWithDetails, error)
	Count(filter StockFilter) (int64, error)
}

// StockAPIClient defines the interface for fetching stocks from external API
type StockAPIClient interface {
	FetchAllStocks(ctx context.Context) ([]*Stock, error)
}
