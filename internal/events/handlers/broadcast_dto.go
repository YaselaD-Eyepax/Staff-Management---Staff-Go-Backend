package handlers

type ManualBroadcastRequest struct {
	Channels []string `json:"channels"` // optional, defaults to all
}
