CREATE TABLE IF NOT EXISTS ticket_transactions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INT NOT NULL REFERENCES users(id),
    amount FLOAT NOT NULL,
    type VARCHAR(20) NOT NULL,
    ref_type VARCHAR(50),
    ref_id INT,
    note VARCHAR(500),
    date TIMESTAMP DEFAULT NOW()
);
