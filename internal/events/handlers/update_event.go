package handlers

import (
	"events-service/internal/events/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *EventHandler) UpdateEvent(c *gin.Context) {
    eventIDStr := c.Param("id")
    eventID, err := uuid.Parse(eventIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    var dto UpdateEventDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Prepare event updates
    eventUpdates := models.Event{
        ID: eventID,
    }

    if dto.Title != nil {
        eventUpdates.Title = *dto.Title
    }
    if dto.Summary != nil {
        eventUpdates.Summary = *dto.Summary
    }
    if dto.ScheduledAt != nil && *dto.ScheduledAt != "" {
        t, _ := time.Parse(time.RFC3339, *dto.ScheduledAt)
        eventUpdates.ScheduledAt = &t
    }

    // Body updates
    bodyUpdates := models.AnnouncementBody{
        EventID: eventID,
    }
    if dto.Body != nil {
        bodyUpdates.Body = *dto.Body
    }
    if dto.Attachments != nil {
        bodyUpdates.Attachments = []byte("[]") // you can convert real attachments later
    }

    // Tags
    var tags []models.EventTag
    for _, t := range dto.Tags {
        tags = append(tags, models.EventTag{
            EventID: eventID,
            Tag:     t,
        })
    }

    // Run update
    if err := h.Service.UpdateEvent(eventUpdates, bodyUpdates, tags); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to update event"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "event updated"})
}
