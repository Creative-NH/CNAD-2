package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Result model
type Result struct {
	UserID    int    `json:"user_id"`
	Position  string `json:"position"`
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"is_correct"`
}

// Store result (mock implementation)
var results []Result

// Store a user's attempt
func storeResult(w http.ResponseWriter, r *http.Request) {
	var result Result
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Add result to the in-memory store
	results = append(results, result)
	log.Printf("Result stored: %+v\n", result)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Result stored successfully!"})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/store-result", storeResult).Methods("POST")

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
