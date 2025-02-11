package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	//localPort := os.Getenv("LOCAL_PORT")
	localPort := "5000"
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
	router.HandleFunc("/api/addAssessmentResults", func(w http.ResponseWriter, r *http.Request) {
		addAssessmentHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/getLastAssessment", func(w http.ResponseWriter, r *http.Request) {
		getLastAssessmentHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/assessmentHistory", func(w http.ResponseWriter, r *http.Request) {
		assessmentHistoryHandler(w, r, db)
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

type Assessment struct {
	AssessmentID   int    `json:"id"`
	TotalScore     int    `json:"totalScore"`
	RiskLevel      string `json:"riskLevel"`
	Recommendation string `json:"recommendation"`
	DateCreated    int    `json:"dateCreated"`
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
			log.Println("Data retrieval error: ", err)
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

func addAssessmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		UserID  int         `json:"user_id"`
		Answers map[int]int `json:"answers"`
	}

	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding JSON request:", err)
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate User ID
	if req.UserID <= 0 {
		log.Println("Invalid UserID:", req.UserID)
		http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
		return
	}

	// Convert answers to JSON format
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		log.Println("Error marshalling answers JSON:", err)
		http.Error(w, "Failed to process answers", http.StatusInternalServerError)
		return
	}

	// Prepare JSON body for Risk Assessment Service
	riskRequestBody, err := json.Marshal(req)
	if err != nil {
		log.Println("Error encoding risk assessment request:", err)
		http.Error(w, "Failed to encode risk assessment request", http.StatusInternalServerError)
		return
	}

	// Call Risk Assessment Service
	riskResponse, err := http.Post("http://localhost:8080/api/analyzeRisk", "application/json", bytes.NewBuffer(riskRequestBody))
	if err != nil {
		log.Println("Error calling Risk Assessment Service:", err)
		http.Error(w, "Failed to process risk assessment", http.StatusInternalServerError)
		return
	}
	defer riskResponse.Body.Close()

	// Parse Risk Assessment Response
	var riskResult struct {
		TotalScore     int    `json:"total_score"`
		RiskLevel      string `json:"risk_level"`
		Recommendation string `json:"recommendation"`
	}

	if err := json.NewDecoder(riskResponse.Body).Decode(&riskResult); err != nil {
		log.Println("Error decoding risk assessment response:", err)
		http.Error(w, "Failed to parse risk assessment response", http.StatusInternalServerError)
		return
	}

	// Store results in the database
	insertQuery := `INSERT INTO Assessments (UserID, QuestionResponses, TotalScore, RiskLevel, Recommendation, DateCreated) 
                    VALUES (?, ?, ?, ?, ?, NOW())`
	result, err := db.Exec(insertQuery, req.UserID, answersJSON, riskResult.TotalScore, riskResult.RiskLevel, riskResult.Recommendation)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, "Failed to store assessment data", http.StatusInternalServerError)
		return
	}

	assessmentID, _ := result.LastInsertId()

	// Send response
	response := map[string]interface{}{
		"assessment_id":      assessmentID,
		"user_id":            req.UserID,
		"total_score":        riskResult.TotalScore,
		"risk_level":         riskResult.RiskLevel,
		"recommendation":     riskResult.Recommendation,
		"question_responses": req.Answers,
	}

	log.Println("Successfully stored assessment:", response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Retrieve results of last assessment
func getLastAssessmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		UserID int `json:"user_id"`
	}
	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid JSON request")
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate User ID
	if req.UserID <= 0 {
		log.Println("Invalid or missing User ID")
		http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
		return
	}

	// Query the database for the latest risk assessment for the user
	query := `SELECT AssessmentID, TotalScore, RiskLevel, Recommendation 
              FROM Assessments 
              WHERE UserID = ? 
              ORDER BY DateCreated DESC LIMIT 1`

	assessment := Assessment{}

	err := db.QueryRow(query, req.UserID).Scan(
		&assessment.AssessmentID,
		&assessment.TotalScore,
		&assessment.RiskLevel,
		&assessment.Recommendation,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No risk assessments found for this user", http.StatusNotFound)
		} else {
			log.Println("Database query error:", err)
			http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		}
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessment)
}

func assessmentHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		UserID int `json:"user_id"`
	}
	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid JSON request")
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate User ID
	if req.UserID <= 0 {
		log.Println("Invalid or missing User ID")
		http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
		return
	}

	// Query the database for all risk assessments for the user
	query := `SELECT AssessmentID, TotalScore, RiskLevel, Recommendation, DateCreated
              FROM Assessments 
              WHERE UserID = ? 
              ORDER BY CreatedAt DESC`

	rows, err := db.Query(query, req.UserID)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var assessments []Assessment

	// Iterate over rows
	for rows.Next() {
		var assessment Assessment
		if err := rows.Scan(&assessment.AssessmentID, &assessment.TotalScore, &assessment.RiskLevel, &assessment.Recommendation, &assessment.DateCreated); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to process data", http.StatusInternalServerError)
			return
		}
		assessments = append(assessments, assessment)
	}

	// Check if no records were found
	if len(assessments) == 0 {
		http.Error(w, "No risk assessments found for this user", http.StatusNotFound)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessments)
}
