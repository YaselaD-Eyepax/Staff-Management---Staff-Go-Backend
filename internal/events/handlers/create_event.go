package handlers

import (
	"events-service/internal/events/models"
	"events-service/internal/events/repository"
	"events-service/internal/events/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventHandler struct {
    Service *service.EventService
}

func NewEventHandler(db *gorm.DB) *EventHandler {
    repo := repository.NewEventRepository(db)
    svc := service.NewEventService(repo)
    return &EventHandler{Service: svc}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
    var dto CreateEventDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    eventID := uuid.New()
    bodyID := uuid.New()

    var scheduledAt *time.Time
    if dto.ScheduledAt != "" {
        t, _ := time.Parse(time.RFC3339, dto.ScheduledAt)
        scheduledAt = &t
    }

    createdBy, err := uuid.Parse(dto.CreatedBy)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid created_by"})
        return
    }

    event := models.Event{
        ID:          eventID,
        Title:       dto.Title,
        Summary:     dto.Summary,
        CreatedBy:   createdBy,
        Status:      "draft",
        ScheduledAt: scheduledAt,
    }

    body := models.AnnouncementBody{
        ID:          bodyID,
        EventID:     eventID,
        Body:        dto.Body,
        Attachments: []byte("[]"),
    }

    var tags []models.EventTag
    for _, t := range dto.Tags {
        tags = append(tags, models.EventTag{
            EventID: eventID,
            Tag:     t,
        })
    }

    if err := h.Service.CreateEvent(event, body, tags); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": eventID})
}
