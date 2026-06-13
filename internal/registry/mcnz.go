package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MCNZClient verifies registrations against the Medical Council of New Zealand
// public register at https://www.mcnz.org.nz/registration/register-of-doctors/.
//
// The MCNZ does not publish a formal REST API. This client hits their public
// register search endpoint using the same parameters their web form uses.
// If MCNZ exposes an official API in future, replace fetchRegister accordingly.
type MCNZClient struct {
	*httpClient
}

func NewMCNZClient() *MCNZClient {
	return &MCNZClient{
		httpClient: newHTTPClient("https://www.mcnz.org.nz", 10*time.Second),
	}
}

type mcnzSearchResult struct {
	Count   int `json:"count"`
	Results []struct {
		Name           string `json:"name"`
		RegistrationNo string `json:"registration_number"`
		Status         string `json:"status"`
	} `json:"results"`
}

func (c *MCNZClient) Verify(licenceNumber, fullName string) (Result, error) {
	// MCNZ register search — query by registration number.
	// The endpoint path and parameters reflect the current MCNZ website structure
	// and may need updating if they change their site.
	endpoint := fmt.Sprintf("%s/registration/register-of-doctors/", c.baseURL)
	params := url.Values{}
	params.Set("registration_number", licenceNumber)
	params.Set("format", "json")

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Result{}, fmt.Errorf("mcnz: failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TPT-NZ-Credentials/1.0 (+https://tpt.nz)")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("mcnz: registry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("mcnz: registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return Result{}, fmt.Errorf("mcnz: failed to read response: %w", err)
	}

	var result mcnzSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		// Fall back to basic name-match in HTML response for scraped endpoint
		return c.verifyFromHTML(string(body), licenceNumber, fullName)
	}

	for _, r := range result.Results {
		if r.RegistrationNo == licenceNumber && r.Status == "Current" {
			if fullName != "" && !nameMatch(r.Name, fullName) {
				return Result{Verified: false, Notes: "name mismatch"}, nil
			}
			return Result{Verified: true, Notes: "active registration found"}, nil
		}
	}
	return Result{Verified: false, Notes: "no matching active registration"}, nil
}

// verifyFromHTML is a fallback for when the endpoint returns HTML rather than JSON.
// It checks for the presence of the licence number and confirms the name appears.
func (c *MCNZClient) verifyFromHTML(body, licenceNumber, fullName string) (Result, error) {
	if !strings.Contains(body, licenceNumber) {
		return Result{Verified: false, Notes: "licence number not found in register"}, nil
	}
	if fullName != "" && !strings.Contains(strings.ToLower(body), strings.ToLower(fullName)) {
		return Result{Verified: false, Notes: "name not found alongside licence"}, nil
	}
	return Result{Verified: true, Notes: "found via register search"}, nil
}

// nameMatch does a loose comparison between two name strings.
func nameMatch(a, b string) bool {
	return strings.Contains(strings.ToLower(a), strings.ToLower(b)) ||
		strings.Contains(strings.ToLower(b), strings.ToLower(a))
}
