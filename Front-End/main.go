package main

import (
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

var questions = []Question{
	{
		QuestionID:      "1",
		QuestionContent: "How often do you feel dizzy?",
		QuestionOptions: []string{"Never", "Rarely", "Sometimes", "Often"},
	},
	{
		QuestionID:      "2",
		QuestionContent: "Do you have any mobility issues?",
		QuestionOptions: []string{"No", "Minor", "Moderate", "Severe"},
	},
	// Add more questions as needed
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
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

	fmt.Println("User Responses:", responses)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Responses submitted successfully!"))
}
