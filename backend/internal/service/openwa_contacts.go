package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenWAContact struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Pushname       string `json:"pushName"` // API returns camelCase pushName
	Number         string `json:"number"`
	ProfilePicUrl  string `json:"profilePicUrl"`
}

// CheckNumberExists checks if a phone number exists on WhatsApp
func (s *OpenWAService) CheckNumberExists(sessionID, phone string) (bool, error) {
	if !s.cfg.OpenWAEnabled {
		return false, nil
	}

	cleaned := CleanPhoneNumber(phone)
	url := fmt.Sprintf("%s/api/sessions/%s/contacts/check/%s", s.cfg.OpenWABaseURL, sessionID, cleaned)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return false, err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to check number existence: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Exists, nil
}

// GetContactInfo retrieves the contact information (pushname and avatar) from OpenWA.
// contactID must be in the format: number@c.us (use FormatContactID to convert a phone number).
func (s *OpenWAService) GetContactInfo(sessionID, contactID string) (*OpenWAContact, error) {
	if !s.cfg.OpenWAEnabled {
		return nil, fmt.Errorf("OpenWA is disabled")
	}

	url := fmt.Sprintf("%s/api/sessions/%s/contacts/%s", s.cfg.OpenWABaseURL, sessionID, contactID)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get contact info: %d %s", resp.StatusCode, string(body))
	}

	var contact OpenWAContact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return nil, err
	}

	// If no profilePicUrl in contact response, try the dedicated profile-picture endpoint
	if contact.ProfilePicUrl == "" {
		picURL := fmt.Sprintf("%s/api/sessions/%s/contacts/%s/profile-picture", s.cfg.OpenWABaseURL, sessionID, contactID)
		picReq, err2 := http.NewRequest("GET", picURL, http.NoBody)
		if err2 == nil {
			if s.cfg.OpenWAApiKey != "" {
				picReq.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
			}
			picResp, err2 := s.httpClient.Do(picReq)
			if err2 == nil {
				defer func() { _ = picResp.Body.Close() }()
				if picResp.StatusCode == http.StatusOK {
					var picResult struct {
						URL string `json:"url"`
					}
					if json.NewDecoder(picResp.Body).Decode(&picResult) == nil && picResult.URL != "" {
						contact.ProfilePicUrl = picResult.URL
					}
				}
			}
		}
	}

	return &contact, nil
}
