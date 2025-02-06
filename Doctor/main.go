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

func connectDB(dbUser, dbPassword, dbHost, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func main() {
	dbUser, dbPassword, dbHost, dbName := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME")
	localPort := os.Getenv("LOCAL_PORT")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" || localPort == "" {
		log.Fatal("Missing required environment variables")
	}

	db, err := connectDB(dbUser, dbPassword, dbHost, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	router := mux.NewRouter().PathPrefix("/api/doctor_management").Subrouter()
	router.HandleFunc("/api/doctor/login", func(w http.ResponseWriter, r *http.Request) {
		doctorLoginHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/doctor/alerts", func(w http.ResponseWriter, r *http.Request) {
		viewAllAlertsHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/api/doctor/report", func(w http.ResponseWriter, r *http.Request) {
		viewReportHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/api/doctor/resolveAlert", func(w http.ResponseWriter, r *http.Request) {
		resolveAlertHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/doctor/list", func(w http.ResponseWriter, r *http.Request) {
		listAvailableDoctorsHandler(w, r, db)
	}).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	log.Printf("Starting Doctor Service on :%s", localPort)
	if err := http.ListenAndServe(":"+localPort, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func doctorLoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&credentials)

	var doctorID int
	err := db.QueryRow("SELECT id FROM Doctors WHERE username = ? AND password = ?", credentials.Username, credentials.Password).Scan(&doctorID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"doctor_id": doctorID})
}

func viewAllAlertsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query("SELECT id, user_id, alert_message, created_at, resolved FROM Alerts ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var alerts []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var alertMessage string
		var createdAt time.Time
		var resolved bool
		rows.Scan(&id, &userID, &alertMessage, &createdAt, &resolved)
		alerts = append(alerts, map[string]interface{}{"id": id, "user_id": userID, "alert_message": alertMessage, "created_at": createdAt, "resolved": resolved})
	}

	json.NewEncoder(w).Encode(alerts)
}

func viewReportHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID := r.URL.Query().Get("user_id")
	row := db.QueryRow("SELECT report FROM Reports WHERE user_id = ?", userID)
	var report string
	if err := row.Scan(&report); err != nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"report": report})
}

func resolveAlertHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	alertID := r.URL.Query().Get("alert_id")
	_, err := db.Exec("UPDATE Alerts SET resolved = TRUE WHERE id = ?", alertID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

func listAvailableDoctorsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query("SELECT id, name FROM Doctors WHERE available = TRUE")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var doctors []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		doctors = append(doctors, map[string]interface{}{"id": id, "name": name})
	}
	json.NewEncoder(w).Encode(doctors)
}
