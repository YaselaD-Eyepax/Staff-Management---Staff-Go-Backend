package handlers

// Incoming query params
type EventFeedQuery struct {
	Page    int    `form:"page"`
	Size    int    `form:"size"`
	Since   string `form:"since"`
	Channel string `form:"channel"`
}

// Outgoing event list item
type EventFeedItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Status      string   `json:"status"`
	ScheduledAt *string  `json:"scheduled_at"`
	CreatedAt   string   `json:"created_at"`
	Tags        []string `json:"tags"`
}

// Final paginated response
type EventFeedResponse struct {
	Page   int             `json:"page"`
	Size   int             `json:"size"`
	Total  int64           `json:"total"`
	Events []EventFeedItem `json:"events"`
}
