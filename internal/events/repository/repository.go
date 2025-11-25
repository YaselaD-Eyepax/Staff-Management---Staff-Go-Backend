package repository

import (
    "events-service/internal/events/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type EventRepository struct {
    DB *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
    return &EventRepository{DB: db}
}

func (r *EventRepository) CreateEvent(event *models.Event, body *models.AnnouncementBody, tags []models.EventTag) error {
    return r.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(event).Error; err != nil {
            return err
        }

        if err := tx.Create(body).Error; err != nil {
            return err
        }

        if len(tags) > 0 {
            if err := tx.Create(&tags).Error; err != nil {
                return err
            }
        }

        return nil
    })
}

func (r *EventRepository) GetEvent(id uuid.UUID) (*models.Event, error) {
    var event models.Event
    err := r.DB.
        Preload("Body").
        Preload("Tags").
        First(&event, "id = ?", id).Error

    if err != nil {
        return nil, err
    }

    return &event, nil
}
