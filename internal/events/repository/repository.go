package repository

import (
	"encoding/json"
	"events-service/internal/events/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EventRepository struct {
	DB *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{DB: db}
}

// Helper function to convert map to datatypes.JSON
func toJSON(data map[string]interface{}) datatypes.JSON {
	if data == nil {
		return nil
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return datatypes.JSON(jsonBytes)
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

func (r *EventRepository) ModerateEvent(eventID uuid.UUID, status string, moderatorID uuid.UUID, notes string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {

		// Update event status
		if err := tx.Model(&models.Event{}).
			Where("id = ?", eventID).
			Updates(map[string]interface{}{
				"status": status,
			}).Error; err != nil {
			return err
		}

		// Insert audit entry - FIXED: convert map to datatypes.JSON
		audit := models.PublishAudit{
			EventID: eventID,
			Channel: "moderation",
			Status:  status,
			Details: toJSON(map[string]interface{}{
				"moderator_id": moderatorID.String(),
				"notes":        notes,
			}),
			CreatedAt: time.Now(),
		}

		if err := tx.Create(&audit).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *EventRepository) GetFeedVersion() (int, error) {
	var meta models.FeedMeta

	// FeedMeta row should always be id=1
	err := r.DB.First(&meta, "id = 1").Error
	if err != nil {
		return 0, err
	}

	return meta.Version, nil
}

func (r *EventRepository) IncrementFeedVersion() error {
	return r.DB.Exec(`
        UPDATE feed_meta 
        SET version = version + 1, updated_at = NOW()
        WHERE id = 1
    `).Error
}

// Enqueue a broadcast job for a given event and channel
func (r *EventRepository) EnqueueBroadcast(eventID uuid.UUID, channel string, payload map[string]any) error {
	job := models.BroadcastQueue{
		EventID: eventID,
		Channel: channel,
		Payload: datatypes.JSONMap(payload),
		Status:  "pending",
	}

	return r.DB.Create(&job).Error
}

// Fetch pending jobs up to a limit and mark them processing (returns rows)
func (r *EventRepository) FetchPendingBroadcasts(limit int) ([]models.BroadcastQueue, error) {
	// var jobs []models.BroadcastQueue

	// Use FOR UPDATE SKIP LOCKED pattern via raw SQL to avoid races if you have multiple workers.
	tx := r.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Select pending jobs and mark as processing
	rows := []models.BroadcastQueue{}
	err := tx.Raw(`
		UPDATE broadcast_queue
		SET status = 'processing', updated_at = now()
		WHERE id IN (
		  SELECT id FROM broadcast_queue
		  WHERE status = 'pending'
		  ORDER BY created_at ASC
		  LIMIT ?
		  FOR UPDATE SKIP LOCKED
		)
		RETURNING *
	`, limit).Scan(&rows).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return rows, nil
}

// Update job status after processing
func (r *EventRepository) UpdateBroadcastJobStatus(jobID int, status string, attempts int, lastError *string) error {
	return r.DB.Model(&models.BroadcastQueue{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":     status,
			"attempts":   attempts,
			"last_error": lastError,
			"updated_at": time.Now(),
		}).Error
}

func (r *EventRepository) CreatePublishAudit(eventID uuid.UUID, channel, status string, details map[string]any) error {
	// FIXED: convert map to datatypes.JSON
	audit := models.PublishAudit{
		EventID:   eventID,
		Channel:   channel,
		Status:    status,
		Details:   toJSON(details),
		CreatedAt: time.Now(),
	}
	return r.DB.Create(&audit).Error
}

func (r *EventRepository) SearchGlobalTags(query string) ([]models.GlobalTag, error) {
	var tags []models.GlobalTag
	q := r.DB.Model(&models.GlobalTag{})

	if query != "" {
		q = q.Where("tag ILIKE ?", "%"+query+"%")
	}

	if err := q.Order("tag ASC").Limit(20).Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *EventRepository) FetchActiveStaffEmails() ([]string, error) {
    var emails []string
    err := r.DB.Raw(`SELECT email FROM app_users`).Scan(&emails).Error
    return emails, err
}
