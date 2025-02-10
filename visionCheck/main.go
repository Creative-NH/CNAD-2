package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

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

	result.UserID = 5
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
