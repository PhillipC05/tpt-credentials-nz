package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/services"
	"github.com/gorilla/mux"
)

type publicHandler struct {
	svc   *services.CredentialService
	qrSvc *services.QRService
}

// NewPublicHandler creates a new public handler for unauthenticated routes.
func NewPublicHandler(svc *services.CredentialService, qrSvc *services.QRService) *publicHandler {
	return &publicHandler{
		svc:   svc,
		qrSvc: qrSvc,
	}
}

// GetPublicProfile handles GET /api/public/professionals/{id}.
func (h *publicHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	credID := vars["id"]

	cred, err := h.svc.GetPublicCredential(credID)
	if err != nil {
		log.Printf("public profile lookup failed for %s: %v", credID, err)
		http.Error(w, `{"error":"professional not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cred)
}

// VerifyCredential handles GET /api/verify/{qr_id}.
func (h *publicHandler) VerifyCredential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qrID := vars["qr_id"]

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	cred, tokenID, credentialID, err := h.qrSvc.ResolveToken(qrID)
	if err != nil {
		log.Printf("QR verification failed for token %s: %v", qrID, err)
		h.qrSvc.LogQRScan(tokenID, credentialID, false, clientIP)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	h.qrSvc.LogQRScan(tokenID, credentialID, true, clientIP)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":          true,
		"full_name":      cred.FullName,
		"professional":   cred.Professional,
		"licence_number": cred.LicenceNumber,
		"status":         cred.Status,
		"verified_at":    cred.VerifiedAt,
		"expires_at":     cred.ExpiresAt,
	})
}
