package models

type SyncRequest struct {
	LastUpdate string `json:"last_update"`
}

type SyncResponse struct {
	Cards      []Card `json:"cards"`
	LastUpdate string `json:"last_update"`
	TotalCards int    `json:"total_cards"`
}
