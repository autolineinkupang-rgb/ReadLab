CREATE TABLE IF NOT EXISTS news (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    title VARCHAR(500) NOT NULL,
    content TEXT,
    type VARCHAR(50) DEFAULT 'news',
    slug VARCHAR(500) UNIQUE
);
