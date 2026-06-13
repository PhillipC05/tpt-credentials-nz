package registry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// EngineeringNZClient verifies Chartered Professional Engineer (CPEng) status
// against the Engineering New Zealand member directory.
type EngineeringNZClient struct {
	*httpClient
}

func NewEngineeringNZClient() *EngineeringNZClient {
	return &EngineeringNZClient{
		httpClient: newHTTPClient("https://www.engineeringnz.org", 10*time.Second),
	}
}

func (c *EngineeringNZClient) Verify(licenceNumber, fullName string) (Result, error) {
	endpoint := fmt.Sprintf("%s/find-an-engineer/", c.baseURL)
	params := url.Values{}
	if fullName != "" {
		params.Set("name", fullName)
	}
	params.Set("cpeng_number", licenceNumber)

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Result{}, fmt.Errorf("engineeringnz: failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "TPT-NZ-Credentials/1.0 (+https://tpt.nz)")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("engineeringnz: registry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("engineeringnz: registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return Result{}, fmt.Errorf("engineeringnz: failed to read response: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, licenceNumber) {
		return Result{Verified: false, Notes: "CPEng number not found"}, nil
	}
	return Result{Verified: true, Notes: "found in Engineering NZ directory"}, nil
}
