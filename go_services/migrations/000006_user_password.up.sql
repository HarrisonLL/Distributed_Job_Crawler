ALTER TABLE users
ADD COLUMN password VARCHAR(255) NULL;

UPDATE users
SET password = '$2a$10$AmY6I.DQhJj2J0.Sy4X4yuZFsqlaIbN9Ij8qbHdbz3RpUAJ0uSXK6'
WHERE password IS NULL OR password = '';
