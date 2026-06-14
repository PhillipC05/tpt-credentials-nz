package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/handlers"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/middleware"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/repository"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/services"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "credentials")
	serverPort := getEnv("SERVER_PORT", "8094")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("database connection established")

	credentialRepo := repository.NewCredentialRepository(db)
	qrService := services.NewQRService(credentialRepo)
	credentialService := services.NewCredentialService(credentialRepo, qrService)

	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	publicHandler := handlers.NewPublicHandler(credentialService, qrService)
	credentialHandler := handlers.NewCredentialHandler(credentialService, qrService)
	pbHandler := handlers.NewProfessionalBodiesHandler(credentialRepo)
	extHandler := handlers.NewCredentialExtHandler(credentialService, qrService)
	verifierHandler := handlers.NewVerifierHandler(credentialRepo, jwtSecret)

	authRateLimit := middleware.RateLimit(10, time.Minute)

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/public/professionals/{id}", publicHandler.GetPublicProfile).Methods("GET")
	r.HandleFunc("/api/verify/{qr_id}", publicHandler.VerifyCredential).Methods("GET")
	r.HandleFunc("/api/professional-bodies", pbHandler.List).Methods("GET")

	// Auth routes (rate-limited)
	auth := r.PathPrefix("/api/auth").Subrouter()
	auth.Use(authRateLimit)
	auth.HandleFunc("/register", authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Verifier auth routes (rate-limited)
	verifierAuth := r.PathPrefix("/api/verifiers").Subrouter()
	verifierAuth.Use(authRateLimit)
	verifierAuth.HandleFunc("/register", verifierHandler.Register).Methods("POST")
	verifierAuth.HandleFunc("/login", verifierHandler.Login).Methods("POST")

	// Protected professional routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(handlers.AuthMiddleware(jwtSecret))
	protected.HandleFunc("/credentials", credentialHandler.CreateCredential).Methods("POST")
	protected.HandleFunc("/credentials", credentialHandler.ListCredentials).Methods("GET")
	protected.HandleFunc("/credentials/qr/combined", extHandler.GenerateCombinedQR).Methods("GET")
	protected.HandleFunc("/credentials/{id}", credentialHandler.GetCredential).Methods("GET")
	protected.HandleFunc("/credentials/{id}/refresh", credentialHandler.RefreshCredential).Methods("POST")
	protected.HandleFunc("/credentials/{id}/revoke", credentialHandler.RevokeCredential).Methods("POST")
	protected.HandleFunc("/credentials/{id}/qr", credentialHandler.GenerateQR).Methods("GET")
	protected.HandleFunc("/credentials/{id}/events", extHandler.GetTimeline).Methods("GET")
	protected.HandleFunc("/credentials/{id}/vc", extHandler.GetVerifiableCredential).Methods("GET")
	protected.HandleFunc("/credentials/{id}/visibility", extHandler.GetVisibility).Methods("GET")
	protected.HandleFunc("/credentials/{id}/visibility", extHandler.UpdateVisibility).Methods("PUT")

	// Protected verifier routes
	verifierProtected := r.PathPrefix("/api/verifiers").Subrouter()
	verifierProtected.Use(handlers.VerifierAuthMiddleware(jwtSecret))
	verifierProtected.HandleFunc("/webhooks", verifierHandler.RegisterWebhook).Methods("POST")
	verifierProtected.HandleFunc("/webhooks", verifierHandler.ListWebhooks).Methods("GET")

	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{corsOrigins},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	handler := c.Handler(r)

	log.Printf("credentials server starting on port %s", serverPort)
	if err := http.ListenAndServe(":"+serverPort, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
