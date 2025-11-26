package service

import (
	"events-service/internal/events/models"
	"events-service/internal/events/repository"
	"time"

	"github.com/google/uuid"
)

type EventService struct {
    Repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
    return &EventService{Repo: repo}
}

func (s *EventService) CreateEvent(dto models.Event, body models.AnnouncementBody, tags []models.EventTag) error {
    return s.Repo.CreateEvent(&dto, &body, tags)
}

func (s *EventService) GetEvent(id uuid.UUID) (*models.Event, error) {
    return s.Repo.GetEvent(id)
}

func (s *EventService) GetEventFeed(page int, size int, since *time.Time, channel string) ([]models.Event, int64, error) {
	return s.Repo.GetEventFeed(page, size, since, channel)
}

func (s *EventService) UpdateEvent(event models.Event, body models.AnnouncementBody, tags []models.EventTag) error {
    return s.Repo.UpdateEvent(&event, &body, tags)
}

func (s *EventService) ModerateEvent(eventID uuid.UUID, status string, moderator uuid.UUID, notes string) error {
    return s.Repo.ModerateEvent(eventID, status, moderator, notes)
}

func (s *EventService) GetFeedVersion() (int, error) {
    return s.Repo.GetFeedVersion()
}

func (s *EventService) IncrementFeedVersion() error {
    return s.Repo.IncrementFeedVersion()
}

func (s *EventService) EnqueueBroadcast(eventID uuid.UUID, channel string, payload map[string]any) error {
    return s.Repo.EnqueueBroadcast(eventID, channel, payload)
}

func (s *EventService) FetchPendingBroadcasts(limit int) ([]models.BroadcastQueue, error) {
    return s.Repo.FetchPendingBroadcasts(limit)
}

func (s *EventService) UpdateBroadcastJobStatus(jobID int, status string, attempts int, lastError *string) error {
    return s.Repo.UpdateBroadcastJobStatus(jobID, status, attempts, lastError)
}

func (s *EventService) CreatePublishAudit(eventID uuid.UUID, channel, status string, details map[string]any) error {
    return s.Repo.CreatePublishAudit(eventID, channel, status, details)
}
