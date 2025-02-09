-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS risk_assessment_db;
USE risk_assessment_db;

-- Drop existing table if needed (Optional: Uncomment if you want to reset data)
-- DROP TABLE IF EXISTS RiskAssessments;

-- Create RiskAssessments table (Includes Recommendation column)
CREATE TABLE IF NOT EXISTS RiskAssessments (
    ID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    TotalScore INT NOT NULL,
    RiskLevel ENUM('Low', 'Moderate', 'High') NOT NULL,
    Recommendation TEXT NOT NULL,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample risk assessment data
INSERT INTO RiskAssessments (UserID, TotalScore, RiskLevel, Recommendation) VALUES
(1, 15, 'Moderate', 'Consider physical therapy, improve home safety, and monitor medications.'),
(2, 8, 'Low', 'Maintain a healthy lifestyle and exercise regularly.'),
(3, 18, 'High', 'Consult a healthcare provider for a fall risk assessment and use mobility aids.'),
(4, 5, 'Low', 'Keep an active lifestyle with light exercises and regular check-ups.'),
(5, 10, 'Moderate', 'Consider using walking aids and reviewing medication side effects.');
