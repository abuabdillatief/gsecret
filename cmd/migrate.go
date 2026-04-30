package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/mohammadrendra/gsecret/internal/github"
	"github.com/mohammadrendra/gsecret/internal/retriever"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate <source-owner/repo> <target-owner/repo> <secret-name> [secret-name...]",
	Short: "Migrate GitHub repository secrets to another repository",
	Long: `Migrate one or more GitHub repository secrets from a source repository to a target repository.

The tool retrieves secret values from the source repository using the same temporary
GitHub Actions workflow as the get command, then sets each secret on the target
repository using the GitHub CLI.

Examples:
  gsecret migrate myorg/repo-a myorg/repo-b DATABASE_URL
  gsecret migrate myorg/repo-a myorg/repo-b API_KEY DATABASE_URL SECRET_TOKEN
  gsecret migrate myorg/repo-a myorg/repo-b --all`,
	Args: cobra.MinimumNArgs(2),
	RunE: runMigrate,
}

var migrateAll bool

type migrationSummary struct {
	SourceRepository string   `json:"source_repository"`
	TargetRepository string   `json:"target_repository"`
	Secrets          []string `json:"secrets"`
	MigratedCount    int      `json:"migrated_count"`
	DryRun           bool     `json:"dry_run,omitempty"`
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(&migrateAll, "all", false, "Migrate all repository secrets")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	sourceRepo := args[0]
	targetRepo := args[1]
	secretNames := args[2:]

	if err := validateRepoSlug(sourceRepo); err != nil {
		return fmt.Errorf("invalid source repository: %w", err)
	}
	if err := validateRepoSlug(targetRepo); err != nil {
		return fmt.Errorf("invalid target repository: %w", err)
	}
	if !migrateAll && len(secretNames) == 0 {
		return fmt.Errorf("secret name(s) required (or use --all flag)")
	}
	if migrateAll && len(secretNames) > 0 {
		return fmt.Errorf("cannot combine --all with explicit secret names")
	}

	if !quiet {
		if migrateAll {
			fmt.Fprintf(cmd.OutOrStderr(), "Migrating all repository secrets from %s to %s...\n", sourceRepo, targetRepo)
		} else {
			fmt.Fprintf(cmd.OutOrStderr(), "Migrating %d repository secret(s) from %s to %s...\n", len(secretNames), sourceRepo, targetRepo)
		}
	}

	if dryRun {
		names, err := dryRunMigrationSecrets(cmd, sourceRepo, secretNames)
		if err != nil {
			return err
		}
		return outputMigrationSummary(cmd, migrationSummary{
			SourceRepository: sourceRepo,
			TargetRepository: targetRepo,
			Secrets:          names,
			MigratedCount:    len(names),
			DryRun:           true,
		})
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("failed to find GitHub CLI: %w", err)
	}

	r, err := retriever.NewRetriever(!quiet)
	if err != nil {
		return fmt.Errorf("failed to initialize retriever: %w", err)
	}

	secrets, err := r.RetrieveRepoSecrets(cmd.Context(), sourceRepo, secretNames, migrateAll)
	if err != nil {
		return fmt.Errorf("failed to retrieve secrets: %w", err)
	}

	names := sortedSecretNames(secrets)
	for _, name := range names {
		if !quiet {
			fmt.Fprintf(cmd.OutOrStderr(), "Setting secret %s on %s...\n", name, targetRepo)
		}
		if err := setRepositorySecret(cmd, targetRepo, name, secrets[name]); err != nil {
			return fmt.Errorf("failed to set secret %s on %s: %w", name, targetRepo, err)
		}
	}

	return outputMigrationSummary(cmd, migrationSummary{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		Secrets:          names,
		MigratedCount:    len(names),
	})
}

func dryRunMigrationSecrets(cmd *cobra.Command, sourceRepo string, secretNames []string) ([]string, error) {
	if !migrateAll {
		names := append([]string(nil), secretNames...)
		sort.Strings(names)
		return names, nil
	}

	client, err := github.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	names, err := client.ListRepoSecrets(cmd.Context(), sourceRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to list source repository secrets: %w", err)
	}
	sort.Strings(names)

	return names, nil
}

func setRepositorySecret(cmd *cobra.Command, repo, name, value string) error {
	ghCmd := exec.CommandContext(cmd.Context(), "gh", "secret", "set", name, "--repo", repo)
	ghCmd.Stdin = strings.NewReader(value)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ghCmd.Stdout = &stdout
	ghCmd.Stderr = &stderr

	if err := ghCmd.Run(); err != nil {
		output := strings.TrimSpace(stderr.String())
		if output == "" {
			output = strings.TrimSpace(stdout.String())
		}
		if output != "" {
			return fmt.Errorf("%s: %w", output, err)
		}
		return err
	}

	return nil
}

func outputMigrationSummary(cmd *cobra.Command, summary migrationSummary) error {
	if outputFile != "" {
		return writeMigrationSummaryToFile(summary)
	}

	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}

	if summary.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[DRY RUN] Would migrate %d secret(s) from %s to %s:\n", summary.MigratedCount, summary.SourceRepository, summary.TargetRepository)
		for _, name := range summary.Secrets {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
		}
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Migrated %d secret(s) from %s to %s:\n", summary.MigratedCount, summary.SourceRepository, summary.TargetRepository)
	for _, name := range summary.Secrets {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
	}

	return nil
}

func writeMigrationSummaryToFile(summary migrationSummary) error {
	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if jsonOutput {
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}

	if summary.DryRun {
		_, err = fmt.Fprintf(f, "[DRY RUN] Would migrate %d secret(s) from %s to %s:\n", summary.MigratedCount, summary.SourceRepository, summary.TargetRepository)
	} else {
		_, err = fmt.Fprintf(f, "Migrated %d secret(s) from %s to %s:\n", summary.MigratedCount, summary.SourceRepository, summary.TargetRepository)
	}
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	for _, name := range summary.Secrets {
		if _, err := fmt.Fprintf(f, "  - %s\n", name); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}

func sortedSecretNames(secrets map[string]string) []string {
	names := make([]string, 0, len(secrets))
	for name := range secrets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func validateRepoSlug(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("expected 'owner/repo'")
	}
	return nil
}
