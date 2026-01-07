package http

import (
	"net/http"
	"newsletter/transport/http/handler"

	"github.com/gorilla/mux"
)

type App struct {
	Userhandler handler.UserHandler
}

func (app *App) Routes() http.Handler {
	r := mux.NewRouter()

	userRoutes := r.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/signup", app.Userhandler.SignUp).Methods("POST")
	userRoutes.HandleFunc("/signin", app.Userhandler.Signin).Methods("POST")

	return r
}
