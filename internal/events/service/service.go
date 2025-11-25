package service

import (
    "events-service/internal/events/models"
    "events-service/internal/events/repository"

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
