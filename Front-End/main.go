package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Question struct {
	QuestionID      string   `json:"question_id"`
	QuestionContent string   `json:"question_content"`
	QuestionOptions []string `json:"question_options"`
}

type UserResponse struct {
	QuestionID string `json:"question_id"`
	Response   int    `json:"response"`
}

func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/language", serveLanguage)
	http.HandleFunc("/quiz", serveQuiz)
	http.HandleFunc("/signup", serveSignup)
	http.HandleFunc("/api/self_assessment/questionnaire", handleQuestionnaire)
	http.HandleFunc("/api/self_assessment/submit", handleSubmit)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func serveLanguage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/language.html")
}

func serveQuiz(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/quiz.html")
}

func serveSignup(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/signup.html")
}

func handleQuestionnaire(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "English"
	}

	// Fetch questions from the external API
	apiURL := fmt.Sprintf("http://localhost/api/frontend/questionnaire?language=%s", language)
	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to fetch questions from the API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read API response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	var responses map[string]int
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &responses); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Convert responses to the format expected by the external API
	var userResponses []UserResponse
	for qID, resp := range responses {
		userResponses = append(userResponses, UserResponse{
			QuestionID: qID,
			Response:   resp,
		})
	}

	// Submit responses to the external API
	apiURL := "http://localhost/api/frontend/submit"
	jsonData, err := json.Marshal(userResponses)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to submit responses to the API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "API returned an error", resp.StatusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Responses submitted successfully!"))
}
