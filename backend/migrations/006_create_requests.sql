CREATE TABLE IF NOT EXISTS requests (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INT NOT NULL REFERENCES users(id),
    novel_title VARCHAR(500) NOT NULL,
    novel_url VARCHAR(1000),
    source VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    votes INT DEFAULT 0
);
