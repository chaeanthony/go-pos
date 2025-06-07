package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chaeanthony/go-pos/api"
	"github.com/chaeanthony/go-pos/internal/database"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	domain := os.Getenv("DOMAIN") // domain hosting our server
	frontend_origin := os.Getenv("FRONTEND_ORIGIN")

	pathToDB := os.Getenv("DATABASE_URL")
	if pathToDB == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	db, err := database.NewClient(pathToDB)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Ensure logs folder exists
	err = os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create logs folder: %v", err)
	}

	// Open or create the log file (append mode). Use date for name
	f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", time.Now().Format("01-02-2006")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer f.Close()

	logger := log.NewWithOptions(f, log.Options{
		ReportTimestamp: true,
		Formatter:       log.TextFormatter,
	})

	hub := api.NewHub()

	cfg := api.APIConfig{
		DB:             db,
		Port:           port,
		JWTSecret:      jwtSecret,
		Logger:         logger,
		Hub:            hub,
		CookieSecure:   strings.HasPrefix(frontend_origin, "https"),
		CookieSameSite: http.SameSiteNoneMode,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)

	// mux.HandleFunc("POST /api/signup", cfg.HandlerUsersCreate)
	mux.HandleFunc("POST /api/login", cfg.HandlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.HandlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.HandlerRevoke)
	mux.HandleFunc("GET /api/session", cfg.HandlerSession)

	mux.HandleFunc("GET /api/items", cfg.HandlerItemsGet)
	mux.HandleFunc("GET /api/items/{itemID}", cfg.HandlerItemGetByID)
	mux.Handle("POST /api/items", cfg.StoreAuthMiddleware(http.HandlerFunc(cfg.HandlerItemsCreate)))
	mux.Handle("PUT /api/items", cfg.StoreAuthMiddleware(http.HandlerFunc(cfg.HandlerItemsUpdate)))
	mux.Handle("DELETE /api/items/{itemID}", cfg.StoreAuthMiddleware(http.HandlerFunc(cfg.HandlerItemsDelete)))

	mux.HandleFunc("GET /api/orders", cfg.HandlerOrdersGet)
	mux.HandleFunc("POST /api/orders", cfg.HandlerOrdersCreate)
	mux.Handle("PUT /api/orders", http.HandlerFunc(cfg.HandlerOrdersUpdate))

	mux.Handle("/ws", http.HandlerFunc(cfg.WsHandler))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: enableCORS(mux, frontend_origin),
	}
	logger.Infof("Serving on: %s:%s/", domain, port)
	log.Printf("Serving on: %s:%s/. Logging to: %s", domain, port, f.Name())
	log.Fatal(srv.ListenAndServe())
}

func enableCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
