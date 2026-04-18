package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	client *api.RESTClient
}

func NewClient() (*Client, error) {
	opts := api.ClientOptions{
		EnableCache: false,
	}
	
	client, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w (ensure 'gh auth login' is configured)", err)
	}

	return &Client{client: client}, nil
}

// ListRepoSecrets lists all secret names in a repository
func (c *Client) ListRepoSecrets(ctx context.Context, repo string) ([]string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	var response struct {
		TotalCount int `json:"total_count"`
		Secrets    []struct {
			Name string `json:"name"`
		} `json:"secrets"`
	}

	err := c.client.Get(fmt.Sprintf("repos/%s/%s/actions/secrets", owner, repoName), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list repository secrets: %w", err)
	}

	names := make([]string, len(response.Secrets))
	for i, secret := range response.Secrets {
		names[i] = secret.Name
	}

	return names, nil
}

// ListEnvironmentSecrets lists all secret names in an environment
func (c *Client) ListEnvironmentSecrets(ctx context.Context, repo, envName string) ([]string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	var response struct {
		TotalCount int `json:"total_count"`
		Secrets    []struct {
			Name string `json:"name"`
		} `json:"secrets"`
	}

	err := c.client.Get(fmt.Sprintf("repos/%s/%s/environments/%s/secrets", owner, repoName, envName), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment secrets: %w", err)
	}

	names := make([]string, len(response.Secrets))
	for i, secret := range response.Secrets {
		names[i] = secret.Name
	}

	return names, nil
}

// ListOrgSecrets lists all secret names in an organization
func (c *Client) ListOrgSecrets(ctx context.Context, org string) ([]string, error) {
	var response struct {
		TotalCount int `json:"total_count"`
		Secrets    []struct {
			Name string `json:"name"`
		} `json:"secrets"`
	}

	err := c.client.Get(fmt.Sprintf("orgs/%s/actions/secrets", org), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization secrets: %w", err)
	}

	names := make([]string, len(response.Secrets))
	for i, secret := range response.Secrets {
		names[i] = secret.Name
	}

	return names, nil
}

// CreateWorkflowFile creates a workflow file in the repository
func (c *Client) CreateWorkflowFile(ctx context.Context, repo, path, content, message string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	// Get default branch first to commit to correct branch
	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	
	if err := c.client.Get(fmt.Sprintf("repos/%s/%s", owner, repoName), &repoInfo); err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	payload := map[string]interface{}{
		"message": message,
		"content": content, // base64 encoded
		"branch":  repoInfo.DefaultBranch,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var result map[string]interface{}
	err = c.client.Put(fmt.Sprintf("repos/%s/%s/contents/%s", owner, repoName, path), bytes.NewReader(payloadBytes), &result)
	if err != nil {
		return fmt.Errorf("failed to create workflow file: %w", err)
	}

	return nil
}

// DeleteWorkflowFile deletes a workflow file from the repository
func (c *Client) DeleteWorkflowFile(ctx context.Context, repo, path, message string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	// Get file SHA first (required for deletion)
	var fileInfo struct {
		SHA string `json:"sha"`
	}
	
	if err := c.client.Get(fmt.Sprintf("repos/%s/%s/contents/%s", owner, repoName, path), &fileInfo); err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	payload := map[string]interface{}{
		"message": message,
		"sha":     fileInfo.SHA,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var result map[string]interface{}
	err = c.client.Do("DELETE", fmt.Sprintf("repos/%s/%s/contents/%s", owner, repoName, path), bytes.NewReader(payloadBytes), &result)
	if err != nil {
		return fmt.Errorf("failed to delete workflow file: %w", err)
	}

	return nil
}

// GetDefaultBranch returns the default branch of a repository
func (c *Client) GetDefaultBranch(ctx context.Context, repo string) (string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	
	if err := c.client.Get(fmt.Sprintf("repos/%s/%s", owner, repoName), &repoInfo); err != nil {
		return "", fmt.Errorf("failed to get repository info: %w", err)
	}

	return repoInfo.DefaultBranch, nil
}
