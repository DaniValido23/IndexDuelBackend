package handlers

import (
	"encoding/json"
	"fmt"
	"index-duel-backend/models"
	"index-duel-backend/service"
	"log"
	"net/http"
)

// CardHandler handles HTTP requests for cards
type CardHandler struct {
	cardService *service.CardService
}

// NewCardHandler creates a new card handler
func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{
		cardService: cardService,
	}
}

// HealthCheckHandler provides a health check endpoint
func (h *CardHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Get card count to verify database connectivity
	count, err := h.cardService.GetCardCount()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"status":      "healthy",
		"cards_count": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SyncCardsForMobileHandler handles synchronization requests from mobile app
func (h *CardHandler) SyncCardsForMobileHandler(w http.ResponseWriter, r *http.Request) {
	var syncRequest models.SyncRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&syncRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Sync request received with last_update: '%s'", syncRequest.LastUpdate)

	// Get cards based on last update timestamp
	cards, currentTime, err := h.cardService.SyncCards(syncRequest.LastUpdate)
	if err != nil {
		log.Printf("Error during sync: %v", err)
		http.Error(w, fmt.Sprintf("Failed to sync cards: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := models.SyncResponse{
		Cards:      cards,
		LastUpdate: currentTime,
		TotalCards: len(cards),
	}

	log.Printf("Sending %d cards to mobile client. New last_update: %s", len(cards), currentTime)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
