package http

import (
	"log"
	"net/http"
	"newsletter/transport/http/handler"

	"github.com/gorilla/mux"

	"newsletter/internal/infrastructure/database"
	newsletterapp "newsletter/internal/newsletters/application"
	newsletterrepo "newsletter/internal/newsletters/infrastructure/postgres"
	userapp "newsletter/internal/users/application"
	userrepo "newsletter/internal/users/infrastructure/postgres"
)

type App struct {
	uh handler.UserHandler
	nh handler.NewsletterHandler
}

func NewApp() *App {
	conn := database.ConnectWithRetry()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	userRepo := userrepo.NewUserRepository(conn)
	newsletterRepo := newsletterrepo.NewNewsletterRepository(conn)

	userService := userapp.NewUserService(userRepo)
	authService := userapp.NewAuthenticationService(userRepo)
	newsletterService := newsletterapp.NewNewsletterService(newsletterRepo)

	userHandler := handler.NewUserHandler(userService, authService)
	newsletterHandler := handler.NewNewsletterHandler(newsletterService)

	return &App{uh: *userHandler, nh: *newsletterHandler}
}

func (app *App) Routes() http.Handler {
	r := mux.NewRouter()

	userRoutes := r.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/signup", app.uh.SignUp).Methods("POST")
	userRoutes.HandleFunc("/signin", app.uh.Signin).Methods("POST")

	newsletterRoutes := r.PathPrefix("/newsletters").Subrouter()
	newsletterRoutes.Handle("", app.Validate(http.HandlerFunc(app.nh.Create))).Methods("POST")
	newsletterRoutes.Handle("", app.Validate(http.HandlerFunc(app.nh.GetAll))).Methods("GET")

	return r
}
