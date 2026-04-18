package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mohammadrendra/gsecret/internal/retriever"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <owner/repo> <secret-name> [secret-name...]",
	Short: "Retrieve GitHub secret values",
	Long: `Retrieve one or more GitHub secret values using a temporary GitHub Actions workflow.
	
The tool will:
1. Create a temporary workflow in the repository
2. Trigger it to export the specified secrets
3. Download and decode the secret values
4. Automatically clean up all traces

Examples:
  gsecret get myorg/myrepo DATABASE_URL
  gsecret get myorg/myrepo API_KEY SECRET_TOKEN
  gsecret get myorg/myrepo --all
  gsecret get myorg/myrepo DATABASE_URL --env production
  gsecret get --org myorg ORG_SECRET`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGet,
}

var (
	getAll    bool
	getEnvName string
	getOrgName string
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVar(&getAll, "all", false, "Retrieve all secrets")
	getCmd.Flags().StringVar(&getEnvName, "env", "", "Retrieve from environment")
	getCmd.Flags().StringVar(&getOrgName, "org", "", "Retrieve organization secret (provide org name)")
}

func runGet(cmd *cobra.Command, args []string) error {
	var repo string
	var secretNames []string

	if getOrgName != "" {
		// Organization secrets mode: gsecret get --org myorg SECRET_NAME
		repo = ""
		secretNames = args
	} else {
		// Repository/environment secrets mode: gsecret get owner/repo SECRET_NAME
		if len(args) < 1 {
			return fmt.Errorf("repository required")
		}
		repo = args[0]
		if !getAll && len(args) < 2 {
			return fmt.Errorf("secret name(s) required (or use --all flag)")
		}
		if !getAll {
			secretNames = args[1:]
		}
	}

	if !quiet {
		if getOrgName != "" {
			fmt.Fprintf(cmd.OutOrStderr(), "Retrieving organization secret(s) from %s...\n", getOrgName)
		} else if getEnvName != "" {
			fmt.Fprintf(cmd.OutOrStderr(), "Retrieving environment secret(s) from %s (env: %s)...\n", repo, getEnvName)
		} else if getAll {
			fmt.Fprintf(cmd.OutOrStderr(), "Retrieving all secrets from %s...\n", repo)
		} else {
			fmt.Fprintf(cmd.OutOrStderr(), "Retrieving secret(s) from %s...\n", repo)
		}
	}

	if dryRun {
		fmt.Fprintln(cmd.OutOrStderr(), "[DRY RUN] Would retrieve secrets:", strings.Join(secretNames, ", "))
		return nil
	}

	r, err := retriever.NewRetriever(!quiet)
	if err != nil {
		return fmt.Errorf("failed to initialize retriever: %w", err)
	}

	var secrets map[string]string
	
	if getOrgName != "" {
		secrets, err = r.RetrieveOrgSecrets(cmd.Context(), getOrgName, secretNames)
	} else if getEnvName != "" {
		secrets, err = r.RetrieveEnvironmentSecrets(cmd.Context(), repo, getEnvName, secretNames, getAll)
	} else {
		secrets, err = r.RetrieveRepoSecrets(cmd.Context(), repo, secretNames, getAll)
	}
	
	if err != nil {
		return fmt.Errorf("failed to retrieve secrets: %w", err)
	}

	// Output results
	if outputFile != "" {
		return writeToFile(secrets)
	}

	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(secrets)
	}

	// Plain text output
	for name, value := range secrets {
		fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", name, value)
	}

	return nil
}

func writeToFile(secrets map[string]string) error {
	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if jsonOutput {
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(secrets)
	}

	for name, value := range secrets {
		if _, err := fmt.Fprintf(f, "%s=%s\n", name, value); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}
