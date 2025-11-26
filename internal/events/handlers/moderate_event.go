package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *EventHandler) ModerateEvent(c *gin.Context) {
    eventIDStr := c.Param("id")
    eventID, err := uuid.Parse(eventIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
        return
    }

    var dto ModerateEventDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if dto.Action != "approve" && dto.Action != "reject" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "action must be approve or reject"})
        return
    }

    moderator, err := uuid.Parse(dto.ModeratorID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid moderator_id"})
        return
    }

    var status string
    if dto.Action == "approve" {
        status = "approved"
    } else {
        status = "rejected"
    }

    if err := h.Service.ModerateEvent(eventID, status, moderator, dto.Notes); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "moderation failed"})
        return
    }

    // ETag bump
    _ = h.Service.IncrementFeedVersion()

    c.JSON(http.StatusOK, gin.H{
        "message": "event moderated",
        "status":  status,
    })
}
