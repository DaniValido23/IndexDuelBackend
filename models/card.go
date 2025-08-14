package models

import (
	"time"
)

type Card struct {
	ID          int64       `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Type        string      `json:"type" db:"type"`
	FrameType   string      `json:"frameType" db:"frame_type"`
	Description string      `json:"desc" db:"description"`
	ATK         *int        `json:"atk" db:"atk"`
	DEF         *int        `json:"def" db:"def"`
	Level       *int        `json:"level" db:"level"`
	Race        string      `json:"race" db:"race"`
	Attribute   string      `json:"attribute" db:"attribute"`
	CardSets    []CardSet   `json:"card_sets"`
	CardImages  []CardImage `json:"card_images"`
	CardPrices  []CardPrice `json:"card_prices"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

type CardSet struct {
	ID            int       `json:"id" db:"id"`
	CardID        int64     `json:"-" db:"card_id"`
	SetName       string    `json:"set_name" db:"set_name"`
	SetCode       string    `json:"set_code" db:"set_code"`
	SetRarity     string    `json:"set_rarity" db:"set_rarity"`
	SetRarityCode string    `json:"set_rarity_code" db:"set_rarity_code"`
	SetPrice      *string   `json:"set_price" db:"set_price"`
	CreatedAt     time.Time `db:"created_at"`
}

type CardImage struct {
	ID               int       `json:"id" db:"id"`
	CardID           int64     `json:"-" db:"card_id"`
	ImageURL         string    `json:"image_url" db:"image_url"`
	ImageURLSmall    string    `json:"image_url_small" db:"image_url_small"`
	ImageURLCropped  string    `json:"image_url_cropped" db:"image_url_cropped"`
	ImageData        []byte    `json:"-" db:"image_data"`
	ImageSmallData   []byte    `json:"-" db:"image_small_data"`
	ImageCroppedData []byte    `json:"-" db:"image_cropped_data"`
	ContentType      string    `json:"-" db:"content_type"`
	FileSize         *int      `json:"-" db:"file_size"`
	CreatedAt        time.Time `db:"created_at"`
}

type CardPrice struct {
	ID                int       `json:"id" db:"id"`
	CardID            int64     `json:"-" db:"card_id"`
	CardMarketPrice   *string   `json:"cardmarket_price" db:"cardmarket_price"`
	TCGPlayerPrice    *string   `json:"tcgplayer_price" db:"tcgplayer_price"`
	EbayPrice         *string   `json:"ebay_price" db:"ebay_price"`
	AmazonPrice       *string   `json:"amazon_price" db:"amazon_price"`
	CoolStuffIncPrice *string   `json:"coolstuffinc_price" db:"coolstuffinc_price"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

type APIResponse struct {
	Data []Card `json:"data"`
}
