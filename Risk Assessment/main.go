package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Database instance
var db *sql.DB

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	dsn := "root:04D685362v98@tcp(127.0.0.1:3306)/risk_assessment_db"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Database connected successfully!")

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 10)

	return db, nil
}

// Get recommendations based on risk level
func getRecommendation(riskLevel string) string {
	switch riskLevel {
	case "Low":
		return "Maintain a healthy lifestyle with balance exercises and check-ups."
	case "Moderate":
		return "Consider physical therapy, improve home safety, and monitor medications."
	case "High":
		return "Consult a healthcare provider for a fall risk assessment and use mobility aids."
	default:
		return "No recommendation available."
	}
}

// Analyze risk (UserID now assigned from local storage, hardcoded as 6 for now)
func analyzeRisk(w http.ResponseWriter, r *http.Request) {
	var input map[string]int

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// UserID is hardcoded to 6 for now
	userID := 6

	// Assigning risk points based on answers
	riskPoints := map[int]int{
		1:  2, // Dizziness: Yes = 2, No = 0
		2:  1, // Balance: Good = 1, Moderate = 2, Poor = 3
		3:  1, // Falls: 0 = 1, 1-2 = 2, 3+ = 3
		4:  1, // Mobility aid: None = 0, Cane = 1, Walker = 2, Wheelchair = 3
		5:  1, // Unsteady Walking: Never = 0, Sometimes = 1, Often = 2, Always = 3
		6:  2, // Recent fall: Yes = 2, No = 0
		7:  1, // Stand without hands: Yes = 0, No = 2
		8:  1, // Medications: Yes = 2, No = 0, Not sure = 1
		9:  1, // Exercise: Yes = 0, No = 2
		10: 1, // Numbness: Yes = 2, No = 0, Sometimes = 1
	}

	totalScore := 0
	for qID, answer := range input {
		// Convert qID from string to int
		qIDInt, err := strconv.Atoi(qID)
		if err != nil {
			log.Println("Invalid question ID:", qID)
			continue
		}

		if points, exists := riskPoints[qIDInt]; exists {
			totalScore += points * answer
		}
	}

	// Determine risk level
	var riskLevel string
	switch {
	case totalScore <= 5:
		riskLevel = "Low"
	case totalScore <= 10:
		riskLevel = "Moderate"
	default:
		riskLevel = "High"
	}

	// Generate recommendation
	recommendation := getRecommendation(riskLevel)

	// Store in Database (Now includes recommendation)
	_, err := db.Exec(`INSERT INTO RiskAssessments (UserID, TotalScore, RiskLevel, Recommendation, CreatedAt) VALUES (?, ?, ?, ?, NOW())`,
		userID, totalScore, riskLevel, recommendation)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	response := map[string]interface{}{
		"user_id":        userID,
		"total_score":    totalScore,
		"risk_level":     riskLevel,
		"recommendation": recommendation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get last risk assessment for a user
func getUserLastRiskAssessment(w http.ResponseWriter, r *http.Request) {
	// UserID is hardcoded to 6 for now
	userID := 6

	// Declare variables for database fields
	var assessmentID int
	var totalScore int
	var riskLevel string
	var recommendation string
	var createdAt string

	// Query the database for the latest risk assessment for the user
	query := `SELECT ID, TotalScore, RiskLevel, Recommendation, CreatedAt 
              FROM RiskAssessments 
              WHERE UserID = ? 
              ORDER BY CreatedAt DESC LIMIT 1`

	err := db.QueryRow(query, userID).Scan(
		&assessmentID,
		&totalScore,
		&riskLevel,
		&recommendation,
		&createdAt,
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

	// Convert data to JSON and send response
	response := map[string]interface{}{
		"id":             assessmentID,
		"user_id":        userID,
		"total_score":    totalScore,
		"risk_level":     riskLevel,
		"recommendation": recommendation,
		"created_at":     createdAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	var err error
	db, err = InitDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/analyze-risk", analyzeRisk).Methods("POST")
	router.HandleFunc("/risk-history", getUserLastRiskAssessment).Methods("GET")

	// Enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins (for development only)
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler.Handler(router)))
}
