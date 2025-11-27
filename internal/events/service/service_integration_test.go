package service_test

import (
	"testing"
	"time"

	"events-service/internal/events/models"
	"events-service/internal/events/repository"
	"events-service/internal/events/service"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
