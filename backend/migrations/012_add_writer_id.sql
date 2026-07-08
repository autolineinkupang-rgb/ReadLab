ALTER TABLE novels ADD COLUMN writer_id INT REFERENCES users(id);
