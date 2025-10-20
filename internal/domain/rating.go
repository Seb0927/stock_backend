package domain

import (
	"context"
	"time"
)

// Rating represents a stock rating term (independent of brokerage)
type Rating struct {
	ID        int64     `json:"id,string" db:"id"`
	Term      string    `json:"term" db:"term" binding:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RatingRepository defines the interface for rating data persistence
type RatingRepository interface {
	Create(rating *Rating) error
	FindByID(id int64) (*Rating, error)
	FindByTerm(term string) (*Rating, error)
	FindAll(ctx context.Context) ([]*Rating, error)
}
