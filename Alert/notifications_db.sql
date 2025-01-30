-- Notifications and Alerts Service
CREATE database notifications_db;
USE notifications_db;

CREATE TABLE Notifications (
    NotificationID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    Message TEXT NOT NULL,
    SentAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO Notifications (UserID, Message) VALUES
(1, 'Your assessment indicates a moderate risk. Please follow the recommended exercises.'),
(2, 'High risk detected. Please contact your healthcare provider immediately.');