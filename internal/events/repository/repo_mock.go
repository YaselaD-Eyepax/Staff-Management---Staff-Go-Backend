package repository

import (
	"events-service/internal/events/models"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockEventRepo struct {
    mock.Mock
}

func (m *MockEventRepo) CreateEvent(e *models.Event, b *models.AnnouncementBody, t []models.EventTag) error {
    args := m.Called(e, b, t)
    return args.Error(0)
}

func (m *MockEventRepo) GetEvent(id uuid.UUID) (*models.Event, error) {
    args := m.Called(id)
    return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventRepo) GetEventFeed(page int, size int, since *time.Time, channel string) ([]models.Event, int64, error) {
    args := m.Called(page, size, since, channel)
    return args.Get(0).([]models.Event), args.Get(1).(int64), args.Error(2)
}

func (m *MockEventRepo) UpdateEvent(e *models.Event, b *models.AnnouncementBody, t []models.EventTag) error {
    args := m.Called(e, b, t)
    return args.Error(0)
}

func (m *MockEventRepo) ModerateEvent(eventID uuid.UUID, status string, moderatorID uuid.UUID, notes string) error {
    args := m.Called(eventID, status, moderatorID, notes)
    return args.Error(0)
}

func (m *MockEventRepo) GetFeedVersion() (int, error) {
    args := m.Called()
    return args.Int(0), args.Error(1)
}

func (m *MockEventRepo) IncrementFeedVersion() error {
    args := m.Called()
    return args.Error(0)
}
