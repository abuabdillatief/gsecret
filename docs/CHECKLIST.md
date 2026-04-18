# Pre-Use Checklist

Before using gsecret, verify:

## ✅ Prerequisites

- [ ] GitHub CLI installed: `gh --version`
- [ ] Authenticated with GitHub: `gh auth status`
- [ ] Repository has secrets set up (or you're testing with a known secret)
- [ ] You have write access to the repository
- [ ] GitHub Actions is enabled in the repository

## ✅ First-Time Setup

1. **Install using Makefile** (recommended):
   ```bash
   make check         # Verify Go and gh are installed
   make install-user  # Install to ~/bin (no sudo needed)
   ```
   
   Or build manually:
   ```bash
   go build -o gsecret .
   ```

2. **Test authentication**:
   ```bash
   gh auth status
   # Should show: ✓ Logged in to github.com
   ```

3. **Test with list command** (safe, doesn't retrieve values):
   ```bash
   ./gsecret list YOUR_ORG/YOUR_REPO
   ```

4. **Check a secret exists**:
   - Go to GitHub.com → Your repo → Settings → Secrets and variables → Actions
   - Note the name of at least one secret (case-sensitive!)

## ✅ First Retrieval

1. **Try dry-run first**:
   ```bash
   ./gsecret get YOUR_ORG/YOUR_REPO SECRET_NAME --dry-run
   ```

2. **Retrieve a single secret**:
   ```bash
   ./gsecret get YOUR_ORG/YOUR_REPO SECRET_NAME
   ```

3. **If it fails**:
   - Go to repo → Actions tab
   - Find the "gsecret-retrieval-temp" workflow run
   - Click on it to see the logs
   - Look for error messages in the "Export secrets to artifact" step

4. **Verify cleanup**:
   - Check Actions tab - the workflow run should be gone
   - Check `.github/workflows/` - no `gsecret-temp-*.yml` files should remain

## ✅ If Workflow Fails

**The workflow logs will tell you exactly what's wrong. Common issues:**

- ✗ Secret NAME_HERE not found or empty
  → Secret doesn't exist or name is misspelled

- ✗ Environment 'production' not found  
  → Environment name is wrong or doesn't exist

- Error: No secrets found
  → Repository has no secrets (when using --all)

**Always check the workflow logs first!**

## ✅ For Production Use

- [ ] Tested on a non-critical repository first
- [ ] Verified cleanup works (checked Actions tab after retrieval)
- [ ] Understand security implications (see README.md)
- [ ] Know where secrets will be saved (stdout by default, or use `-o` flag)
- [ ] Plan to rotate secrets if security is a concern

## ✅ Safe Usage Pattern

```bash
# 1. List secrets (safe - no retrieval)
./gsecret list org/repo

# 2. Dry run (safe - doesn't execute)
./gsecret get org/repo SECRET_NAME --dry-run

# 3. Retrieve (creates workflow, retrieves, cleans up)
./gsecret get org/repo SECRET_NAME

# 4. Verify cleanup
# - Check repo Actions tab - workflow run should be deleted
# - Check repo files - no workflow file should remain

# 5. If saving to file, secure it
./gsecret get org/repo --all -o secrets.env
chmod 600 secrets.env  # Already done by tool, but verify
```

## ✅ Emergency Cleanup

If automatic cleanup fails:

```bash
# 1. Delete workflow file manually
# Go to repo → .github/workflows/ → Delete gsecret-temp-*.yml

# 2. Delete workflow runs
# Go to repo → Actions → Select run → Delete workflow run

# 3. Or use gh CLI:
gh api repos/OWNER/REPO/contents/.github/workflows/gsecret-temp-TIMESTAMP.yml \
  --method DELETE \
  -f message="Cleanup" \
  -f sha=FILE_SHA
```

## ✅ Success Indicators

You'll know it worked when:
- ✓ Tool prints progress messages
- ✓ Workflow completes successfully
- ✓ Secrets are printed to terminal (or saved to file)
- ✓ Cleanup messages appear
- ✓ No workflow files remain in repository
- ✓ No workflow runs remain in Actions tab (they're deleted)

## Need Help?

- Read [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for detailed debugging
- Read [README.md](README.md) for comprehensive documentation  
- Check workflow logs in Actions tab (most helpful!)
