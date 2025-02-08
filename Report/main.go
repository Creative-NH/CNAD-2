package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	userDB, err := sql.Open("mysql", "root:yourpassword@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		log.Fatal(err)
	}
	defer userDB.Close()

	selfAssessmentDB, err := sql.Open("mysql", "root:yourpassword@tcp(127.0.0.1:3306)/self_assessment_db")
	if err != nil {
		log.Fatal(err)
	}
	defer selfAssessmentDB.Close()

	riskAssessmentDB, err := sql.Open("mysql", "root:yourpassword@tcp(127.0.0.1:3306)/risk_assessment_db")
	if err != nil {
		log.Fatal(err)
	}
	defer riskAssessmentDB.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		displayReport(w, userDB, selfAssessmentDB, riskAssessmentDB)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type User struct {
	Name        string
	Email       string
	DateOfBirth string
	PhoneNumber string
	Address     string
}

type Assessment struct {
	PhysicalTestResults string
}

type RiskAssessment struct {
	RiskLevel      string
	Recommendation string
}

func displayReport(w http.ResponseWriter, userDB *sql.DB, selfAssessmentDB *sql.DB, riskAssessmentDB *sql.DB) {
	// Query user information
	user := User{}
	userRow := userDB.QueryRow("SELECT Name, Email, DateOfBirth, PhoneNumber, Address FROM Users WHERE UserID = 1")
	if err := userRow.Scan(&user.Name, &user.Email, &user.DateOfBirth, &user.PhoneNumber, &user.Address); err != nil {
		log.Printf("Error fetching user data: %v", err)
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Query assessment information
	assessment := Assessment{}
	assessmentRow := selfAssessmentDB.QueryRow("SELECT PhysicalTestResults FROM Assessments WHERE UserID = 1")
	if err := assessmentRow.Scan(&assessment.PhysicalTestResults); err != nil {
		log.Printf("Error fetching assessment data: %v", err)
		http.Error(w, "Error fetching assessment data", http.StatusInternalServerError)
		return
	}

	// Query risk assessment information
	riskAssessment := RiskAssessment{}
	riskRow := riskAssessmentDB.QueryRow("SELECT RiskLevel, Recommendation FROM RiskAssessments WHERE UserID = 1")
	if err := riskRow.Scan(&riskAssessment.RiskLevel, &riskAssessment.Recommendation); err != nil {
		log.Printf("Error fetching risk assessment data: %v", err)
		http.Error(w, "Error fetching risk assessment data", http.StatusInternalServerError)
		return
	}

	// Log the data
	log.Printf("User: %+v\n", user)
	log.Printf("Assessment: %+v\n", assessment)
	log.Printf("RiskAssessment: %+v\n", riskAssessment)

	// Load and execute the template
	tmpl, err := template.ParseFiles("../Front-End/static/report.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		User           User
		Assessment     Assessment
		RiskAssessment RiskAssessment
		ReportDate     string
	}{
		User:           user,
		Assessment:     assessment,
		RiskAssessment: riskAssessment,
		ReportDate:     time.Now().Format("January 2, 2006"),
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
