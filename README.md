# gsecret - GitHub Secrets Retrieval Tool

A CLI tool to retrieve GitHub secret values that are already stored but not accessible through the GitHub API.

## Problem

GitHub's API doesn't expose secret values (by design for security). This makes it difficult to:
- Migrate secrets to another platform
- Recover secrets that weren't backed up elsewhere
- Audit or verify secret values

Current workarounds are hacky (e.g., SSH into a VPS, create a workflow to log secrets manually).

## Solution

`gsecret` automates the retrieval process by:
1. Creating a temporary GitHub Actions workflow in your repository
2. Triggering the workflow to safely export secrets via artifacts
3. Downloading and decoding the secret values
4. Automatically cleaning up all traces (workflow file, runs, artifacts)

## Features

- ✅ Retrieve repository secrets
- ✅ Retrieve environment secrets
- ✅ Retrieve organization secrets (requires repo with access)
- ✅ List secret names without values
- ✅ Retrieve single or multiple secrets
- ✅ Retrieve all secrets at once
- ✅ JSON output format
- ✅ Output to file with secure permissions
- ✅ Automatic cleanup of temporary resources
- ✅ Dry-run mode for testing
- ✅ Progress indicators

## Installation

**See [INSTALL.md](INSTALL.md) for detailed installation instructions.**

### Quick Install (Recommended)

The easiest way to install gsecret is using the Makefile:

```bash
# Clone the repository
git clone https://github.com/mohammadrendra/gsecret.git
cd gsecret

# Check prerequisites
make check

# Install to ~/bin (no sudo required)
make install-user

# Or install system-wide to /usr/local/bin (requires sudo)
make install
```

After installation, you can run `gsecret` from anywhere!

### Manual Installation

#### Option 1: Build from source
```bash
git clone https://github.com/mohammadrendra/gsecret.git
cd gsecret
go build -o gsecret .
sudo mv gsecret /usr/local/bin/
```

#### Option 2: Go install
```bash
go install github.com/mohammadrendra/gsecret@latest
```

### Prerequisites

1. **Install GitHub CLI**: `gsecret` uses `gh` for authentication
   ```bash
   # macOS
   brew install gh
   
   # Linux
   curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
   sudo apt update
   sudo apt install gh
   
   # Windows
   winget install GitHub.cli
   ```

2. **Authenticate with GitHub**:
   ```bash
   gh auth login
   ```
   
   Make sure you grant the following scopes:
   - `repo` - Access repositories
   - `workflow` - Manage workflows
   - `admin:org` - Access organization secrets (if needed)

### Install gsecret

#### Option 1: Install from source (requires Go 1.21+)
```bash
git clone https://github.com/mohammadrendra/gsecret.git
cd gsecret
go install
```

#### Option 2: Build binary
```bash
git clone https://github.com/mohammadrendra/gsecret.git
cd gsecret
go build -o gsecret
sudo mv gsecret /usr/local/bin/
```

## Usage

### List Secrets

List all secret names (without values) in a repository:

```bash
# Repository secrets
gsecret list owner/repo

# Environment secrets
gsecret list owner/repo --env production

# Organization secrets
gsecret list --org myorg owner/repo
```

### Retrieve Secrets

Retrieve one or more secret values:

```bash
# Single secret
gsecret get owner/repo DATABASE_URL

# Multiple secrets
gsecret get owner/repo API_KEY DATABASE_URL SECRET_TOKEN

# All secrets
gsecret get owner/repo --all

# Environment secret
gsecret get owner/repo DATABASE_URL --env production

# Organization secret (requires repo with access)
gsecret get owner/repo ORG_SECRET --org myorg
```

### Output Options

```bash
# JSON format
gsecret get owner/repo DATABASE_URL --json

# Save to file (with secure 0600 permissions)
gsecret get owner/repo DATABASE_URL -o secrets.env

# JSON to file
gsecret get owner/repo --all --json -o secrets.json

# Quiet mode (suppress progress messages)
gsecret get owner/repo DATABASE_URL --quiet
```

### Dry Run

Test what would happen without actually executing:

```bash
gsecret get owner/repo DATABASE_URL --dry-run
```

## How It Works

1. **Create Branch**: Creates a generated temporary branch like `gsecret-retrieval-generated-123456` from your default branch
2. **Create Workflow**: Pushes a temporary workflow file to that branch (`.github/workflows/gsecret-temp-TIMESTAMP.yml`)
3. **Trigger**: Triggers the workflow via API with secret names as inputs
4. **Export**: Workflow accesses secrets and exports them base64-encoded to artifacts
5. **Download**: Downloads the artifact containing the encoded secrets
6. **Decode**: Decodes the base64 values and returns plain text
7. **Cleanup**: Automatically deletes:
   - Workflow run and artifacts
   - Workflow file
   - The entire generated temporary branch

**Your main branch stays completely clean** - all activity happens on an isolated branch that's deleted after use.

## Security Considerations

⚠️ **Important Security Notes**:

- Secrets are briefly exposed in GitHub Actions artifacts (encrypted by GitHub)
- All operations happen on an isolated generated branch like `gsecret-retrieval-generated-123456` (keeps main clean)
- The temporary branch is automatically deleted after retrieval
- Workflow runs and artifacts are deleted immediately after retrieval
- The tool requires write access to create/delete workflows and branches
- Use with caution in production environments
- Consider using in a dedicated test repository first
- Secrets are transmitted through GitHub's infrastructure only

**Best Practices**:
- Run in private repositories only
- Review the temporary workflow file before it's created (`--dry-run`)
- Verify cleanup succeeded (check Actions tab in GitHub)
- Don't share `gsecret` output publicly
- Use secure file permissions when saving to file (0600)
- Rotate secrets after retrieval if concerned about exposure

## Troubleshooting

**If you encounter any issues, see the comprehensive [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) guide.**

### Quick Diagnostics

If `gsecret get` fails with "workflow failed with conclusion: failure":

1. **Check the workflow logs**: Go to your repo → Actions tab → Find the failed run
2. **Verify secret exists**: Run `gsecret list owner/repo` first
3. **Try a single secret**: Don't use `--all` initially, specify a secret name
4. **Check the logs output**: The workflow logs show exactly what went wrong

### Common Quick Fixes

### "failed to create GitHub client"
- Ensure `gh` is installed and authenticated: `gh auth status`
- Re-authenticate: `gh auth login`

### "failed to create workflow file"
- Check you have write access to the repository
- Verify repository exists and is accessible

### "failed to trigger workflow"
- Wait a few seconds after creating the workflow file
- Check if `.github/workflows/` directory exists
- Verify workflow file was created correctly

### "workflow failed with conclusion: failure"
This is the most common issue. To debug:

1. **Check the workflow logs in GitHub**:
   - Go to your repository on GitHub
   - Click "Actions" tab
   - Find the workflow run named "gsecret-retrieval-temp"
   - Click on it to see the logs
   
2. **Common causes**:
   - **Secret doesn't exist**: The secret name might be misspelled or doesn't exist in the repository
   - **No secrets in repository**: If using `--all`, the repository must have at least one secret
   - **Environment doesn't exist**: When using `--env`, the environment name must match exactly
   - **No access to secrets**: Workflow might not have access to the secrets (check repository settings)
   
3. **Quick fixes**:
   ```bash
   # First, verify secrets exist
   gsecret list owner/repo
   
   # Then try retrieving a specific secret (not --all)
   gsecret get owner/repo SECRET_NAME
   
   # If it still fails, check the Actions tab for detailed logs
   ```

### "failed to download artifact"
- Artifact might not have been created
- Check workflow logs in GitHub Actions tab
- Ensure workflow completed successfully

### Cleanup didn't complete
- Manually delete workflow file in `.github/workflows/gsecret-temp-*.yml`
- Manually delete workflow runs in repository Actions tab
- Check repository permissions

### Testing the tool safely

Before using on important repositories:
```bash
# 1. Create a test repository with a test secret
# 2. Try listing secrets first
gsecret list test-user/test-repo

# 3. Try dry-run mode
gsecret get test-user/test-repo TEST_SECRET --dry-run

# 4. Try retrieving the test secret
gsecret get test-user/test-repo TEST_SECRET

# 5. Verify cleanup in Actions tab
```

## Examples

### Export all secrets

```bash
# Export all secrets to .env file
gsecret get owner/repo --all -o .env

# Export as JSON for programmatic use
gsecret get owner/repo --all --json -o secrets.json
```

### Verify a specific secret value

```bash
gsecret get owner/repo DATABASE_URL
# Output: DATABASE_URL=postgresql://user:pass@host:5432/db
```

### Backup all environment secrets

```bash
gsecret get owner/repo --all --env production -o production-secrets.env
gsecret get owner/repo --all --env staging -o staging-secrets.env
```

## Limitations

- Requires GitHub Actions to be enabled in the repository
- Subject to GitHub Actions usage limits
- Organization secrets require specifying a repository with access
- Secrets larger than GitHub's artifact size limits cannot be retrieved
- Rate limits apply (GitHub API and Actions)


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Disclaimer

This tool is provided as-is. Use at your own risk. Always follow your organization's security policies when handling secrets.

## Credits

Created to solve the problem of retrieving GitHub secrets without hacky workarounds.

Built with:
- [github.com/cli/go-gh](https://github.com/cli/go-gh) - GitHub CLI library
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework

## Makefile Commands

The project includes a Makefile for easy building and installation:

```bash
make help           # Show all available commands
make check          # Check prerequisites (Go, gh CLI, authentication)
make build          # Build the binary
make install        # Install to /usr/local/bin (requires sudo)
make install-user   # Install to ~/bin (no sudo required)
make uninstall      # Remove from /usr/local/bin
make uninstall-user # Remove from ~/bin
make clean          # Remove build artifacts
make verify         # Verify installation is working
make dev            # Build with debug symbols
make test           # Run tests
```

**Recommended workflow:**
```bash
make check          # Verify prerequisites first
make install-user   # Install without sudo
gsecret --help      # Test it works
```

## Branch Isolation

To keep your main branch clean, gsecret uses a dedicated temporary branch:

- **Branch name**: `gsecret-retrieval-generated-<random-number>`
- **Created from**: Your default branch (main/master)
- **Contains**: Only the temporary workflow file
- **Lifetime**: Exists only during secret retrieval
- **Cleanup**: Completely deleted after retrieval

This means **zero commits to your main branch** - all gsecret activity is isolated and cleaned up automatically.
