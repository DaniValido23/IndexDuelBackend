package repository

import (
	"database/sql"
	"fmt"
	"index-duel-backend/database"
	"index-duel-backend/models"
)

type CardRepository struct {
	db *database.DB
}

func NewCardRepository(db *database.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) CreateCard(card *models.Card) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO cards (id, name, type, frame_type, description, atk, def, level, race, attribute)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			frame_type = EXCLUDED.frame_type,
			description = EXCLUDED.description,
			atk = EXCLUDED.atk,
			def = EXCLUDED.def,
			level = EXCLUDED.level,
			race = EXCLUDED.race,
			attribute = EXCLUDED.attribute,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = tx.Exec(query, card.ID, card.Name, card.Type, card.FrameType, card.Description,
		card.ATK, card.DEF, card.Level, card.Race, card.Attribute)
	if err != nil {
		return fmt.Errorf("failed to insert card: %w", err)
	}

	if err := r.deleteCardRelatedData(tx, card.ID); err != nil {
		return fmt.Errorf("failed to delete existing related data: %w", err)
	}

	for _, set := range card.CardSets {
		if err := r.insertCardSet(tx, card.ID, &set); err != nil {
			return fmt.Errorf("failed to insert card set: %w", err)
		}
	}

	for _, image := range card.CardImages {
		if err := r.insertCardImage(tx, card.ID, &image); err != nil {
			return fmt.Errorf("failed to insert card image: %w", err)
		}
	}

	for _, price := range card.CardPrices {
		if err := r.insertCardPrice(tx, card.ID, &price); err != nil {
			return fmt.Errorf("failed to insert card price: %w", err)
		}
	}

	return tx.Commit()
}

func (r *CardRepository) deleteCardRelatedData(tx *sql.Tx, cardID int64) error {
	tables := []string{"card_sets", "card_images", "card_prices"}
	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE card_id = $1", table), cardID)
		if err != nil {
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}
	return nil
}

func (r *CardRepository) insertCardSet(tx *sql.Tx, cardID int64, set *models.CardSet) error {
	query := `
		INSERT INTO card_sets (card_id, set_name, set_code, set_rarity, set_rarity_code, set_price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(query, cardID, set.SetName, set.SetCode, set.SetRarity, set.SetRarityCode, set.SetPrice)
	return err
}

func (r *CardRepository) insertCardImage(tx *sql.Tx, cardID int64, image *models.CardImage) error {
	query := `
		INSERT INTO card_images (card_id, image_url, image_url_small, image_url_cropped, 
								image_data, image_small_data, image_cropped_data, content_type, file_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := tx.Exec(query, cardID, image.ImageURL, image.ImageURLSmall, image.ImageURLCropped,
		image.ImageData, image.ImageSmallData, image.ImageCroppedData, image.ContentType, image.FileSize)
	return err
}

func (r *CardRepository) insertCardPrice(tx *sql.Tx, cardID int64, price *models.CardPrice) error {
	query := `
		INSERT INTO card_prices (card_id, cardmarket_price, tcgplayer_price, ebay_price, amazon_price, coolstuffinc_price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(query, cardID, price.CardMarketPrice, price.TCGPlayerPrice,
		price.EbayPrice, price.AmazonPrice, price.CoolStuffIncPrice)
	return err
}

func (r *CardRepository) GetCard(cardID int64) (*models.Card, error) {
	card := &models.Card{}
	query := `
		SELECT id, name, type, frame_type, description, atk, def, level, race, attribute, created_at, updated_at
		FROM cards WHERE id = $1
	`
	err := r.db.QueryRow(query, cardID).Scan(
		&card.ID, &card.Name, &card.Type, &card.FrameType, &card.Description,
		&card.ATK, &card.DEF, &card.Level, &card.Race, &card.Attribute,
		&card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	if err := r.loadCardSets(card); err != nil {
		return nil, fmt.Errorf("failed to load card sets: %w", err)
	}
	if err := r.loadCardImages(card); err != nil {
		return nil, fmt.Errorf("failed to load card images: %w", err)
	}
	if err := r.loadCardPrices(card); err != nil {
		return nil, fmt.Errorf("failed to load card prices: %w", err)
	}

	return card, nil
}

func (r *CardRepository) loadCardSets(card *models.Card) error {
	query := `SELECT id, set_name, set_code, set_rarity, set_rarity_code, set_price, created_at 
			 FROM card_sets WHERE card_id = $1`
	rows, err := r.db.Query(query, card.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		set := models.CardSet{}
		err := rows.Scan(&set.ID, &set.SetName, &set.SetCode, &set.SetRarity,
			&set.SetRarityCode, &set.SetPrice, &set.CreatedAt)
		if err != nil {
			return err
		}
		card.CardSets = append(card.CardSets, set)
	}
	return rows.Err()
}

func (r *CardRepository) loadCardImages(card *models.Card) error {
	query := `SELECT id, image_url, image_url_small, image_url_cropped, content_type, file_size, created_at 
			 FROM card_images WHERE card_id = $1`
	rows, err := r.db.Query(query, card.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		image := models.CardImage{}
		err := rows.Scan(&image.ID, &image.ImageURL, &image.ImageURLSmall, &image.ImageURLCropped,
			&image.ContentType, &image.FileSize, &image.CreatedAt)
		if err != nil {
			return err
		}
		card.CardImages = append(card.CardImages, image)
	}
	return rows.Err()
}

func (r *CardRepository) loadCardPrices(card *models.Card) error {
	query := `SELECT id, cardmarket_price, tcgplayer_price, ebay_price, amazon_price, coolstuffinc_price, created_at, updated_at 
			 FROM card_prices WHERE card_id = $1`
	rows, err := r.db.Query(query, card.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		price := models.CardPrice{}
		err := rows.Scan(&price.ID, &price.CardMarketPrice, &price.TCGPlayerPrice,
			&price.EbayPrice, &price.AmazonPrice, &price.CoolStuffIncPrice, &price.CreatedAt, &price.UpdatedAt)
		if err != nil {
			return err
		}
		card.CardPrices = append(card.CardPrices, price)
	}
	return rows.Err()
}

// GetCardCount returns the total number of cards
func (r *CardRepository) GetCardCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM cards").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get card count: %w", err)
	}
	return count, nil
}

// GetCardsUpdatedAfter retrieves cards updated after the given timestamp
func (r *CardRepository) GetCardsUpdatedAfter(lastUpdate string) ([]models.Card, error) {
	query := `
		SELECT id, name, type, frame_type, description, atk, def, level, race, attribute, created_at, updated_at
		FROM cards 
		WHERE updated_at > $1 OR created_at > $1
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(query, lastUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated cards: %w", err)
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		card := models.Card{}
		err := rows.Scan(
			&card.ID, &card.Name, &card.Type, &card.FrameType, &card.Description,
			&card.ATK, &card.DEF, &card.Level, &card.Race, &card.Attribute,
			&card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// Load related data for each card
		if err := r.loadCardSets(&card); err != nil {
			return nil, fmt.Errorf("failed to load card sets: %w", err)
		}
		if err := r.loadCardImages(&card); err != nil {
			return nil, fmt.Errorf("failed to load card images: %w", err)
		}
		if err := r.loadCardPrices(&card); err != nil {
			return nil, fmt.Errorf("failed to load card prices: %w", err)
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// GetAllCardsForFirstSync retrieves all cards for new clients
func (r *CardRepository) GetAllCardsForFirstSync() ([]models.Card, error) {
	query := `
		SELECT id, name, type, frame_type, description, atk, def, level, race, attribute, created_at, updated_at
		FROM cards 
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all cards: %w", err)
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		card := models.Card{}
		err := rows.Scan(
			&card.ID, &card.Name, &card.Type, &card.FrameType, &card.Description,
			&card.ATK, &card.DEF, &card.Level, &card.Race, &card.Attribute,
			&card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// Load related data for each card
		if err := r.loadCardSets(&card); err != nil {
			return nil, fmt.Errorf("failed to load card sets: %w", err)
		}
		if err := r.loadCardImages(&card); err != nil {
			return nil, fmt.Errorf("failed to load card images: %w", err)
		}
		if err := r.loadCardPrices(&card); err != nil {
			return nil, fmt.Errorf("failed to load card prices: %w", err)
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}
