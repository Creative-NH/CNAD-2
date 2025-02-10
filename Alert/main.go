package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

	doctorServiceDB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/doctor_service")
	if err != nil {
		log.Fatal(err)
	}
	defer doctorServiceDB.Close()

	notificationsDB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/notifications_db")
	if err != nil {
		log.Fatal(err)
	}
	defer notificationsDB.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../Front-End/static/notification.html")
	})

	http.HandleFunc("/doc_notifications", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../Front-End/static/doc_notification.html")
	})

	http.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		displayNotifications(w, r, userDB, selfAssessmentDB, riskAssessmentDB, doctorServiceDB, notificationsDB)
	})

	http.HandleFunc("/login", loginHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
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
	RiskLevel           string
}

type RiskAssessment struct {
	RiskLevel      string
	Recommendation string
}

type Notification struct {
	Name      string `json:"name"`
	RiskLevel string `json:"risk_level"`
	Advice    string `json:"advice"`
	Timestamp string `json:"timestamp"`
	Doctor    string `json:"doctor"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	// For demonstration purposes, we set the UserID to 1
	session.Values["UserID"] = 1
	session.Save(r, w)
	fmt.Fprintln(w, "Logged in")
}

func displayNotifications(w http.ResponseWriter, r *http.Request, userDB *sql.DB, selfAssessmentDB *sql.DB, riskAssessmentDB *sql.DB, doctorServiceDB *sql.DB, notificationsDB *sql.DB) {
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["UserID"].(int)
	if !ok {
		userID = 1 // Default to UserID 1 if not logged in
	}

	// Query user information
	user := User{}
	userRow := userDB.QueryRow("SELECT Name FROM Users WHERE UserID = ?", userID)
	if err := userRow.Scan(&user.Name); err != nil {
		log.Printf("Error fetching user data: %v", err)
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Query assessment information
	assessment := Assessment{}
	assessmentRow := selfAssessmentDB.QueryRow("SELECT RiskLevel FROM Assessments WHERE UserID = ?", userID)
	if err := assessmentRow.Scan(&assessment.RiskLevel); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No assessment data found for UserID = %d", userID)
			http.Error(w, "No assessment data found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching assessment data: %v", err)
			http.Error(w, "Error fetching assessment data", http.StatusInternalServerError)
		}
		return
	}

	// Query corresponding recommendation
	notification := Notification{Name: user.Name, RiskLevel: assessment.RiskLevel, Timestamp: time.Now().Format(time.RFC3339)}
	notificationRow := riskAssessmentDB.QueryRow("SELECT Advice FROM Recommendations WHERE RiskLevel = ?", assessment.RiskLevel)
	if err := notificationRow.Scan(&notification.Advice); err != nil {
		log.Printf("Error fetching recommendation data: %v", err)
		http.Error(w, "Error fetching recommendation data", http.StatusInternalServerError)
		return
	}

	// Query doctor information
	doctor := ""
	doctorRow := doctorServiceDB.QueryRow("SELECT name FROM users WHERE id = 1") // Assuming you want to get the doctor with id 1
	if err := doctorRow.Scan(&doctor); err != nil {
		log.Printf("Error fetching doctor data: %v", err)
		http.Error(w, "Error fetching doctor data", http.StatusInternalServerError)
		return
	}
	notification.Doctor = doctor

	// Insert new notification into Notifications table
	message := fmt.Sprintf("Risk Level: %s. Advice: %s", notification.RiskLevel, notification.Advice)
	log.Printf("Inserting notification: UserID=%d, Message=%s", userID, message) // Log the message being inserted
	_, err := notificationsDB.Exec("INSERT INTO Notifications (UserID, Message) VALUES (?, ?)", userID, message)
	if err != nil {
		log.Printf("Error inserting notification: %v", err)
		http.Error(w, "Error inserting notification", http.StatusInternalServerError)
		return
	}

	// Serve the notification as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notification); err != nil {
		log.Printf("Error encoding notification data: %v", err)
		http.Error(w, "Error encoding notification data", http.StatusInternalServerError)
	}
}
