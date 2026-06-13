// Package registry provides HTTP clients for verifying professional licences
// against New Zealand professional body registries.
package registry

import (
	"fmt"
	"net/http"
	"time"
)

// Result holds the outcome of a registry verification.
type Result struct {
	Verified bool
	Notes    string
}

// Client can verify a licence number against a professional body's registry.
type Client interface {
	Verify(licenceNumber, fullName string) (Result, error)
}

// httpClient is a shared base for all registry clients.
type httpClient struct {
	http    *http.Client
	baseURL string
}

func newHTTPClient(baseURL string, timeout time.Duration) *httpClient {
	return &httpClient{
		http:    &http.Client{Timeout: timeout},
		baseURL: baseURL,
	}
}

// BySlug returns the registry client for a professional body slug.
// Returns a StubClient if no real client is registered for the slug (allows
// graceful degradation while each body's API integration is built out).
func BySlug(slug string) Client {
	switch slug {
	case "nz-medical-council":
		return NewMCNZClient()
	case "nz-law-society":
		return NewNZLSClient()
	case "nz-nursing-council":
		return NewNursingCouncilClient()
	case "nz-teaching-council":
		return NewTeachingCouncilClient()
	case "engineering-nz":
		return NewEngineeringNZClient()
	case "nz-psychologists-board":
		return NewPsychologistsBoardClient()
	default:
		return &StubClient{slug: slug}
	}
}

// StubClient is used for professional bodies without a real API integration yet.
// It logs a warning and returns unverified so the credential stays pending.
type StubClient struct {
	slug string
}

func (s *StubClient) Verify(licenceNumber, _ string) (Result, error) {
	return Result{}, fmt.Errorf("no registry integration available for %q — credential stays pending until manual verification", s.slug)
}
