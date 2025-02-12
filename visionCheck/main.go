package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gopkg.in/gomail.v2"

	_ "github.com/go-sql-driver/mysql"
)

type VisionResult struct {
	UserID        int    `json:"UserID"`
	LeftEyeScore  int    `json:"LeftEyeScore"`
	RightEyeScore int    `json:"RightEyeScore"`
	Comments      string `json:"Comments"`
}

func main() {
	http.HandleFunc("/main.go", handlePostRequest)
	http.HandleFunc("/getLatestResult", getLatestResult)
	http.HandleFunc("/sendReportToDoctor", handleSendReportToDoctor) // Added email API endpoint

	log.Println("Server running on port 8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var result VisionResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:04D685362v98@tcp(127.0.0.1:3306)/risk_assessment_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `INSERT INTO visionResult (UserID, LeftEyeScore, RightEyeScore, Comments) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(query, result.UserID, result.LeftEyeScore, result.RightEyeScore, result.Comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if report should be sent to doctor
	if result.LeftEyeScore <= 2 || result.RightEyeScore <= 2 {
		go sendEmailToDoctor(result) // Send email asynchronously
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Results stored successfully"))
}

func getLatestResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:04D685362v98@tcp(127.0.0.1:3306)/risk_assessment_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var result VisionResult
	query := `SELECT LeftEyeScore, RightEyeScore, Comments FROM visionResult WHERE UserID = ? ORDER BY CreatedAt DESC LIMIT 1`
	err = db.QueryRow(query, userID).Scan(&result.LeftEyeScore, &result.RightEyeScore, &result.Comments)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No results found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	result.UserID, _ = strconv.Atoi(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Function to send email using Gmail SMTP
func sendEmailToDoctor(result VisionResult) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "newuploadedvideo@gmail.com") // Replace with your Gmail
	m.SetHeader("To", "s10247445@connect.np.edu.sg")  // Doctor's email
	m.SetHeader("Subject", "Urgent: Vision Test Report for User ID "+strconv.Itoa(result.UserID))

	// Email body
	body := fmt.Sprintf(`
		<h2>Vision Test Report</h2>
		<p><strong>User ID:</strong> %d</p>
		<p><strong>Left Eye Score:</strong> %d</p>
		<p><strong>Right Eye Score:</strong> %d</p>
		<p><strong>Comments:</strong> %s</p>
		<p>Please review the report and advise accordingly.</p>
	`, result.UserID, result.LeftEyeScore, result.RightEyeScore, result.Comments)

	m.SetBody("text/html", body)

	// Configure Gmail SMTP settings
	d := gomail.NewDialer("smtp.gmail.com", 587, "newuploadedvideo@gmail.com", "agof rvwb lreo tups")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // Needed for Gmail

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// API Endpoint for Sending Email
func handleSendReportToDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var result VisionResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Send email
	err = sendEmailToDoctor(result)
	if err != nil {
		log.Println("Failed to send email:", err)
		http.Error(w, "Failed to send report to doctor", http.StatusInternalServerError)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Report sent successfully to doctor"})
}
