package retriever

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
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
	configPath := ".gsecret-config.json"
	branchName, err := generatedBranchName()
	if err != nil {
		return nil, fmt.Errorf("failed to generate branch name: %w", err)
	}

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

	// Step 2: Create configuration file with secret names
	r.log("Creating configuration file...")
	config := map[string]interface{}{
		"secret_names": secretNames,
		"secret_type":  secretType,
	}
	if envName != "" {
		config["environment_name"] = envName
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	configContent := base64.StdEncoding.EncodeToString(configJSON)
	if err := r.client.CreateFile(ctx, repo, configPath, configContent, "[gsecret] Add config", branchName); err != nil {
		return nil, fmt.Errorf("failed to create config file: %w", err)
	}
	cleaner.AddWorkflowFile(configPath)
	r.log("✓ Configuration created")

	// Step 3: Create workflow file - this push will trigger the workflow
	r.log("Creating workflow file (this will trigger the workflow)...")
	if err := r.client.CreateWorkflowFileFromTemplate(ctx, repo, workflowPath, branchName); err != nil {
		return nil, fmt.Errorf("failed to create workflow file: %w", err)
	}
	cleaner.AddWorkflowFile(workflowPath)
	r.log("✓ Workflow file created and triggered")

	// Step 4: Wait a bit for workflow to start
	r.log("Waiting for workflow to start...")
	time.Sleep(5 * time.Second)

	// Step 5: Get the latest workflow run
	r.log("Finding workflow run...")
	runID, err := r.client.GetLatestWorkflowRun(ctx, repo, workflowFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}
	cleaner.AddWorkflowRun(runID)
	r.log("✓ Workflow run found (ID: %d)", runID)

	// Step 6: Wait for completion
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

func generatedBranchName() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("gsecret-retrieval-generated-%06d", n.Int64()), nil
}
