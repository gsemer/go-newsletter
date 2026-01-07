package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	userapp "newsletter/internal/users/application"
	userrepo "newsletter/internal/users/infrastructure/postgres"
	transporthttp "newsletter/transport/http"
	"newsletter/transport/http/handler"
)

var counts int64

func main() {
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	userRepo := userrepo.NewUserRepository(conn)
	userService := userapp.NewUserService(userRepo)
	authService := userapp.NewAuthenticationService(userRepo)

	userHandler := handler.NewUserHandler(userService, authService)

	app := transporthttp.App{Userhandler: *userHandler}

	server := &http.Server{
		Addr:    ":8001",
		Handler: app.Routes(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down...")
	server.Shutdown(ctx)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}
		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds ...")
		time.Sleep(2 * time.Second)
		continue
	}
}
