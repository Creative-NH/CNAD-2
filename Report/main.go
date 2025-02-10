package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

func main() {
	userDB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		log.Fatal(err)
	}
	defer userDB.Close()

	selfAssessmentDB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/self_assessment_db")
	if err != nil {
		log.Fatal(err)
	}
	defer selfAssessmentDB.Close()

	riskAssessmentDB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/risk_assessment_db")
	if err != nil {
		log.Fatal(err)
	}
	defer riskAssessmentDB.Close()

	reportDB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/report_db")
	if err != nil {
		log.Fatal(err)
	}
	defer reportDB.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		displayReport(w, r, userDB, selfAssessmentDB, riskAssessmentDB, reportDB)
	})
	http.HandleFunc("/login", loginHandler)
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	// For demonstration purposes, we set the UserID to 1
	session.Values["UserID"] = 1
	session.Save(r, w)
	fmt.Fprintln(w, "Logged in")
}

func displayReport(w http.ResponseWriter, r *http.Request, userDB *sql.DB, selfAssessmentDB *sql.DB, riskAssessmentDB *sql.DB, reportDB *sql.DB) {
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["UserID"].(int)
	if !ok {
		userID = 1 // Default to UserID 1 if not logged in
	}

	// Query user information
	user := User{}
	userRow := userDB.QueryRow("SELECT Name, Email, DateOfBirth, PhoneNumber, Address FROM Users WHERE UserID = ?", userID)
	if err := userRow.Scan(&user.Name, &user.Email, &user.DateOfBirth, &user.PhoneNumber, &user.Address); err != nil {
		log.Printf("Error fetching user data: %v", err)
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Query assessment information
	assessment := Assessment{}
	assessmentRow := selfAssessmentDB.QueryRow("SELECT PhysicalTestResults FROM Assessments WHERE UserID = ?", userID)
	if err := assessmentRow.Scan(&assessment.PhysicalTestResults); err != nil {
		log.Printf("Error fetching assessment data: %v", err)
		http.Error(w, "Error fetching assessment data", http.StatusInternalServerError)
		return
	}

	// Query risk assessment information
	riskAssessment := RiskAssessment{}
	riskRow := riskAssessmentDB.QueryRow("SELECT RiskLevel, Recommendation FROM RiskAssessments WHERE UserID = ?", userID)
	if err := riskRow.Scan(&riskAssessment.RiskLevel, &riskAssessment.Recommendation); err != nil {
		log.Printf("Error fetching risk assessment data: %v", err)
		http.Error(w, "Error fetching risk assessment data", http.StatusInternalServerError)
		return
	}

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
		return
	}

	// Insert new report into Reports table
	filePath := fmt.Sprintf("/reports/user%d_report%d.pdf", userID, time.Now().Unix()) // Example file path
	_, err = reportDB.Exec("INSERT INTO Reports (UserID, FilePath) VALUES (?, ?)", userID, filePath)
	if err != nil {
		log.Printf("Error inserting report: %v", err)
		http.Error(w, "Error inserting report", http.StatusInternalServerError)
		return
	}
	log.Printf("Inserted new report: UserID=%d, FilePath=%s", userID, filePath)
}
