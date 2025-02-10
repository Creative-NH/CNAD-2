CREATE DATABASE IF NOT EXISTS doctor_service;
USE doctor_service;

-- Table for storing doctor information
CREATE TABLE IF NOT EXISTS Doctors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- Changed to store hashed passwords
    available BOOLEAN DEFAULT TRUE
);

-- Table for storing alerts related to patient cases
CREATE TABLE IF NOT EXISTS Alerts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    alert_message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES Reports(user_id) ON DELETE CASCADE
);

-- Table for storing patient reports accessible by doctors
CREATE TABLE IF NOT EXISTS Reports (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    report TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Doctors(id) ON DELETE CASCADE
);

-- Insert a test doctor with a hashed password
INSERT INTO Doctors (name, username, password_hash, available) VALUES
('Dr. John Doe', 'johndoe', '$2a$12$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36wb8F2oUqJ36PpsW.pxUUm', TRUE);
