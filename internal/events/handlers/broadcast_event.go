package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *EventHandler) ManualBroadcast(c *gin.Context) {
    idStr := c.Param("id")
    eventID, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
        return
    }

    // Bind request body
    var req ManualBroadcastRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // If no body provided → use defaults
        req.Channels = []string{}
    }

    // If no channels -> send to all
    if len(req.Channels) == 0 {
        req.Channels = []string{"fcm", "email", "teams"}
    }

    // fetch event details for payload
    evt, err := h.Service.GetEvent(eventID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
        return
    }

    // Prepare basic payload
    payload := map[string]any{
        "title":   evt.Title,
        "summary": evt.Summary,
    }

    // Enqueue jobs
    for _, ch := range req.Channels {
        _ = h.Service.EnqueueBroadcast(eventID, ch, payload)
    }

    // Increment feed ETag
    _ = h.Service.IncrementFeedVersion()

    // Write audit entry for manual broadcast trigger
    _ = h.Service.CreatePublishAudit(eventID, "manual", "triggered", map[string]any{
        "channels": req.Channels,
    })

    c.JSON(http.StatusOK, gin.H{
        "message":  "broadcast queued",
        "channels": req.Channels,
    })
}
