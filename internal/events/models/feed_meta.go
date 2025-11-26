package models

import "time"

type FeedMeta struct {
    ID        int       `gorm:"primaryKey"`
    Version   int       `json:"version"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (FeedMeta) TableName() string {
    return "feed_meta"
}
