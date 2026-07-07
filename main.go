package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"ticket-system/internal/handlers"
	"ticket-system/internal/middleware"
	"ticket-system/internal/store"
)

func main() {
	_ = godotenv.Load()

	s := store.New()
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(s)
	ticketHandler := handlers.NewTicketHandler(s)

	mux := http.NewServeMux()

	mux.Handle("/health", healthHandler)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)

	ticketRoutes := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case path == "/tickets" && r.Method == http.MethodPost:
			ticketHandler.Create(w, r)
		case path == "/tickets" && r.Method == http.MethodGet:
			ticketHandler.List(w, r)
		case strings.HasPrefix(path, "/tickets/") && strings.HasSuffix(path, "/status"):
			ticketHandler.UpdateStatus(w, r)
		case strings.HasPrefix(path, "/tickets/") && r.Method == http.MethodGet:
			ticketHandler.GetByID(w, r)
		default:
			handlers.WriteMethodNotAllowed(w)
		}
	})

	mux.Handle("/tickets", middleware.Auth(ticketRoutes))
	mux.Handle("/tickets/", middleware.Auth(ticketRoutes))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
