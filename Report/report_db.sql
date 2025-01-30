-- Reporting Service
CREATE database report_db;
USE report_db;

CREATE TABLE Reports (
    ReportID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    FilePath VARCHAR(255) NOT NULL,
    GeneratedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO Reports (UserID, FilePath) VALUES
(1, '/reports/user1_report1.pdf'),
(2, '/reports/user2_report1.pdf');
