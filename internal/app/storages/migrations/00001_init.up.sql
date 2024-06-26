BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS short_urls(
        id SERIAL PRIMARY KEY,
        short VARCHAR(255) NOT NULL,
        original TEXT NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        is_deleted BOOLEAN NOT NULL default FALSE,
        UNIQUE (original)
    );
COMMIT;