package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Database connection function
func connectDB(dbUser, dbPassword, dbHost, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func main() {
	// Get database connection details from environment variables
	dbUser, dbPassword, dbHost, dbName := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME")
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" {
		log.Fatal("Database credentials not fully set in environment variables")
	}

	// Get local port from environment variables
	localPort := os.Getenv("LOCAL_PORT")
	if localPort == "" {
		log.Fatal("LOCAL_PORT environment variable is not set")
	}

	// Connect to the database
	db, err := connectDB(dbUser, dbPassword, dbHost, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the router
	router := mux.NewRouter()

	// API Routes
	router.HandleFunc("/api/questionnaire", func(w http.ResponseWriter, r *http.Request) {
		questionnaireHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		latestResultsHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/assessmentHistory", func(w http.ResponseWriter, r *http.Request) {
		assessmentHistoryHandler(w, r, db)
	}).Methods("{POST}")
	router.HandleFunc("/api/getResults", func(w http.ResponseWriter, r *http.Request) {
		getResultsHandler(w, r, db)
	}).Methods("POST")

	// CORS Configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	log.Printf("Starting server on :%s", localPort)
	if err := http.ListenAndServe(":"+localPort, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Structure to handle assessment submissions
type AssessmentRequest struct {
	UserID              int    `json:"user_id"`
	HealthQuestions     string `json:"health_questions"`
	PhysicalTestResults string `json:"physical_test_results"`
}

// Handler to retrieve questionnaire questions based on language (GET request with query string)
func questionnaireHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get language from query parameters
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "English" // Default to English if not specified
	}

	// Map supported languages to their respective tables
	languageMap := map[string]string{
		"English": "QuestionsEn",
		"Chinese": "QuestionsCn",
		"Malay":   "QuestionsMy",
		"Tamil":   "QuestionsTa",
	}

	// Check if language is supported
	tableName, exists := languageMap[language]
	if !exists {
		http.Error(w, "Unsupported language", http.StatusBadRequest)
		return
	}

	// Query the database for questions in the selected language
	query := fmt.Sprintf("SELECT QuestionID, QuestionContent, QuestionOptions FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Store retrieved questions
	var questions []map[string]interface{}

	for rows.Next() {
		var questionID int
		var questionContent string
		var questionOptions string // Stored in JSON format

		if err := rows.Scan(&questionID, &questionContent, &questionOptions); err != nil {
			http.Error(w, "Data retrieval error", http.StatusInternalServerError)
			return
		}

		questions = append(questions, map[string]interface{}{
			"question_id":      questionID,
			"question_content": questionContent,
			"question_options": json.RawMessage(questionOptions), // Ensure JSON format
		})
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// Handler to retrieve the latest assessment results for a user
func latestResultsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	query := "SELECT AssessmentID, Date, HealthQuestions, PhysicalTestResults, RiskLevel FROM Assessments WHERE UserID = ? ORDER BY Date DESC LIMIT 1"
	row := db.QueryRow(query, userID)

	var assessmentID int
	var date time.Time
	var healthQuestions, physicalTestResults, riskLevel string

	err := row.Scan(&assessmentID, &date, &healthQuestions, &physicalTestResults, &riskLevel)
	if err == sql.ErrNoRows {
		http.Error(w, "No records found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Respond with assessment details
	response := map[string]interface{}{
		"assessment_id":         assessmentID,
		"date":                  date,
		"health_questions":      healthQuestions,
		"physical_test_results": physicalTestResults,
		"risk_level":            riskLevel,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler to retrieve a user's full assessment history
func assessmentHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	query := "SELECT AssessmentID, Date, HealthQuestions, PhysicalTestResults, RiskLevel FROM Assessments WHERE UserID = ? ORDER BY Date DESC"
	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Store multiple assessments
	var assessments []map[string]interface{}
	for rows.Next() {
		var assessmentID int
		var date time.Time
		var healthQuestions, physicalTestResults, riskLevel string

		if err := rows.Scan(&assessmentID, &date, &healthQuestions, &physicalTestResults, &riskLevel); err != nil {
			http.Error(w, "Data retrieval error", http.StatusInternalServerError)
			return
		}

		assessments = append(assessments, map[string]interface{}{
			"assessment_id":         assessmentID,
			"date":                  date,
			"health_questions":      healthQuestions,
			"physical_test_results": physicalTestResults,
			"risk_level":            riskLevel,
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessments)
}

// Handler to retrieve a specific assessment result based on assessment_id
func getResultsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse assessment ID from query parameters
	assessmentID := r.URL.Query().Get("assessment_id")
	if assessmentID == "" {
		http.Error(w, "Assessment ID is required", http.StatusBadRequest)
		return
	}

	// Query the database for the specific assessment
	query := "SELECT AssessmentID, UserID, Date, HealthQuestions, PhysicalTestResults, RiskLevel FROM Assessments WHERE AssessmentID = ?"
	row := db.QueryRow(query, assessmentID)

	var assessment struct {
		AssessmentID        int       `json:"assessment_id"`
		UserID              int       `json:"user_id"`
		Date                time.Time `json:"date"`
		HealthQuestions     string    `json:"health_questions"`
		PhysicalTestResults string    `json:"physical_test_results"`
		RiskLevel           string    `json:"risk_level"`
	}

	err := row.Scan(&assessment.AssessmentID, &assessment.UserID, &assessment.Date, &assessment.HealthQuestions, &assessment.PhysicalTestResults, &assessment.RiskLevel)
	if err == sql.ErrNoRows {
		http.Error(w, "Assessment not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessment)
}
