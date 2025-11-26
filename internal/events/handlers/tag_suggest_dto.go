package handlers

type TagSuggestRequest struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Body    string `json:"body"`
}

type TagSuggestResponse struct {
	Tags []string `json:"tags"`
}
