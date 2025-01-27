#Create new database for user
CREATE database users;
USE users;

CREATE TABLE Users (
    UserID INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(100) NOT NULL,
    Email VARCHAR(100) UNIQUE NOT NULL,
    PasswordHash VARCHAR(255) NOT NULL,
    Role ENUM('senior', 'caregiver', 'admin') DEFAULT 'senior',
    DateOfBirth DATE,
    PhoneNumber VARCHAR(15),
    Address TEXT,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO Users (Name, Email, PasswordHash, Role, DateOfBirth, PhoneNumber, Address) VALUES
('Alice Smith', 'alice@example.com', 'hashedpassword1', 'senior', '1950-05-15', '1234567890', '123 Senior Street, Cityville'),
('Bob Johnson', 'bob@example.com', 'hashedpassword2', 'caregiver', '1975-10-20', '0987654321', '456 Caregiver Lane, Townsville'),
('Admin User', 'admin@example.com', 'hashedpassword3', 'admin', NULL, NULL, NULL);