package cleanup

import (
	"context"
	"fmt"

	"github.com/mohammadrendra/gsecret/internal/github"
)

type Cleanup struct {
	client        *github.Client
	repo          string
	workflowFiles []string
	workflowRuns  []int64
	branches      []string
	verbose       bool
}

func NewCleanup(client *github.Client, repo string) *Cleanup {
	return &Cleanup{
		client:        client,
		repo:          repo,
		workflowFiles: make([]string, 0),
		workflowRuns:  make([]int64, 0),
		branches:      make([]string, 0),
	}
}

func (c *Cleanup) SetVerbose(verbose bool) {
	c.verbose = verbose
}

func (c *Cleanup) AddWorkflowFile(path string) {
	c.workflowFiles = append(c.workflowFiles, path)
}

func (c *Cleanup) AddWorkflowRun(runID int64) {
	c.workflowRuns = append(c.workflowRuns, runID)
}

func (c *Cleanup) AddBranch(name string) {
	c.branches = append(c.branches, name)
}

func (c *Cleanup) log(format string, args ...interface{}) {
	if c.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// Cleanup removes all temporary resources
func (c *Cleanup) Cleanup(ctx context.Context) error {
	c.log("Cleaning up temporary resources...")

	var errors []error

	// Delete workflow runs (and their artifacts)
	for _, runID := range c.workflowRuns {
		c.log("  Deleting workflow run %d...", runID)
		if err := c.client.DeleteWorkflowRun(ctx, c.repo, runID); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete workflow run %d: %w", runID, err))
		} else {
			c.log("  ✓ Workflow run deleted")
		}
	}

	// Delete branches first - this removes all files on the branch
	// (including .gsecret-config.json and workflow files)
	for _, branch := range c.branches {
		c.log("  Deleting branch %s...", branch)
		if err := c.client.DeleteBranch(ctx, c.repo, branch); err != nil {
			// Don't fail if branch deletion fails - it might have other uses
			c.log("  ⚠ Could not delete branch %s: %v", branch, err)
		} else {
			c.log("  ✓ Branch deleted (including all files on it)")
		}
	}

	// Note: We don't delete workflow files individually because:
	// 1. If they're on the temporary branch, deleting the branch removes them
	// 2. If they're on main branch (shouldn't be), we don't want to touch main
	// The workflowFiles list is kept for reference but not used in cleanup

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d error(s): %v", len(errors), errors)
	}

	c.log("✓ Cleanup completed successfully")
	return nil
}
