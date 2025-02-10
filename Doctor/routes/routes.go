package routes

import (
	"fmt"
	"net/http"
	"user_service/middleware"
	"user_service/models"
	"user_service/utils"

	"user_service/sessions"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {

	r := mux.NewRouter()

	r.HandleFunc("/", loginGetHandler).Methods("GET")
	r.HandleFunc("/", loginPostHandler).Methods("POST")
	r.HandleFunc("/register", registerGetHandler).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")
	r.HandleFunc("/home", middleware.AuthRequired(homeGetHandler)).Methods("GET")
	r.HandleFunc("/profile", middleware.AuthRequired(profileGetHandler)).Methods("GET")
	r.HandleFunc("/alerts", middleware.AuthRequired(alertGetHandler)).Methods("GET")
	r.HandleFunc("/doctors", middleware.AuthRequired(doctorGetHandler)).Methods("GET")
	r.HandleFunc("/reports", middleware.AuthRequired(reportGetHandler)).Methods("GET")
	r.HandleFunc("/logout", logoutPostHandler).Methods("POST")

	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	return r

}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "login.html", nil)

}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	err := models.AuthenticateUser(username, password)

	if err != nil {
		switch err {
		case models.ErrUserNotFound:
			utils.ExecuteTemplate(w, "login.html", "unknown user")
		case models.ErrInvalidLogin:
			utils.ExecuteTemplate(w, "login.html", "Invalid Login")
		default:
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))

		}
		return
	}

	session, _ := sessions.Store.Get(r, "session")

	session.Values["username"] = username
	session.Save(r, w)
	http.Redirect(w, r, "/home", 302)

}

func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "register.html", nil)

}

func profileGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "profile.html", nil)

}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	fmt.Println(username, password)

	err := models.RegisterUser(username, password)

	if err != nil {
		fmt.Printf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	http.Redirect(w, r, "/", 302)

}

func homeGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "home.html", nil)

}

func alertGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "alerts.html", nil)

}
func reportGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "reports.html", nil)

}
func doctorGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "doctors.html", nil)

}

func logoutPostHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessions.Store.Get(r, "session")

	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", 302)
}
