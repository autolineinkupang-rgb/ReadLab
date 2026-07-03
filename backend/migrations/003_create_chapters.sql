CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    novel_id INT NOT NULL REFERENCES novels(id),
    number INT NOT NULL,
    title VARCHAR(500),
    content TEXT,
    is_locked BOOLEAN DEFAULT FALSE,
    ticket_cost INT DEFAULT 0
);

CREATE INDEX idx_chapters_novel_id ON chapters(novel_id);
CREATE INDEX idx_chapters_number ON chapters(number);
