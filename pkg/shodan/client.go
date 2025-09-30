package shodan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Shodan API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Shodan API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.shodan.io",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetHost queries Shodan for information about a specific IP address
func (c *Client) GetHost(ip string) (*Host, error) {
	url := fmt.Sprintf("%s/shodan/host/%s?key=%s", c.baseURL, ip, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query Shodan API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Shodan API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var host Host
	if err := json.Unmarshal(body, &host); err != nil {
		return nil, fmt.Errorf("failed to parse Shodan response: %w", err)
	}

	return &host, nil
}

// GetHostMinimal queries Shodan for basic information about a specific IP address
// This is a lighter version that returns less data and consumes fewer API credits
func (c *Client) GetHostMinimal(ip string) (*Host, error) {
	url := fmt.Sprintf("%s/shodan/host/%s?key=%s&minify=true", c.baseURL, ip, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query Shodan API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Shodan API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var host Host
	if err := json.Unmarshal(body, &host); err != nil {
		return nil, fmt.Errorf("failed to parse Shodan response: %w", err)
	}

	return &host, nil
}

// IsValidAPIKey checks if the provided API key is valid
func (c *Client) IsValidAPIKey() error {
	url := fmt.Sprintf("%s/api-info?key=%s", c.baseURL, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid Shodan API key")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API key validation failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
