package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tpt-nz/realme-go"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// RealMeAuthHandler provides RealMe SAML authentication routes.
type RealMeAuthHandler struct {
	provider *realme.Provider
	logger   *slog.Logger
}

// NewRealMeAuthHandler creates a new RealMe auth handler.
func NewRealMeAuthHandler(provider *realme.Provider, logger *slog.Logger) *RealMeAuthHandler {
	return &RealMeAuthHandler{
		provider: provider,
		logger:   logger.With("handler", "realme_auth"),
	}
}

// Login initiates the RealMe login flow.
func (h *RealMeAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.provider.LoginHandler()(w, r)
}

// Callback handles the RealMe SAML POST-binding callback.
func (h *RealMeAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	h.provider.CallbackHandler(nil)(w, r)
}

// Logout clears the RealMe session cookie.
func (h *RealMeAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.provider.LogoutHandler()(w, r)
}

// Metadata serves the SAML SP metadata XML for DIA registration.
func (h *RealMeAuthHandler) Metadata(w http.ResponseWriter, r *http.Request) {
	h.provider.MetadataHandler()(w, r)
}

// Status returns the current RealMe authentication state.
func (h *RealMeAuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		respondJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"identity": map[string]any{
			"flt":        identity.FLT,
			"fullName":   identity.FullName,
			"assurance":  identity.AssuranceLevel.String(),
			"isVerified": identity.IsVerified(),
		},
	})
}
