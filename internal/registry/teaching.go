package registry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TeachingCouncilClient verifies registrations against the Teaching Council of
// Aotearoa New Zealand public register.
type TeachingCouncilClient struct {
	*httpClient
}

func NewTeachingCouncilClient() *TeachingCouncilClient {
	return &TeachingCouncilClient{
		httpClient: newHTTPClient("https://teachingcouncil.nz", 10*time.Second),
	}
}

func (c *TeachingCouncilClient) Verify(licenceNumber, fullName string) (Result, error) {
	endpoint := fmt.Sprintf("%s/find-a-teacher/", c.baseURL)
	params := url.Values{}
	params.Set("registration_number", licenceNumber)

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Result{}, fmt.Errorf("tcnz: failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "TPT-NZ-Credentials/1.0 (+https://tpt.nz)")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("tcnz: registry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("tcnz: registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return Result{}, fmt.Errorf("tcnz: failed to read response: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, licenceNumber) {
		return Result{Verified: false, Notes: "registration number not found"}, nil
	}
	if fullName != "" && !strings.Contains(strings.ToLower(bodyStr), strings.ToLower(fullName)) {
		return Result{Verified: false, Notes: "name mismatch"}, nil
	}
	return Result{Verified: true, Notes: "found in Teaching Council register"}, nil
}
