package retriever

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mohammadrendra/gsecret/internal/cleanup"
	"github.com/mohammadrendra/gsecret/internal/github"
)

type Retriever struct {
	client  *github.Client
	verbose bool
}

func NewRetriever(verbose bool) (*Retriever, error) {
	client, err := github.NewClient()
	if err != nil {
		return nil, err
	}

	return &Retriever{
		client:  client,
		verbose: verbose,
	}, nil
}

func (r *Retriever) log(format string, args ...interface{}) {
	if r.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// RetrieveRepoSecrets retrieves repository secrets
func (r *Retriever) RetrieveRepoSecrets(ctx context.Context, repo string, secretNames []string, all bool) (map[string]string, error) {
	// If --all flag is set, list all secrets first
	if all {
		r.log("Fetching list of all secrets...")
		names, err := r.client.ListRepoSecrets(ctx, repo)
		if err != nil {
			return nil, err
		}
		secretNames = names
		r.log("Found %d secrets to retrieve", len(secretNames))
	}

	if len(secretNames) == 0 {
		return nil, fmt.Errorf("no secrets to retrieve")
	}

	return r.retrieveSecrets(ctx, repo, secretNames, "repo", "")
}

// RetrieveEnvironmentSecrets retrieves environment secrets
func (r *Retriever) RetrieveEnvironmentSecrets(ctx context.Context, repo, envName string, secretNames []string, all bool) (map[string]string, error) {
	if all {
		r.log("Fetching list of all environment secrets...")
		names, err := r.client.ListEnvironmentSecrets(ctx, repo, envName)
		if err != nil {
			return nil, err
		}
		secretNames = names
		r.log("Found %d secrets to retrieve", len(secretNames))
	}

	if len(secretNames) == 0 {
		return nil, fmt.Errorf("no secrets to retrieve")
	}

	return r.retrieveSecrets(ctx, repo, secretNames, "env", envName)
}

// RetrieveOrgSecrets retrieves organization secrets
func (r *Retriever) RetrieveOrgSecrets(ctx context.Context, org string, secretNames []string) (map[string]string, error) {
	if len(secretNames) == 0 {
		r.log("Fetching list of all organization secrets...")
		names, err := r.client.ListOrgSecrets(ctx, org)
		if err != nil {
			return nil, err
		}
		secretNames = names
		r.log("Found %d secrets to retrieve", len(secretNames))
	}

	if len(secretNames) == 0 {
		return nil, fmt.Errorf("no secrets to retrieve")
	}

	// For org secrets, we need to use a repository that has access to them
	// This is a limitation - user needs to specify a repo with access
	return nil, fmt.Errorf("organization secret retrieval requires specifying a repository with access (use: gsecret get owner/repo SECRET --org)")
}

// retrieveSecrets is the main retrieval logic
func (r *Retriever) retrieveSecrets(ctx context.Context, repo string, secretNames []string, secretType, envName string) (map[string]string, error) {
	workflowFile := fmt.Sprintf("gsecret-temp-%d.yml", time.Now().Unix())
	workflowPath := fmt.Sprintf(".github/workflows/%s", workflowFile)
	branchName := "gsecret-retrieval"

	// Initialize cleanup manager
	cleaner := cleanup.NewCleanup(r.client, repo)
	cleaner.SetVerbose(r.verbose)

	// Ensure cleanup runs even if there's an error
	defer func() {
		if err := cleaner.Cleanup(ctx); err != nil {
			r.log("Warning: cleanup failed: %v", err)
		}
	}()

	// Step 1: Create dedicated branch for gsecret workflows
	r.log("Creating dedicated branch '%s'...", branchName)
	if err := r.client.CreateBranch(ctx, repo, branchName); err != nil {
		r.log("Note: Branch may already exist, continuing...")
	}
	cleaner.AddBranch(branchName)
	r.log("✓ Branch ready")

	// Step 2: Create temporary workflow file on the dedicated branch
	r.log("Creating temporary workflow file on '%s' branch...", branchName)
	if err := r.client.CreateWorkflowFileFromTemplate(ctx, repo, workflowPath); err != nil {
		return nil, fmt.Errorf("failed to create workflow file: %w", err)
	}
	cleaner.AddWorkflowFile(workflowPath)
	r.log("✓ Workflow file created")

	// Wait a bit for GitHub to process the new workflow
	time.Sleep(3 * time.Second)

	// Step 3: Prepare inputs
	secretsJSON, err := json.Marshal(secretNames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret names: %w", err)
	}

	inputs := map[string]interface{}{
		"secrets_json":  string(secretsJSON),
		"secret_type":   secretType,
	}

	if envName != "" {
		inputs["environment_name"] = envName
	}

	// Step 3: Trigger workflow
	r.log("Triggering workflow to export secrets...")
	runID, err := r.client.TriggerWorkflow(ctx, repo, workflowFile, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger workflow: %w", err)
	}
	cleaner.AddWorkflowRun(runID)
	r.log("✓ Workflow triggered (run ID: %d)", runID)

	// Step 4: Wait for completion
	r.log("Waiting for workflow to complete...")
	run, err := r.client.WaitForWorkflowCompletion(ctx, repo, runID, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for workflow: %w", err)
	}

	if run.Conclusion != "success" {
		return nil, fmt.Errorf("workflow failed with conclusion: %s", run.Conclusion)
	}
	r.log("✓ Workflow completed successfully")

	// Step 5: Download artifact
	r.log("Downloading secrets artifact...")
	artifactData, err := r.client.DownloadArtifact(ctx, repo, runID, "retrieved-secrets")
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	r.log("✓ Artifact downloaded")

	// Step 6: Decode secrets
	r.log("Decoding secrets...")
	secrets, err := DecodeSecrets(artifactData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secrets: %w", err)
	}
	r.log("✓ Successfully retrieved %d secret(s)", len(secrets))

	// Step 7: Cleanup happens via defer

	return secrets, nil
}
