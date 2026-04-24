package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type APIError struct {
	Kind     string
	Status   int
	Message  string
	ExitCode int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s (status %d)", e.Kind, e.Message, e.Status)
}

type Client struct {
	token   string
	host    string
	httpCli *http.Client
}

func NewClient() (*Client, error) {
	token := resolveToken()
	return &Client{
		token: token,
		host:  "https://api.github.com",
		httpCli: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func resolveToken() string {
	for _, env := range []string{"LLMGH_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func (c *Client) Get(path string) (map[string]any, error) {
	url := c.host + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request build: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, &APIError{Kind: "network", Message: err.Error(), ExitCode: 5}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{Kind: "network", Message: "read body: " + err.Error(), ExitCode: 5}
	}

	switch {
	case resp.StatusCode == 401:
		return nil, &APIError{Kind: "auth", Status: 401, Message: "unauthorized", ExitCode: 2}
	case resp.StatusCode == 403:
		return nil, &APIError{Kind: "rate_limit", Status: 403, Message: "forbidden/rate limited", ExitCode: 4}
	case resp.StatusCode == 404:
		return nil, &APIError{Kind: "not_found", Status: 404, Message: "not found", ExitCode: 3}
	case resp.StatusCode == 422:
		return nil, &APIError{Kind: "api", Status: 422, Message: string(body), ExitCode: 6}
	case resp.StatusCode == 429:
		return nil, &APIError{Kind: "rate_limit", Status: 429, Message: "rate limited", ExitCode: 4}
	case resp.StatusCode >= 400:
		return nil, &APIError{Kind: "api", Status: resp.StatusCode, Message: string(body), ExitCode: 6}
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("json parse: %w", err)
	}
	return result, nil
}

func (c *Client) GetList(path string) ([]map[string]any, error) {
	url := c.host + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request build: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, &APIError{Kind: "network", Message: err.Error(), ExitCode: 5}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{Kind: "network", Message: "read body: " + err.Error(), ExitCode: 5}
	}

	switch {
	case resp.StatusCode == 401:
		return nil, &APIError{Kind: "auth", Status: 401, Message: "unauthorized", ExitCode: 2}
	case resp.StatusCode == 403:
		return nil, &APIError{Kind: "rate_limit", Status: 403, Message: "forbidden/rate limited", ExitCode: 4}
	case resp.StatusCode == 404:
		return nil, &APIError{Kind: "not_found", Status: 404, Message: "not found", ExitCode: 3}
	case resp.StatusCode >= 400:
		return nil, &APIError{Kind: "api", Status: resp.StatusCode, Message: string(body), ExitCode: 6}
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("json parse: %w", err)
	}
	return result, nil
}

func (c *Client) GetRateLimit() (map[string]any, error) {
	return c.Get("/rate_limit")
}

func (c *Client) AuthUser() string {
	data, err := c.Get("/user")
	if err != nil {
		return ""
	}
	if login, ok := data["login"].(string); ok {
		return login
	}
	return ""
}
