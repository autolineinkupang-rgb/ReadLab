CREATE TABLE IF NOT EXISTS novels (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    title VARCHAR(500) NOT NULL,
    alt_title VARCHAR(500),
    slug VARCHAR(500) UNIQUE NOT NULL,
    author VARCHAR(200),
    author_slug VARCHAR(500),
    status VARCHAR(20) DEFAULT 'ongoing',
    views BIGINT DEFAULT 0,
    rating FLOAT DEFAULT 0,
    rating_count INT DEFAULT 0,
    chapters INT DEFAULT 0,
    readers INT DEFAULT 0,
    chars VARCHAR(20),
    ai_percent VARCHAR(10),
    description TEXT,
    cover_url VARCHAR(1000),
    requested_by VARCHAR(200),
    released_by VARCHAR(200),
    added_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS novel_genres (
    novel_id INT NOT NULL,
    genre_id INT NOT NULL,
    PRIMARY KEY (novel_id, genre_id)
);
