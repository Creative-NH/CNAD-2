-- Self-Assessment Service
CREATE database self_assessment;
USE self_assessment;

CREATE TABLE Assessments (
    AssessmentID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    Date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    HealthQuestions TEXT NOT NULL,
    PhysicalTestResults TEXT,
    RiskLevel ENUM('low', 'moderate', 'high') DEFAULT 'low'
);


INSERT INTO Assessments (UserID, HealthQuestions, PhysicalTestResults) VALUES
(1, '{"dizziness": "no", "balance": "good", "falls": 0}', '{"reaction_time": 0.8, "balance_score": 85}'),
(1, '{"dizziness": "yes", "balance": "poor", "falls": 2}', '{"reaction_time": 1.5, "balance_score": 60}'),
(2, '{"dizziness": "yes", "balance": "poor", "falls": 5}', '{"reaction_time": 2.0, "balance_score": 40}');