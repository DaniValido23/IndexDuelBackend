package scheduler

import (
	"index-duel-backend/service"
	"log"
	"time"
)

type Scheduler struct {
	cardService *service.CardService
	ticker      *time.Ticker
	done        chan bool
}

func NewScheduler(cardService *service.CardService) *Scheduler {
	return &Scheduler{
		cardService: cardService,
		done:        make(chan bool),
	}
}

// Start begins the weekly card synchronization
func (s *Scheduler) Start() {
	// Run immediately on startup
	go func() {
		log.Println("Running initial card synchronization...")
		if err := s.cardService.FetchAndStoreAllCards(); err != nil {
			log.Printf("Error during initial sync: %v", err)
		} else {
			log.Println("Initial card synchronization completed")
		}
	}()

	// Schedule weekly updates (every 7 days)
	s.ticker = time.NewTicker(7 * 24 * time.Hour)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				log.Println("Running weekly card synchronization...")
				if err := s.cardService.FetchAndStoreAllCards(); err != nil {
					log.Printf("Error during weekly sync: %v", err)
				} else {
					log.Println("Weekly card synchronization completed")
				}
			case <-s.done:
				s.ticker.Stop()
				return
			}
		}
	}()

	log.Println("Card synchronization scheduler started (weekly updates)")
}

// Stop terminates the scheduler
func (s *Scheduler) Stop() {
	s.done <- true
}
