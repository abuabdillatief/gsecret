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
	verbose       bool
}

func NewCleanup(client *github.Client, repo string) *Cleanup {
	return &Cleanup{
		client:        client,
		repo:          repo,
		workflowFiles: make([]string, 0),
		workflowRuns:  make([]int64, 0),
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

	// Delete workflow files
	for _, path := range c.workflowFiles {
		c.log("  Deleting workflow file %s...", path)
		if err := c.client.DeleteWorkflowFile(ctx, c.repo, path, "[gsecret] Remove temporary secret retrieval workflow"); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete workflow file %s: %w", path, err))
		} else {
			c.log("  ✓ Workflow file deleted")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d error(s): %v", len(errors), errors)
	}

	c.log("✓ Cleanup completed successfully")
	return nil
}
