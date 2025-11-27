package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)


type BroadcastQueue struct {
	ID        int            `gorm:"primaryKey" json:"id"`
	EventID   uuid.UUID      `gorm:"type:uuid" json:"event_id"`
	Channel   string         `json:"channel"`
	Payload datatypes.JSONMap `gorm:"type:jsonb" json:"payload"`
	Status    string         `json:"status"`
	Attempts  int            `json:"attempts"`
	LastError *string        `json:"last_error"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (BroadcastQueue) TableName() string {
	return "broadcast_queue"
}
