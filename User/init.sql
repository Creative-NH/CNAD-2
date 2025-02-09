USE user_service;

-- Table for storing doctor information
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

INSERT INTO users (name, username, password,is_active)
SELECT * FROM (SELECT 'user', 'test1', '$2a$10$BDGmvFiisPO/QZsuxt8JAudpYi1M/BZBhU8k1qa4pPm64iJ58SirS', 1) AS tmp
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'test1'
) LIMIT 1;

INSERT INTO users (name, username, password,is_active)
SELECT * FROM (SELECT 'user', 'test2', '$2a$10$BDGmvFiisPO/QZsuxt8JAudpYi1M/BZBhU8k1qa4pPm64iJ58SirS', 1) AS tmp
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'test1'
) LIMIT 1;

INSERT INTO users (name, username, password,is_active)
SELECT * FROM (SELECT 'user', 'test3', '$2a$10$BDGmvFiisPO/QZsuxt8JAudpYi1M/BZBhU8k1qa4pPm64iJ58SirS', 1) AS tmp
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'test1'
) LIMIT 1;