package handlers

type UpdateEventDTO struct {
	Title       *string  `json:"title"`
	Summary     *string  `json:"summary"`
	Body        *string  `json:"body"`
	Attachments []any    `json:"attachments"`
	Tags        []string `json:"tags"`
	ScheduledAt *string  `json:"scheduled_at"`
}
