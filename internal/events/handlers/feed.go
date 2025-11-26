package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *EventHandler) ListEvents(c *gin.Context) {

	// ---- ETag Check (Client Cache Validation) ----
	clientEtag := c.GetHeader("If-None-Match")
	// Trim any surrounding quotes or whitespace the client may send
	clientEtag = strings.TrimSpace(strings.Trim(clientEtag, `"`))

	currentVersion, err := h.Service.GetFeedVersion()
	if err != nil {
		log.Printf("ListEvents: GetFeedVersion error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read feed version"})
		return
	}

	etag := strconv.Itoa(currentVersion)
	log.Printf("ListEvents: clientIfNoneMatch=%q currentVersion=%d etag=%q\n", clientEtag, currentVersion, etag)

	// If client ETag matches server version → no need to send data
	if clientEtag != "" && clientEtag == etag {
		// respond 304 (no body)
		c.Status(http.StatusNotModified)
		return
	}

	// ---- Parse Query Params ----
	var query EventFeedQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params"})
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	// Since filter
	var parsedSince *time.Time
	if query.Since != "" {
		t, err := time.Parse(time.RFC3339, query.Since)
		if err == nil {
			parsedSince = &t
		}
	}

	events, total, err := h.Service.GetEventFeed(query.Page, query.Size, parsedSince, query.Channel)
	if err != nil {
		log.Printf("ListEvents: GetEventFeed error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot load events"})
		return
	}

	// ---- Build Response ----
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

	// ---- Add ETag header before sending response ----
	// Use quoted value per RFC: ETag: "7"
	quoted := `"` + etag + `"`
	c.Header("ETag", quoted)

	// log what we returned
	log.Printf("ListEvents: returning %d events, ETag=%s\n", len(responseItems), quoted)

	// ---- Send Response ----
	c.JSON(http.StatusOK, EventFeedResponse{
		Page:   query.Page,
		Size:   query.Size,
		Total:  total,
		Events: responseItems,
	})
}
