package domain

import (
	"context"
	"time"
)

// Action represents an analyst action (e.g., upgrade, downgrade)
type Action struct {
	ID        int64     `json:"id,string" db:"id"`
	Name      string    `json:"name" db:"name" binding:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ActionRepository defines the interface for action data persistence
type ActionRepository interface {
	Create(action *Action) error
	FindByID(id int64) (*Action, error)
	FindByName(name string) (*Action, error)
	FindAll(ctx context.Context) ([]*Action, error)
}
