package service

import (
	"encoding/json"
	"fmt"
	"index-duel-backend/models"
	"index-duel-backend/repository"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// CardService handles card-related business logic
type CardService struct {
	repo   *repository.CardRepository
	client *http.Client
	apiURL string
}

// NewCardService creates a new card service
func NewCardService(repo *repository.CardRepository) *CardService {
	return &CardService{
		repo: repo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiURL: os.Getenv("API"),
	}
}

// FetchAndStoreAllCards fetches all cards from the API and stores them in the database
func (s *CardService) FetchAndStoreAllCards() error {
	if s.apiURL == "" {
		return fmt.Errorf("API environment variable is not set")
	}

	log.Printf("Fetching cards from API: %s", s.apiURL)

	resp, err := s.client.Get(s.apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch cards from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	log.Printf("Found %d cards to process", len(apiResponse.Data))

	// Process cards in batches to avoid overwhelming the system
	batchSize := 10
	for i := 0; i < len(apiResponse.Data); i += batchSize {
		end := i + batchSize
		if end > len(apiResponse.Data) {
			end = len(apiResponse.Data)
		}

		batch := apiResponse.Data[i:end]
		log.Printf("Processing batch %d-%d of %d cards", i+1, end, len(apiResponse.Data))

		for _, card := range batch {
			if err := s.ProcessCard(&card); err != nil {
				log.Printf("Error processing card %d (%s): %v", card.ID, card.Name, err)
				// Continue processing other cards even if one fails
				continue
			}
			log.Printf("Successfully processed card: %s (ID: %d)", card.Name, card.ID)
		}

		// Add a small delay between batches to be respectful to image servers
		time.Sleep(1 * time.Second)
	}

	log.Printf("Completed processing all cards")
	return nil
}

// ProcessCard processes a single card, downloads images, and stores in database
func (s *CardService) ProcessCard(card *models.Card) error {
	// Download images for the card
	for i := range card.CardImages {
		image := &card.CardImages[i]

		// Download main image
		if image.ImageURL != "" {
			data, contentType, size, err := s.downloadImage(image.ImageURL)
			if err != nil {
				log.Printf("Failed to download main image for card %d: %v", card.ID, err)
			} else {
				image.ImageData = data
				image.ContentType = contentType
				image.FileSize = &size
			}
		}

		// Download small image
		if image.ImageURLSmall != "" {
			data, _, _, err := s.downloadImage(image.ImageURLSmall)
			if err != nil {
				log.Printf("Failed to download small image for card %d: %v", card.ID, err)
			} else {
				image.ImageSmallData = data
			}
		}

		// Download cropped image
		if image.ImageURLCropped != "" {
			data, _, _, err := s.downloadImage(image.ImageURLCropped)
			if err != nil {
				log.Printf("Failed to download cropped image for card %d: %v", card.ID, err)
			} else {
				image.ImageCroppedData = data
			}
		}
	}

	// Store the card in the database
	return s.repo.CreateCard(card)
}

// downloadImage downloads an image from a URL and returns its data
func (s *CardService) downloadImage(url string) ([]byte, string, int, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", 0, fmt.Errorf("image download returned status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to read image data: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Try to determine content type from URL extension
		if strings.HasSuffix(strings.ToLower(url), ".jpg") || strings.HasSuffix(strings.ToLower(url), ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(strings.ToLower(url), ".png") {
			contentType = "image/png"
		} else {
			contentType = "image/jpeg" // default
		}
	}

	return data, contentType, len(data), nil
}

// GetCardCount returns the total count of cards
func (s *CardService) GetCardCount() (int, error) {
	return s.repo.GetCardCount()
}

// SyncCards handles card synchronization requests from mobile app
func (s *CardService) SyncCards(lastUpdate string) ([]models.Card, string, error) {
	currentTime := time.Now().UTC().Format(time.RFC3339)

	var cards []models.Card
	var err error

	if lastUpdate == "" {
		// New client - send all cards
		log.Println("New client detected, sending all cards")
		cards, err = s.repo.GetAllCardsForFirstSync()
		if err != nil {
			return nil, "", fmt.Errorf("failed to get all cards: %w", err)
		}
	} else {
		// Existing client - send only updated cards
		log.Printf("Existing client detected, sending cards updated after: %s", lastUpdate)
		cards, err = s.repo.GetCardsUpdatedAfter(lastUpdate)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get updated cards: %w", err)
		}
	}

	log.Printf("Returning %d cards to client", len(cards))
	return cards, currentTime, nil
}
