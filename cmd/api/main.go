package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	transporthttp "newsletter/transport/http"
)

func main() {
	app := transporthttp.NewApp()

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
