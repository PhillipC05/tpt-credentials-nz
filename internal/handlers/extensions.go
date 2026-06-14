package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/repository"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// --- Professional Bodies ---

type professionalBodiesHandler struct {
	repo repository.Store
}

func NewProfessionalBodiesHandler(repo repository.Store) *professionalBodiesHandler {
	return &professionalBodiesHandler{repo: repo}
}

func (h *professionalBodiesHandler) List(w http.ResponseWriter, r *http.Request) {
	bodies, err := h.repo.ListProfessionalBodies()
	if err != nil {
		log.Printf("failed to list professional bodies: %v", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if bodies == nil {
		bodies = []models.ProfessionalBody{}
	}
	respondJSON(w, http.StatusOK, bodies)
}

// --- Credential Extensions (timeline, VC, multi-QR, visibility) ---

type credentialExtHandler struct {
	svc   *services.CredentialService
	qrSvc *services.QRService
}

func NewCredentialExtHandler(svc *services.CredentialService, qrSvc *services.QRService) *credentialExtHandler {
	return &credentialExtHandler{svc: svc, qrSvc: qrSvc}
}

// GetTimeline handles GET /api/credentials/{id}/events.
func (h *credentialExtHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil || cred.UserID != userID {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "credential not found"})
		return
	}

	events, err := h.svc.GetCredentialEvents(credID)
	if err != nil {
		log.Printf("failed to get events for credential %s: %v", credID, err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if events == nil {
		events = []models.CredentialEvent{}
	}
	respondJSON(w, http.StatusOK, events)
}

// GetVerifiableCredential handles GET /api/credentials/{id}/vc.
func (h *credentialExtHandler) GetVerifiableCredential(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil || cred.UserID != userID {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "credential not found"})
		return
	}

	vc, err := h.svc.GetVerifiableCredential(credID)
	if err != nil {
		log.Printf("failed to build VC for credential %s: %v", credID, err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.Header().Set("Content-Type", "application/ld+json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"credential.jsonld\"")
	json.NewEncoder(w).Encode(vc)
}

// GenerateCombinedQR handles GET /api/credentials/qr/combined.
// Produces a single QR linking to all active credentials for the authenticated user.
func (h *credentialExtHandler) GenerateCombinedQR(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	creds, err := h.svc.ListCredentials(userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	var tokens []map[string]string
	for _, cred := range creds {
		if cred.Status != models.CredentialStatusActive {
			continue
		}
		token, qrB64, err := h.qrSvc.GenerateQRToken(cred.ID)
		if err != nil {
			log.Printf("failed to generate QR for credential %s: %v", cred.ID, err)
			continue
		}
		tokens = append(tokens, map[string]string{
			"credential_id": cred.ID,
			"token":         token,
			"verify_url":    h.qrSvc.GetVerificationURL(token),
			"qr_code_base64": qrB64,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"credentials": tokens,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// GetVisibility handles GET /api/credentials/{id}/visibility.
func (h *credentialExtHandler) GetVisibility(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil || cred.UserID != userID {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "credential not found"})
		return
	}

	vis, err := h.svc.GetVisibility(credID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	respondJSON(w, http.StatusOK, vis)
}

// UpdateVisibility handles PUT /api/credentials/{id}/visibility.
func (h *credentialExtHandler) UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetCredential(credID)
	if err != nil || cred.UserID != userID {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "credential not found"})
		return
	}

	var vis models.CredentialVisibility
	if err := json.NewDecoder(r.Body).Decode(&vis); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateVisibility(credID, &vis); err != nil {
		log.Printf("failed to update visibility for %s: %v", credID, err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	updated, _ := h.svc.GetVisibility(credID)
	respondJSON(w, http.StatusOK, updated)
}

// --- Verifier Accounts ---

type verifierHandler struct {
	repo      repository.Store
	jwtSecret string
}

func NewVerifierHandler(repo repository.Store, jwtSecret string) *verifierHandler {
	return &verifierHandler{repo: repo, jwtSecret: jwtSecret}
}

type verifierRegisterRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Organisation string `json:"organisation"`
}

// Register handles POST /api/verifiers/register.
func (h *verifierHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req verifierRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" || req.Organisation == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "email, password, and organisation are required"})
		return
	}
	if !emailRegex.MatchString(req.Email) {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}
	if len(req.Password) < 8 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	account := &models.VerifierAccount{
		Email:        req.Email,
		PasswordHash: string(hash),
		Organisation: req.Organisation,
	}
	if err := h.repo.CreateVerifierAccount(account); err != nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}

	token, err := h.generateVerifierJWT(account.ID, account.Email)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"token":    token,
		"verifier": map[string]string{"id": account.ID, "email": account.Email, "organisation": account.Organisation},
	})
}

// Login handles POST /api/verifiers/login.
func (h *verifierHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	account, err := h.repo.GetVerifierAccountByEmail(req.Email)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := h.generateVerifierJWT(account.ID, account.Email)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"verifier": map[string]string{"id": account.ID, "email": account.Email, "organisation": account.Organisation},
	})
}

// RegisterWebhook handles POST /api/verifiers/webhooks.
func (h *verifierHandler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	verifierID := r.Header.Get("X-Verifier-ID")
	if verifierID == "" {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
		Events string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.URL == "" || req.Secret == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "url and secret are required"})
		return
	}
	if req.Events == "" {
		req.Events = "credential.revoked,credential.expired,credential.verified"
	}

	ep := &models.WebhookEndpoint{
		VerifierID: verifierID,
		URL:        req.URL,
		Secret:     req.Secret,
		Events:     req.Events,
		Active:     true,
	}
	if err := h.repo.CreateWebhookEndpoint(ep); err != nil {
		log.Printf("failed to create webhook endpoint: %v", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         ep.ID,
		"url":        ep.URL,
		"events":     ep.Events,
		"active":     ep.Active,
		"created_at": ep.CreatedAt,
	})
}

// ListWebhooks handles GET /api/verifiers/webhooks.
func (h *verifierHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	verifierID := r.Header.Get("X-Verifier-ID")
	if verifierID == "" {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	webhooks, err := h.repo.ListWebhookEndpointsByVerifier(verifierID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if webhooks == nil {
		webhooks = []models.WebhookEndpoint{}
	}
	respondJSON(w, http.StatusOK, webhooks)
}

func (h *verifierHandler) generateVerifierJWT(id, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  id,
		"email": email,
		"role": "verifier",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// VerifierAuthMiddleware validates JWT tokens for verifier-protected routes.
func VerifierAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
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
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				return
			}
			if claims["role"] != "verifier" {
				respondJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			verifierID, _ := claims["sub"].(string)
			r.Header.Set("X-Verifier-ID", verifierID)
			next.ServeHTTP(w, r)
		})
	}
}
