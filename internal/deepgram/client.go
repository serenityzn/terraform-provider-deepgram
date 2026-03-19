package deepgram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.deepgram.com"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Project types

// Project represents a single Deepgram project.
type Project struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	MIPOptOut bool   `json:"mip_opt_out"`
}

// ListProjectsResponse is the response from listing projects.
type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
}

// ListProjects retrieves all projects associated with the API key.
func (c *Client) ListProjects(ctx context.Context) (*ListProjectsResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("list projects: unexpected status %d: %s", statusCode, string(body))
	}
	var resp ListProjectsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("list projects: failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

// CreateKeyRequest is the payload for creating an API key.
type CreateKeyRequest struct {
	Comment        string   `json:"comment"`
	Scopes         []string `json:"scopes"`
	Tags           []string `json:"tags,omitempty"`
	ExpirationDate string   `json:"expiration_date,omitempty"`
}

// CreateKeyResponse is the response from creating an API key.
type CreateKeyResponse struct {
	APIKeyID       string   `json:"api_key_id"`
	Key            string   `json:"key"`
	Comment        string   `json:"comment"`
	Scopes         []string `json:"scopes"`
	Tags           []string `json:"tags"`
	ExpirationDate string   `json:"expiration_date"`
}

// GetKeyResponse is the response from getting a single API key.
type GetKeyResponse struct {
	Item struct {
		Member struct {
			MemberID  string `json:"member_id"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			APIKey    struct {
				APIKeyID       string   `json:"api_key_id"`
				Comment        string   `json:"comment"`
				Scopes         []string `json:"scopes"`
				Tags           []string `json:"tags"`
				ExpirationDate string   `json:"expiration_date"`
				Created        string   `json:"created"`
			} `json:"api_key"`
		} `json:"member"`
	} `json:"item"`
}

// ListKeysResponseItem represents one key entry in the list response.
type ListKeysResponseItem struct {
	Member struct {
		MemberID string `json:"member_id"`
		Email    string `json:"email"`
	} `json:"member"`
	APIKey struct {
		APIKeyID string   `json:"api_key_id"`
		Comment  string   `json:"comment"`
		Scopes   []string `json:"scopes"`
		Created  string   `json:"created"`
	} `json:"api_key"`
}

// ListKeysResponse is the response from listing API keys.
type ListKeysResponse struct {
	APIKeys []ListKeysResponseItem `json:"api_keys"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// CreateKey creates a new API key for the given project.
func (c *Client) CreateKey(ctx context.Context, projectID string, req CreateKeyRequest) (*CreateKeyResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/projects/%s/keys", projectID), req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("create key: unexpected status %d: %s", statusCode, string(body))
	}
	var resp CreateKeyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("create key: failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

// GetKey retrieves a single API key by ID.
// Returns (nil, nil) when the key is not found (404).
func (c *Client) GetKey(ctx context.Context, projectID, keyID string) (*GetKeyResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/projects/%s/keys/%s", projectID, keyID), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get key: unexpected status %d: %s", statusCode, string(body))
	}
	var resp GetKeyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("get key: failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

// ListKeys retrieves all API keys for the given project.
func (c *Client) ListKeys(ctx context.Context, projectID string) (*ListKeysResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/projects/%s/keys", projectID), nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("list keys: unexpected status %d: %s", statusCode, string(body))
	}
	var resp ListKeysResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("list keys: failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

// DeleteKey deletes an API key for the given project.
func (c *Client) DeleteKey(ctx context.Context, projectID, keyID string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/v1/projects/%s/keys/%s", projectID, keyID), struct{}{})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("delete key: unexpected status %d: %s", statusCode, string(body))
	}
	return nil
}
