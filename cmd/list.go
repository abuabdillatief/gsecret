package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mohammadrendra/gsecret/internal/github"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <owner/repo>",
	Short: "List available secret names in a repository",
	Long: `List all secret names (without values) in a repository.
	
Examples:
  gsecret list myorg/myrepo
  gsecret list myorg/myrepo --json
  gsecret list myorg/myrepo --env production`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

var (
	envName string
	orgName string
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&envName, "env", "", "List environment secrets")
	listCmd.Flags().StringVar(&orgName, "org", "", "List organization secrets (provide org name)")
}

func runList(cmd *cobra.Command, args []string) error {
	repo := args[0]
	
	if !quiet {
		if envName != "" {
			fmt.Fprintf(cmd.OutOrStderr(), "Listing environment secrets for %s (env: %s)...\n", repo, envName)
		} else if orgName != "" {
			fmt.Fprintf(cmd.OutOrStderr(), "Listing organization secrets for %s...\n", orgName)
		} else {
			fmt.Fprintf(cmd.OutOrStderr(), "Listing repository secrets for %s...\n", repo)
		}
	}

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	var secrets []string
	
	if orgName != "" {
		secrets, err = client.ListOrgSecrets(cmd.Context(), orgName)
	} else if envName != "" {
		secrets, err = client.ListEnvironmentSecrets(cmd.Context(), repo, envName)
	} else {
		secrets, err = client.ListRepoSecrets(cmd.Context(), repo)
	}
	
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"repository": repo,
			"secrets":    secrets,
		}
		if envName != "" {
			output["environment"] = envName
		}
		if orgName != "" {
			output["organization"] = orgName
		}
		
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	if len(secrets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No secrets found.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nFound %d secret(s):\n", len(secrets))
	for _, secret := range secrets {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", secret)
	}

	return nil
}
