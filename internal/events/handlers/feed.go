package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *EventHandler) ListEvents(c *gin.Context) {
	var query EventFeedQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params"})
		return
	}

	// defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	// since filter
	var parsedSince *time.Time
	if query.Since != "" {
		t, err := time.Parse(time.RFC3339, query.Since)
		if err == nil {
			parsedSince = &t
		}
	}

	events, total, err := h.Service.GetEventFeed(query.Page, query.Size, parsedSince, query.Channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load events"})
		return
	}

	// Prepare response list
	responseItems := make([]EventFeedItem, 0, len(events))

	for _, e := range events {
		tags := make([]string, 0)
		for _, t := range e.Tags {
			tags = append(tags, t.Tag)
		}

		var scheduled *string
		if e.ScheduledAt != nil {
			s := e.ScheduledAt.Format(time.RFC3339)
			scheduled = &s
		}

		responseItems = append(responseItems, EventFeedItem{
			ID:          e.ID.String(),
			Title:       e.Title,
			Summary:     e.Summary,
			Status:      e.Status,
			ScheduledAt: scheduled,
			CreatedAt:   e.CreatedAt.Format(time.RFC3339),
			Tags:        tags,
		})
	}

	c.JSON(http.StatusOK, EventFeedResponse{
		Page:   query.Page,
		Size:   query.Size,
		Total:  total,
		Events: responseItems,
	})
}
