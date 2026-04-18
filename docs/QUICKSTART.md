# Quick Start Guide

## Installation

**With Makefile (Recommended):**
```bash
git clone https://github.com/mohammadrendra/gsecret.git
cd gsecret
make check          # Verify prerequisites  
make install-user   # Install to ~/bin (no sudo)
# OR
make install        # Install to /usr/local/bin (requires sudo)
```

**Verify installation:**
```bash
gsecret --version
gsecret --help
```

**Manual build:**
```bash
go build -o gsecret .
sudo mv gsecret /usr/local/bin/
```

See [README.md](../README.md) for detailed installation instructions.

## Authentication Setup

1. **Install GitHub CLI** (if not already installed)
   ```bash
   # macOS
   brew install gh
   
   # Linux
   # See README.md for Linux installation instructions
   ```

2. **Authenticate with GitHub**
   ```bash
   gh auth login
   ```
   
   Select:
   - GitHub.com
   - HTTPS protocol
   - Authenticate via web browser
   - Grant necessary scopes (repo, workflow)

3. **Verify authentication**
   ```bash
   gh auth status
   ```

## Basic Usage

### List Secrets
```bash
# List repository secrets (names only, no values)
gsecret list owner/repo

# List with JSON output
gsecret list owner/repo --json

# List environment secrets
gsecret list owner/repo --env production
```

### Retrieve Secrets
```bash
# Get a single secret
gsecret get owner/repo DATABASE_URL

# Get multiple secrets
gsecret get owner/repo API_KEY DATABASE_URL

# Get all secrets
gsecret get owner/repo --all

# Get with JSON output
gsecret get owner/repo DATABASE_URL --json

# Save to file
gsecret get owner/repo --all -o secrets.env

# Quiet mode (no progress messages)
gsecret get owner/repo DATABASE_URL --quiet
```

### Environment Secrets
```bash
# List environment secrets
gsecret list owner/repo --env production

# Get environment secret
gsecret get owner/repo DATABASE_URL --env production

# Get all environment secrets
gsecret get owner/repo --all --env production
```

## How It Works

1. **Temporary Workflow**: Creates `.github/workflows/gsecret-temp-TIMESTAMP.yml`
2. **Trigger**: Triggers the workflow via GitHub API
3. **Export**: Workflow exports secrets as base64-encoded artifacts
4. **Download**: Downloads and decodes the artifact
5. **Cleanup**: Automatically removes workflow file, run, and artifacts

## Important Notes

⚠️ **Security Considerations**:
- Only use in private repositories
- Secrets are briefly exposed in GitHub Actions artifacts
- All traces are cleaned up automatically
- Verify cleanup in repository Actions tab

## Troubleshooting

### Authentication Issues
```bash
# Check auth status
gh auth status

# Re-authenticate
gh auth login

# Check token has correct scopes
gh auth refresh -h github.com -s repo,workflow
```

### Workflow Creation Issues
- Ensure you have write access to the repository
- Check if `.github/workflows/` directory exists
- Try with `--dry-run` first to test

### Cleanup Issues
- Check repository Actions tab for workflow runs
- Manually delete any remaining workflow files if needed

## Examples

### Migrate to Another Platform
```bash
# Export all secrets as .env file
gsecret get myorg/myrepo --all -o .env

# Export as JSON for scripting
gsecret get myorg/myrepo --all --json -o secrets.json
```

### Backup Environment Secrets
```bash
gsecret get myorg/myrepo --all --env production -o prod-backup.env
gsecret get myorg/myrepo --all --env staging -o staging-backup.env
```

### Quick Secret Verification
```bash
gsecret get myorg/myrepo DATABASE_URL --quiet
```

## Next Steps

- Read the full [README.md](README.md) for comprehensive documentation
- Review security considerations before use in production
- Test with `--dry-run` flag first
- Check the [GitHub Actions](https://github.com/features/actions) documentation

## Support

For issues or questions:
1. Check the [README.md](README.md) troubleshooting section
2. Review GitHub Actions logs in your repository
3. Open an issue on the repository

---

**Happy Secret Retrieving!** 🔐
