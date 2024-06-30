BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS short_urls(
        id SERIAL PRIMARY KEY,
        short VARCHAR(255) NOT NULL,
        original TEXT NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        is_deleted BOOLEAN default FALSE
    );
CREATE UNIQUE INDEX idx_unique_original ON short_urls(original) WHERE is_deleted = FALSE;
CREATE UNIQUE INDEX idx_unique_short ON short_urls(short) WHERE is_deleted = FALSE;

COMMIT;