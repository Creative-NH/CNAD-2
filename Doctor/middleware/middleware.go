package middleware

import (
	"net/http"
	"user_service/sessions"
)

func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		session, _ := sessions.Store.Get(r, "session")
		_, ok := session.Values["username"]

		if !ok {
			http.Redirect(w, r, "/", 302)
			return
		}

		handler.ServeHTTP(w, r)
	}
}
