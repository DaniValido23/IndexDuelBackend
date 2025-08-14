CREATE TABLE IF NOT EXISTS cards (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    frame_type VARCHAR(50),
    description TEXT,
    atk INTEGER,
    def INTEGER,
    level INTEGER,
    race VARCHAR(100),
    attribute VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS card_sets (
    id SERIAL PRIMARY KEY,
    card_id BIGINT REFERENCES cards(id) ON DELETE CASCADE,
    set_name VARCHAR(255) NOT NULL,
    set_code VARCHAR(50) NOT NULL,
    set_rarity VARCHAR(100),
    set_rarity_code VARCHAR(20),
    set_price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS card_images (
    id SERIAL PRIMARY KEY,
    card_id BIGINT REFERENCES cards(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    image_url_small TEXT,
    image_url_cropped TEXT,
    image_data BYTEA,
    image_small_data BYTEA,
    image_cropped_data BYTEA,
    content_type VARCHAR(50) DEFAULT 'image/jpeg',
    file_size INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS card_prices (
    id SERIAL PRIMARY KEY,
    card_id BIGINT REFERENCES cards(id) ON DELETE CASCADE,
    cardmarket_price DECIMAL(10, 2),
    tcgplayer_price DECIMAL(10, 2),
    ebay_price DECIMAL(10, 2),
    amazon_price DECIMAL(10, 2),
    coolstuffinc_price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cards_name ON cards(name);
CREATE INDEX IF NOT EXISTS idx_cards_type ON cards(type);
CREATE INDEX IF NOT EXISTS idx_cards_race ON cards(race);
CREATE INDEX IF NOT EXISTS idx_cards_attribute ON cards(attribute);
CREATE INDEX IF NOT EXISTS idx_card_sets_card_id ON card_sets(card_id);
CREATE INDEX IF NOT EXISTS idx_card_images_card_id ON card_images(card_id);
CREATE INDEX IF NOT EXISTS idx_card_prices_card_id ON card_prices(card_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_cards_updated_at BEFORE UPDATE ON cards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_card_prices_updated_at BEFORE UPDATE ON card_prices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();