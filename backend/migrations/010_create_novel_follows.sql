CREATE TABLE IF NOT EXISTS novel_follows (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INT NOT NULL REFERENCES users(id),
    novel_id INT NOT NULL REFERENCES novels(id),
    UNIQUE (user_id, novel_id)
);
