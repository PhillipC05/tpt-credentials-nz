package registry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NursingCouncilClient verifies registrations against the Nursing Council of
// New Zealand public register at https://www.nursingcouncil.org.nz/.
type NursingCouncilClient struct {
	*httpClient
}

func NewNursingCouncilClient() *NursingCouncilClient {
	return &NursingCouncilClient{
		httpClient: newHTTPClient("https://www.nursingcouncil.org.nz", 10*time.Second),
	}
}

func (c *NursingCouncilClient) Verify(licenceNumber, fullName string) (Result, error) {
	endpoint := fmt.Sprintf("%s/nursing/register/", c.baseURL)
	params := url.Values{}
	params.Set("registration_number", licenceNumber)

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Result{}, fmt.Errorf("ncnz: failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "TPT-NZ-Credentials/1.0 (+https://tpt.nz)")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("ncnz: registry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("ncnz: registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return Result{}, fmt.Errorf("ncnz: failed to read response: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, licenceNumber) {
		return Result{Verified: false, Notes: "registration number not found"}, nil
	}
	if fullName != "" && !strings.Contains(strings.ToLower(bodyStr), strings.ToLower(fullName)) {
		return Result{Verified: false, Notes: "name mismatch"}, nil
	}
	return Result{Verified: true, Notes: "found in Nursing Council register"}, nil
}
