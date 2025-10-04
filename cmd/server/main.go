package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	cfgpkg "kpopapi/config"
	"kpopapi/internal/auth"
	"kpopapi/internal/handlers"
	"kpopapi/internal/middleware"
)

func main() {
	// Load config from .env and config.yml
	appConfig, err := cfgpkg.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Setup DB
	dsn := "host=localhost port=5432 user=postgres password=christmiraclepostgres dbname=restapi_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	// Connection pooling best practices
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Health check
	if err := db.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	// Run simple migrations (idempotent)
	if err := cfgpkg.RunMigrations(db); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// Setup services/handlers
	authSvc := auth.NewAuthService(db, appConfig)
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/api/login", authSvc.HandleLogin)
	mux.HandleFunc("/api/logout", authSvc.HandleLogout)

	// Protected endpoints
	mux.HandleFunc("/api/data", handlers.HandleSecretData)
	mux.HandleFunc("/api/idols", handlers.HandleIdols(db))
	mux.HandleFunc("/api/idols/", handlers.HandleIdolByID(db))
	

	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("db not ok"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Swagger (served via CDN with embedded spec)
	mux.HandleFunc("/swagger", handlers.SwaggerUI)
	mux.HandleFunc("/swagger.json", handlers.SwaggerSpec)

	// === Serve frontend ===
	fs := http.FileServer(http.Dir("frontend"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	
	// kalau akses root ("/") langsung arahkan ke index.html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/index.html")
	})

	// kalau akses /login, arahkan ke login.html
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/login.html")
	})

	// Compose middlewares: CORS -> Auth -> mux
	handler := middleware.CORS(auth.JWTMiddleware(authSvc, mux))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
