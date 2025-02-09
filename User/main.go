package main

import (
	"net/http"
	"user_service/models"
	"user_service/routes"
	"user_service/utils"
)

func main() {

	models.InitDb()
	utils.LoadTemplates("templates/*.html")
	r := routes.NewRouter()
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
