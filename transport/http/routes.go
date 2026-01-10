package http

import (
	"context"
	"log"
	"net/http"
	"newsletter/transport/http/handler"

	"github.com/gorilla/mux"

	awsrepo "newsletter/internal/infrastructure/aws"
	"newsletter/internal/infrastructure/database"
	"newsletter/internal/infrastructure/firebase"
	"newsletter/internal/infrastructure/workerpool"
	newsletterapp "newsletter/internal/newsletters/application"
	newsletterrepo "newsletter/internal/newsletters/infrastructure/postgres"
	serviceapp "newsletter/internal/notifications/application"
	subscribeapp "newsletter/internal/subscriptions/application"
	subscriberepo "newsletter/internal/subscriptions/infrastructure/firebase"
	userapp "newsletter/internal/users/application"
	userrepo "newsletter/internal/users/infrastructure/postgres"
)

type App struct {
	uh handler.UserHandler
	nh handler.NewsletterHandler
	sh handler.SubscriptionHandler
}

// NewApp initializes and returns a new instance of the App.
//
// It performs the following steps:
// 1. Connects to the Postgres database with retry logic. Panics if the connection fails.
// 2. Initializes a Firebase Firestore client. Panics if initialization fails.
// 3. Creates repositories for users, newsletters, and subscriptions.
// 4. Creates application services for user management, authentication, newsletters, and subscriptions.
// 5. Creates HTTP handlers for users, newsletters, and subscriptions.
// 6. Returns a pointer to an App struct containing the initialized handlers.
//
// This function is typically called once at application startup to prepare the app for handling HTTP requests.
func NewApp(wp *workerpool.WorkerPool) *App {
	dbConnection := database.InitPostgres()
	if dbConnection == nil {
		log.Fatalf("Can't connect to Postgres!")
	}

	firebaseClient, err := firebase.InitFirestore(context.TODO())
	if err != nil {
		log.Fatalf("Can't connect to Firebase! Error: %v", err)
	}

	sesClient, err := awsrepo.InitSESClient()
	if err != nil {
		log.Fatalf("Can't initialize SES client! Error: %v", err)
	}

	// Initialize repositories
	userRepo := userrepo.NewUserRepository(dbConnection)
	newsletterRepo := newsletterrepo.NewNewsletterRepository(dbConnection)
	subscriptionRepo := subscriberepo.NewSubscriptionRepository(firebaseClient)

	// Initialize services
	userService := userapp.NewUserService(userRepo)
	authService := userapp.NewAuthenticationService(userRepo)
	newsletterService := newsletterapp.NewNewsletterService(newsletterRepo)
	subscriptionService := subscribeapp.NewSubscriptionService(subscriptionRepo)
	emailService := serviceapp.NewEmailService(sesClient)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, authService)
	newsletterHandler := handler.NewNewsletterHandler(newsletterService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, emailService, wp)

	return &App{
		uh: *userHandler,
		nh: *newsletterHandler,
		sh: *subscriptionHandler,
	}
}

// Routes sets up all the HTTP routes for the application and returns an http.Handler.
//
// It uses Gorilla Mux to create subrouters for different resource types:
func (app *App) Routes() http.Handler {
	r := mux.NewRouter()

	// User routes
	userRoutes := r.PathPrefix("/users").Subrouter()
	// POST /users/signup - Handles user registration
	userRoutes.HandleFunc("/signup", app.uh.SignUp).Methods("POST")
	// POST /users/signin - Handles user login
	userRoutes.HandleFunc("/signin", app.uh.Signin).Methods("POST")

	// Newsletter routes
	newsletterRoutes := r.PathPrefix("/newsletters").Subrouter()
	// POST /newsletters - Creates a new newsletter (requires validation)
	newsletterRoutes.Handle("", app.Validate(http.HandlerFunc(app.nh.Create))).Methods("POST")
	// GET /newsletters - Retrieves all newsletters (requires validation)
	newsletterRoutes.Handle("", app.Validate(http.HandlerFunc(app.nh.GetAll))).Methods("GET")

	// Subscription routes
	subscriptionRoutes := r.PathPrefix("/subscriptions").Subrouter()
	// POST /subscriptions/{newsletter_id} - Subscribes the current user to a newsletter.
	subscriptionRoutes.HandleFunc("/{newsletter_id}", app.sh.Subscribe).Methods("POST")
	// POST /subscriptions/{newsletter_id} - Unsubscribes the current user from a newsletter.
	subscriptionRoutes.HandleFunc("/unsubscribe", app.sh.Unsubscribe).Methods("DELETE")

	return r
}
