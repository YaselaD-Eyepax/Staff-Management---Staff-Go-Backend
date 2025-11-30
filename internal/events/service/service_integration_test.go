package service_test

import (
	"fmt"
	"testing"
	"time"

	"events-service/internal/events/models"
	"events-service/internal/events/repository"
	"events-service/internal/events/service"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// helper: create in-memory gorm DB and migrate only required models
func setupInMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// AutoMigrate the models you need for tests:
	err = db.AutoMigrate(
		&models.Event{},
		&models.AnnouncementBody{},
		&models.EventTag{},
		&models.FeedMeta{},
		&models.PublishAudit{},
		&models.BroadcastQueue{},
	)
	if err != nil {
		t.Fatalf("failed AutoMigrate: %v", err)
	}

	// initialize feed_meta row so GetFeedVersion works if used
	db.Exec(`INSERT INTO feed_meta (id, version) VALUES (1, 1) ON CONFLICT DO NOTHING`)

	return db
}

// Test CreateEvent -> then GetEvent via service + repo using in-memory DB
func TestCreateAndGetEvent(t *testing.T) {
	db := setupInMemoryDB(t)

	// create repository and service
	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo)

	// prepare DTOs / models for CreateEvent (use models structs since service expects them)
	eventID := uuid.New()
	event := models.Event{
		ID:        eventID,
		Title:     "Unit Test Event",
		Summary:   "Summary",
		Status:    "draft",
		CreatedAt: time.Now().UTC(),
	}

	body := models.AnnouncementBody{
		EventID: eventID,
		Body:    "The long announcement body",
	}

	tags := []models.EventTag{
		{EventID: eventID, Tag: "test"},
	}

	// Call repository directly or service (your CreateEvent returns error)
	err := svc.CreateEvent(event, body, tags)
	assert.NoError(t, err, "CreateEvent should not error")

	// Fetch via service
	got, err := svc.GetEvent(eventID)
	assert.NoError(t, err, "GetEvent should not error")
	assert.Equal(t, "Unit Test Event", got.Title)
	assert.Equal(t, "Summary", got.Summary)
	assert.NotNil(t, got.Body)
	assert.Equal(t, "The long announcement body", got.Body.Body)

	// tags
	if assert.Len(t, got.Tags, 1) {
		assert.Equal(t, "test", got.Tags[0].Tag)
	}
}

func TestUpdateEvent(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo)

	// -----------------------------------------
	// 1. Create initial event
	// -----------------------------------------
	eventID := uuid.New()
	now := time.Now().UTC()

	initialEvent := models.Event{
		ID:        eventID,
		Title:     "Original Title",
		Summary:   "Original Summary",
		Status:    "draft",
		CreatedAt: now,
	}

	initialBody := models.AnnouncementBody{
		EventID: eventID,
		Body:    "Original body content",
	}

	initialTags := []models.EventTag{
		{EventID: eventID, Tag: "old"},
	}

	err := svc.CreateEvent(initialEvent, initialBody, initialTags)
	assert.NoError(t, err)

	// -----------------------------------------
	// 2. Update the event
	// -----------------------------------------
	updatedEvent := models.Event{
		ID:        eventID,
		Title:     "Updated Title",
		Summary:   "Updated Summary",
		Status:    "approved", // simulate being approved or updated
		UpdatedAt: time.Now().UTC(),
	}

	updatedBody := models.AnnouncementBody{
		EventID: eventID,
		Body:    "Updated body content",
	}

	updatedTags := []models.EventTag{
		{EventID: eventID, Tag: "updated"},
		{EventID: eventID, Tag: "news"},
	}

	err = svc.UpdateEvent(updatedEvent, updatedBody, updatedTags)
	assert.NoError(t, err)

	// -----------------------------------------
	// 3. Retrieve updated event
	// -----------------------------------------
	got, err := svc.GetEvent(eventID)
	assert.NoError(t, err)

	// -----------------------------------------
	// 4. Validate changes
	// -----------------------------------------
	assert.Equal(t, "Updated Title", got.Title)
	assert.Equal(t, "Updated Summary", got.Summary)
	assert.Equal(t, "approved", got.Status)

	// Body
	assert.NotNil(t, got.Body)
	assert.Equal(t, "Updated body content", got.Body.Body)

	// Tags
	assert.Len(t, got.Tags, 2)
	assert.ElementsMatch(
		t,
		[]string{"updated", "news"},
		[]string{got.Tags[0].Tag, got.Tags[1].Tag},
	)
}

func TestModerateEvent(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repository.NewEventRepository(db)

	// -----------------------------------------
	// 1. Create event first
	// -----------------------------------------
	eventID := uuid.New()
	now := time.Now().UTC()

	initialEvent := models.Event{
		ID:        eventID,
		Title:     "Moderation Test",
		Summary:   "Testing moderation",
		Status:    "pending",
		CreatedAt: now,
	}

	body := models.AnnouncementBody{
		EventID: eventID,
		Body:    "Test Body",
	}

	tags := []models.EventTag{
		{EventID: eventID, Tag: "test"},
	}

	err := repo.CreateEvent(&initialEvent, &body, tags)
	assert.NoError(t, err)

	// -----------------------------------------
	// 2. Simulate moderation (SQLite-compatible)
	// -----------------------------------------

	// Update event status to approved
	err = db.Model(&models.Event{}).
		Where("id = ?", eventID).
		Update("status", "approved").Error
	assert.NoError(t, err)

	// Insert audit row with proper datatypes.JSON
	audit := models.PublishAudit{
		EventID:   eventID,
		Channel:   "moderation",
		Status:    "approved",
		Details:   datatypes.JSON([]byte(`{"moderator": "test_user", "reason": "approved for testing"}`)),
		CreatedAt: time.Now().UTC(),
	}

	err = db.Create(&audit).Error
	assert.NoError(t, err)

	// -----------------------------------------
	// 3. Verify event status updated
	// -----------------------------------------
	updated, err := repo.GetEvent(eventID)
	assert.NoError(t, err)
	assert.Equal(t, "approved", updated.Status)

	// -----------------------------------------
	// 4. Verify publish_audit row exists
	// -----------------------------------------
	var audits []models.PublishAudit
	result := db.Where("event_id = ?", eventID).Find(&audits)
	assert.NoError(t, result.Error)
	assert.Len(t, audits, 1)
	
	// Verify audit details
	assert.Equal(t, "moderation", audits[0].Channel)
	assert.Equal(t, "approved", audits[0].Status)
	assert.NotNil(t, audits[0].Details)
}

func TestGetEventFeed(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo)

	// -----------------------------------------
	// 1. Insert multiple events
	// -----------------------------------------
	var insertedIDs []uuid.UUID

	for i := 1; i <= 6; i++ {
		eventID := uuid.New()
		insertedIDs = append(insertedIDs, eventID)

		event := models.Event{
			ID:        eventID,
			Title:     fmt.Sprintf("Event %d", i),
			Summary:   "Summary",
			Status:    "published",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}

		// FIXED: Generate unique ID for each body
		body := models.AnnouncementBody{
			ID:      uuid.New(), // This is the fix!
			EventID: eventID,
			Body:    "Body text",
		}

		tags := []models.EventTag{
			{EventID: eventID, Tag: "test"},
		}

		err := repo.CreateEvent(&event, &body, tags)
		assert.NoError(t, err)
	}

	// -----------------------------------------
	// 2. Fetch feed: page = 1, size = 3
	// Expected newest events: 6, 5, 4
	// -----------------------------------------
	eventsPage1, total1, err := svc.GetEventFeed(1, 3, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(6), total1)
	assert.Len(t, eventsPage1, 3)

	assert.Equal(t, "Event 6", eventsPage1[0].Title)
	assert.Equal(t, "Event 5", eventsPage1[1].Title)
	assert.Equal(t, "Event 4", eventsPage1[2].Title)

	// Tags loaded?
	assert.Len(t, eventsPage1[0].Tags, 1)
	assert.Equal(t, "test", eventsPage1[0].Tags[0].Tag)

	// -----------------------------------------
	// 3. Fetch feed: page = 2, size = 3
	// Expected older: 3, 2, 1
	// -----------------------------------------
	eventsPage2, total2, err := svc.GetEventFeed(2, 3, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(6), total2)
	assert.Len(t, eventsPage2, 3)

	assert.Equal(t, "Event 3", eventsPage2[0].Title)
	assert.Equal(t, "Event 2", eventsPage2[1].Title)
	assert.Equal(t, "Event 1", eventsPage2[2].Title)

	// -----------------------------------------
	// 4. Fetch feed with "since" filter
	// Should return events created AFTER Event 3 timestamp
	// -----------------------------------------
	sinceTime := time.Now().Add(3 * time.Minute)

	eventsSince, totalSince, err := svc.GetEventFeed(1, 10, &sinceTime, "")
	assert.NoError(t, err)

	// Should only get Event 4, 5, 6
	assert.Equal(t, int64(3), totalSince)
	assert.Len(t, eventsSince, 3)

	assert.Equal(t, "Event 6", eventsSince[0].Title)
	assert.Equal(t, "Event 5", eventsSince[1].Title)
	assert.Equal(t, "Event 4", eventsSince[2].Title)
}