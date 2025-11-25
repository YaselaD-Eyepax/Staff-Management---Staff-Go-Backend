package handlers

type CreateEventDTO struct {
    Title       string   `json:"title" binding:"required"`
    Summary     string   `json:"summary"`
    Body        string   `json:"body" binding:"required"`
    Attachments []any    `json:"attachments"`
    Tags        []string `json:"tags"`

    CreatedBy   string `json:"created_by" binding:"required"`
    ScheduledAt string `json:"scheduled_at"` // optional
}
