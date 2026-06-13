package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// --- JWT Auth Handler ---

type authHandler struct {
	db        *sql.DB
	jwtSecret string
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
}

type userRow struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(db *sql.DB, jwtSecret string) *authHandler {
	return &authHandler{db: db, jwtSecret: jwtSecret}
}

// Register handles user registration.
func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, `{"error":"email, password, and name are required"}`, http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = h.db.Exec(
		`INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, req.Email, string(hash), req.Name, now, now,
	)
	if err != nil {
		log.Printf("failed to create user: %v", err)
		http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		return
	}

	token, err := h.generateJWT(id, req.Email)
	if err != nil {
		log.Printf("failed to generate JWT: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	resp := authResponse{}
	resp.Token = token
	resp.User.ID = id
	resp.User.Email = req.Email
	resp.User.Name = req.Name

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login handles user authentication.
func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	var user userRow
	err := h.db.QueryRow(
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.generateJWT(user.ID, user.Email)
	if err != nil {
		log.Printf("failed to generate JWT: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	resp := authResponse{}
	resp.Token = token
	resp.User.ID = user.ID
	resp.User.Email = user.Email
	resp.User.Name = user.Name

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *authHandler) generateJWT(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// AuthMiddleware is a middleware that validates JWT tokens on protected routes.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr = authHeader[7:]
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, `{"error":"invalid user in token"}`, http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-ID", userID)
			next.ServeHTTP(w, r)
		})
	}
}

// --- Credential Handler ---

type credentialHandler struct {
	svc   *services.CredentialService
	qrSvc *services.QRService
}

// NewCredentialHandler creates a new credential handler for protected routes.
func NewCredentialHandler(svc *services.CredentialService, qrSvc *services.QRService) *credentialHandler {
	return &credentialHandler{
		svc:   svc,
		qrSvc: qrSvc,
	}
}

// CreateCredential handles POST /api/credentials.
func (h *credentialHandler) CreateCredential(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req models.CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ProfessionalBodySlug == "" {
		http.Error(w, `{"error":"professional_body_slug is required"}`, http.StatusBadRequest)
		return
	}
	if req.LicenceNumber == "" {
		http.Error(w, `{"error":"licence_number is required"}`, http.StatusBadRequest)
		return
	}
	if req.FullName == "" {
		http.Error(w, `{"error":"full_name is required"}`, http.StatusBadRequest)
		return
	}

	cred, err := h.svc.CreateCredential(userID, &req)
	if err != nil {
		log.Printf("failed to create credential: %v", err)
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cred)
}

// ListCredentials handles GET /api/credentials.
func (h *credentialHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	creds, err := h.svc.ListCredentials(userID)
	if err != nil {
		log.Printf("failed to list credentials: %v", err)
		http.Error(w, `{"error":"failed to list credentials"}`, http.StatusInternalServerError)
		return
	}

	if creds == nil {
		creds = []models.Credential{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
}

// GetCredential handles GET /api/credentials/{id}.
func (h *credentialHandler) GetCredential(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	if cred.UserID != userID {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
}

// RefreshCredential handles POST /api/credentials/{id}/refresh.
func (h *credentialHandler) RefreshCredential(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	if cred.UserID != userID {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	updated, err := h.svc.RefreshCredentialStatus(credID)
	if err != nil {
		log.Printf("failed to refresh credential %s: %v", credID, err)
		http.Error(w, `{"error":"failed to refresh credential"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// RevokeCredential handles POST /api/credentials/{id}/revoke.
func (h *credentialHandler) RevokeCredential(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	if cred.UserID != userID {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	if err := h.svc.RevokeCredential(credID); err != nil {
		log.Printf("failed to revoke credential %s: %v", credID, err)
		http.Error(w, `{"error":"failed to revoke credential"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "revoked"})
}

// GenerateQR handles GET /api/credentials/{id}/qr.
func (h *credentialHandler) GenerateQR(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	if cred.UserID != userID {
		http.Error(w, `{"error":"credential not found"}`, http.StatusNotFound)
		return
	}

	token, qrPNG, err := h.qrSvc.GenerateQRToken(credID)
	if err != nil {
		log.Printf("failed to generate QR code: %v", err)
		http.Error(w, `{"error":"failed to generate QR code"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":          token,
		"qr_code_base64": qrPNG,
		"verify_url":     h.qrSvc.GetVerificationURL(token),
	})
}
