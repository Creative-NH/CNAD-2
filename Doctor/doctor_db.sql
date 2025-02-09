CREATE DATABASE IF NOT EXISTS doctor_service;
USE doctor_service;

-- Table for storing doctor information
CREATE TABLE IF NOT EXISTS Doctors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    available BOOLEAN DEFAULT TRUE
);

-- Table for storing alerts related to patient cases
CREATE TABLE IF NOT EXISTS Alerts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    alert_message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE
);

-- Table for storing patient reports accessible by doctors
CREATE TABLE IF NOT EXISTS Reports (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    report TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert a test doctor for login testing
INSERT INTO Doctors (name, username, password, available) VALUES
('Dr. John Doe', 'johndoe', 'securepassword', TRUE);
