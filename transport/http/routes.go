package http

import (
	"log"
	"net/http"
	"newsletter/transport/http/handler"

	"github.com/gorilla/mux"

	"newsletter/internal/infrastructure/database"
	userapp "newsletter/internal/users/application"
	userrepo "newsletter/internal/users/infrastructure/postgres"
)

type App struct {
	uh handler.UserHandler
}

func NewApp() *App {
	conn := database.ConnectWithRetry()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	userRepo := userrepo.NewUserRepository(conn)
	userService := userapp.NewUserService(userRepo)
	authService := userapp.NewAuthenticationService(userRepo)

	userHandler := handler.NewUserHandler(userService, authService)

	return &App{uh: *userHandler}
}

func (app *App) Routes() http.Handler {
	r := mux.NewRouter()

	userRoutes := r.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/signup", app.uh.SignUp).Methods("POST")
	userRoutes.HandleFunc("/signin", app.uh.Signin).Methods("POST")

	return r
}
