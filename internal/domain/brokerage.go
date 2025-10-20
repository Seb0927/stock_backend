package domain

import (
	"context"
	"time"
)

// Brokerage represents a brokerage firm
type Brokerage struct {
	ID        int64     `json:"id,string" db:"id"`
	Name      string    `json:"name" db:"name" binding:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// BrokerageRepository defines the interface for brokerage data persistence
type BrokerageRepository interface {
	Create(brokerage *Brokerage) error
	FindByID(id int64) (*Brokerage, error)
	FindByName(name string) (*Brokerage, error)
	FindAll(ctx context.Context) ([]*Brokerage, error)
}
