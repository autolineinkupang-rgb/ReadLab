CREATE TABLE IF NOT EXISTS reading_histories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INT NOT NULL REFERENCES users(id),
    novel_id INT NOT NULL REFERENCES novels(id),
    chapter_id INT NOT NULL REFERENCES chapters(id),
    progress FLOAT DEFAULT 0,
    UNIQUE (user_id, novel_id, chapter_id)
);
