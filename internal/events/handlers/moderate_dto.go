package handlers

type ModerateEventDTO struct {
	Action      string `json:"action" binding:"required"`       // approve | reject
	ModeratorID string `json:"moderator_id" binding:"required"` // uuid
	Notes       string `json:"notes"`
}
