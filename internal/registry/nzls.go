package registry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NZLSClient verifies practising certificates against the New Zealand Law Society
// public register at https://www.lawsociety.org.nz/find-a-lawyer/.
type NZLSClient struct {
	*httpClient
}

func NewNZLSClient() *NZLSClient {
	return &NZLSClient{
		httpClient: newHTTPClient("https://www.lawsociety.org.nz", 10*time.Second),
	}
}

func (c *NZLSClient) Verify(licenceNumber, fullName string) (Result, error) {
	// NZLS find-a-lawyer search. The Law Society uses a name-based search;
	// licence number is the Bar admission reference. We search by name and
	// confirm the certificate number appears in the result.
	endpoint := fmt.Sprintf("%s/find-a-lawyer/", c.baseURL)
	params := url.Values{}
	if fullName != "" {
		params.Set("name", fullName)
	} else {
		params.Set("ref", licenceNumber)
	}

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Result{}, fmt.Errorf("nzls: failed to build request: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "TPT-NZ-Credentials/1.0 (+https://tpt.nz)")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("nzls: registry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("nzls: registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return Result{}, fmt.Errorf("nzls: failed to read response: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, licenceNumber) {
		return Result{Verified: false, Notes: "certificate number not found in register"}, nil
	}
	if fullName != "" && !strings.Contains(strings.ToLower(bodyStr), strings.ToLower(fullName)) {
		return Result{Verified: false, Notes: "name not found alongside certificate"}, nil
	}
	return Result{Verified: true, Notes: "found in NZLS register"}, nil
}
