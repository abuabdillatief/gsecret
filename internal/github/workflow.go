package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
)

type WorkflowRun struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// TriggerWorkflow triggers a workflow_dispatch event
func (c *Client) TriggerWorkflow(ctx context.Context, repo, workflowFile string, inputs map[string]interface{}) (int64, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	// Use dedicated branch for gsecret workflows
	branch := "gsecret-retrieval"

	payload := map[string]interface{}{
		"ref":    branch,
		"inputs": inputs,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var result map[string]interface{}
	err = c.client.Post(fmt.Sprintf("repos/%s/%s/actions/workflows/%s/dispatches", owner, repoName, workflowFile), bytes.NewReader(payloadBytes), &result)
	if err != nil {
		return 0, fmt.Errorf("failed to trigger workflow: %w", err)
	}

	// Wait a bit for the workflow run to appear in the API
	time.Sleep(2 * time.Second)

	// Get the latest workflow run ID
	runID, err := c.GetLatestWorkflowRun(ctx, repo, workflowFile)
	if err != nil {
		return 0, fmt.Errorf("failed to get workflow run ID: %w", err)
	}

	return runID, nil
}

// GetLatestWorkflowRun gets the latest workflow run ID for a workflow file
func (c *Client) GetLatestWorkflowRun(ctx context.Context, repo, workflowFile string) (int64, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	var response struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}

	err := c.client.Get(fmt.Sprintf("repos/%s/%s/actions/workflows/%s/runs?per_page=1", owner, repoName, workflowFile), &response)
	if err != nil {
		return 0, fmt.Errorf("failed to get workflow runs: %w", err)
	}

	if len(response.WorkflowRuns) == 0 {
		return 0, fmt.Errorf("no workflow runs found")
	}

	return response.WorkflowRuns[0].ID, nil
}

// GetWorkflowRunStatus gets the status of a workflow run
func (c *Client) GetWorkflowRunStatus(ctx context.Context, repo string, runID int64) (*WorkflowRun, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	var run WorkflowRun
	err := c.client.Get(fmt.Sprintf("repos/%s/%s/actions/runs/%d", owner, repoName, runID), &run)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run status: %w", err)
	}

	return &run, nil
}

// WaitForWorkflowCompletion waits for a workflow run to complete
func (c *Client) WaitForWorkflowCompletion(ctx context.Context, repo string, runID int64, timeout time.Duration) (*WorkflowRun, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for workflow completion")
		}

		run, err := c.GetWorkflowRunStatus(ctx, repo, runID)
		if err != nil {
			return nil, err
		}

		if run.Status == "completed" {
			return run, nil
		}

		// Wait before polling again
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue polling
		}
	}
}

// DownloadArtifact downloads a workflow artifact
func (c *Client) DownloadArtifact(ctx context.Context, repo string, runID int64, artifactName string) ([]byte, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	// List artifacts for the run
	var response struct {
		Artifacts []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"artifacts"`
	}

	err := c.client.Get(fmt.Sprintf("repos/%s/%s/actions/runs/%d/artifacts", owner, repoName, runID), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	var artifactID int64
	for _, artifact := range response.Artifacts {
		if artifact.Name == artifactName {
			artifactID = artifact.ID
			break
		}
	}

	if artifactID == 0 {
		return nil, fmt.Errorf("artifact %s not found", artifactName)
	}

	// Download the artifact - this endpoint returns a redirect, we need to follow it
	// Using RequestWithContext through the RESTClient
	httpClient, err := api.DefaultHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	
	// Get the artifact download URL (it's a redirect)
	downloadURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/artifacts/%d/zip", owner, repoName, artifactID)
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add auth token
	token, _ := auth.TokenForHost("github.com")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download artifact: status %d, body: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// DeleteWorkflowRun deletes a workflow run
func (c *Client) DeleteWorkflowRun(ctx context.Context, repo string, runID int64) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	owner, repoName := parts[0], parts[1]

	var result map[string]interface{}
	err := c.client.Delete(fmt.Sprintf("repos/%s/%s/actions/runs/%d", owner, repoName, runID), &result)
	if err != nil {
		return fmt.Errorf("failed to delete workflow run: %w", err)
	}

	return nil
}

// CreateWorkflowFileFromTemplate creates a workflow file from the embedded template
func (c *Client) CreateWorkflowFileFromTemplate(ctx context.Context, repo, workflowPath string) error {
	// Read the workflow template
	templateContent, err := os.ReadFile("templates/workflow.yml")
	if err != nil {
		return fmt.Errorf("failed to read workflow template: %w", err)
	}

	// Base64 encode the content
	encodedContent := base64.StdEncoding.EncodeToString(templateContent)

	// Create the workflow file
	return c.CreateWorkflowFile(ctx, repo, workflowPath, encodedContent, "[gsecret] Add temporary secret retrieval workflow")
}
