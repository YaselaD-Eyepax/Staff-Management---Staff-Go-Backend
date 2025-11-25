package repository

import (
	"events-service/internal/events/models"
	"time"

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

func (r *EventRepository) GetEventFeed(page int, size int, since *time.Time, channel string) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	offset := (page - 1) * size

	q := r.DB.Model(&models.Event{}).
		Preload("Tags").
		Order("created_at DESC")

	if since != nil {
		q = q.Where("created_at > ?", *since)
	}

	// Future feature: feed per channel (Teams, mobile, etc.)
	if channel != "" {
		_ = channel // placeholder for future filters
	}

	// Count total items
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if err := q.Offset(offset).Limit(size).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *EventRepository) UpdateEvent(event *models.Event, body *models.AnnouncementBody, tags []models.EventTag) error {
    return r.DB.Transaction(func(tx *gorm.DB) error {

        // Update event fields
        if err := tx.Model(&models.Event{}).
            Where("id = ?", event.ID).
            Updates(event).Error; err != nil {
            return err
        }

        // Update body
        if err := tx.Model(&models.AnnouncementBody{}).
            Where("event_id = ?", event.ID).
            Updates(body).Error; err != nil {
            return err
        }

        // Remove old tags
        if err := tx.Where("event_id = ?", event.ID).
            Delete(&models.EventTag{}).Error; err != nil {
            return err
        }

        // Add new tags
        if len(tags) > 0 {
            if err := tx.Create(&tags).Error; err != nil {
                return err
            }
        }

        return nil
    })
}
