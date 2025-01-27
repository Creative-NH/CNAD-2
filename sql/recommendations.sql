-- recommendations Service
CREATE database recommendations;
USE recommendations;

CREATE TABLE Recommendations (
    RecommendationID INT AUTO_INCREMENT PRIMARY KEY,
    RiskLevel ENUM('low', 'moderate', 'high') NOT NULL,
    Advice TEXT NOT NULL
);

INSERT INTO Recommendations (RiskLevel, Advice) VALUES
('low', 'Encourage healthy habits such as regular exercise and annual check-ups.'),
('moderate', 'Suggest specific exercises to improve balance and make home modifications.'),
('high', 'Recommend immediate clinical assessment by a healthcare professional.');
