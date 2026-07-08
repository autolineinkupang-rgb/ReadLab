ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'member';
UPDATE users SET role = 'admin' WHERE is_admin = TRUE;
UPDATE users SET role = 'member' WHERE is_admin = FALSE OR is_admin IS NULL;
ALTER TABLE users DROP COLUMN is_admin;
