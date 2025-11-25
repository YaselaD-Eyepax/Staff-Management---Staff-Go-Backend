package models

import (
    "time"

    "github.com/google/uuid"
)

type Event struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
    Title       string
    Summary     string
    CreatedBy   uuid.UUID `gorm:"type:uuid"`
    Status      string
    ScheduledAt *time.Time
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time

    Body AnnouncementBody `gorm:"foreignKey:EventID"`
    Tags []EventTag       `gorm:"foreignKey:EventID"`
}

type AnnouncementBody struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
    EventID     uuid.UUID `gorm:"type:uuid"`
    Body        string
    Attachments []byte `gorm:"type:jsonb"`
}

type EventTag struct {
    ID      uint      `gorm:"primaryKey"`
    EventID uuid.UUID `gorm:"type:uuid"`
    Tag     string
}
