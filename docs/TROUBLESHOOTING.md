# Troubleshooting Guide

## Common Issues and Solutions

### 1. Workflow Failed with "conclusion: failure"

**Symptoms**: Tool says "workflow failed with conclusion: failure" after cleanup completes.

**How to diagnose**:
1. Go to your GitHub repository → Actions tab
2. Find the workflow run (it will have a red X)
3. Click on it to see the detailed logs
4. Look at the "Export secrets to artifact" step

**Common causes and fixes**:

#### Secret doesn't exist
```
Error in logs: "Secret SECRET_NAME not found or empty"
```
**Fix**: 
- Check secret name spelling (case-sensitive!)
- Verify secret exists: `gsecret list owner/repo`
- Make sure you're using the correct repository

#### No secrets in repository
```
Error: Using --all but repository has no secrets
```
**Fix**: 
- Add at least one secret to the repository first
- Or specify a secret name instead of using --all

#### Environment doesn't exist
```
Error: Environment 'production' not found
```
**Fix**:
- Check environment name spelling (case-sensitive!)
- Verify environment exists in repository settings
- List environments in GitHub: Settings → Environments

#### Secrets not accessible to workflow
```
Error: Secrets context is empty
```
**Fix**:
- Check repository settings → Secrets and variables → Actions
- Ensure secrets are set at the repository level (not environment) for repo secrets
- For environment secrets, ensure the environment name is correct

### 2. Authentication Issues

**Symptoms**: "failed to create GitHub client"

**Solutions**:
```bash
# Check authentication status
gh auth status

# Re-authenticate
gh auth login

# Refresh token with required scopes
gh auth refresh -h github.com -s repo,workflow
```

### 3. Permission Issues

**Symptoms**: "failed to create workflow file" or "403 Forbidden"

**Solutions**:
- Ensure you have write access to the repository
- Check you're authenticated as the correct user: `gh auth status`
- Verify the repository exists: `gh repo view owner/repo`

### 4. Workflow Not Triggering

**Symptoms**: Tool hangs at "Triggering workflow..."

**Solutions**:
- GitHub Actions must be enabled in the repository
- Check repository settings → Actions → Allow all actions
- Ensure `.github/workflows/` directory exists (tool creates it)
- Wait a few seconds and try again

### 5. Artifact Download Fails

**Symptoms**: "failed to download artifact" or "artifact not found"

**Solutions**:
- Check if workflow completed successfully (not failed/cancelled)
- Verify artifact was created in the workflow logs
- GitHub artifact retention is set to 1 day - very rarely it might expire

## Step-by-Step Debugging

### Step 1: Verify Authentication
```bash
gh auth status
# Should show: Logged in to github.com as YOUR_USERNAME
```

### Step 2: Verify Repository Access
```bash
gh repo view owner/repo
# Should show repository details
```

### Step 3: List Secrets (Doesn't Retrieve Values)
```bash
gsecret list owner/repo
# Should list all secret names
```

### Step 4: Try Dry Run
```bash
gsecret get owner/repo SECRET_NAME --dry-run
# Should show what would happen
```

### Step 5: Retrieve Single Secret
```bash
gsecret get owner/repo SECRET_NAME
# If this fails, check Actions tab for logs
```

### Step 6: Check Actions Tab
1. Go to `https://github.com/owner/repo/actions`
2. Find the "gsecret-retrieval-temp" workflow run
3. Click on it
4. Click "retrieve-secrets" job
5. Click "Export secrets to artifact" step
6. Read the logs - they show exactly what happened

## What the Logs Show

Good workflow output:
```
Exporting secrets...
Secret type: repo
Secrets JSON input: ["DATABASE_URL"]
Secret names to retrieve:
DATABASE_URL
Processing secret: DATABASE_URL
✓ Exported DATABASE_URL

Secrets exported successfully
DATABASE_URL=dGVzdC12YWx1ZQ==
```

Failed workflow output (secret not found):
```
Exporting secrets...
Secret type: repo
Secrets JSON input: ["DATABASE_URL"]
Secret names to retrieve:
DATABASE_URL
Processing secret: DATABASE_URL
✗ Secret DATABASE_URL not found or empty

Secrets exported successfully
DATABASE_URL=__NOT_FOUND__
```

## Testing in a Safe Environment

Create a test repository first:
```bash
# 1. Create test repo (on GitHub web UI or via gh CLI)
gh repo create test-gsecret --private

# 2. Add a test secret via web UI or CLI
gh secret set TEST_SECRET -b "test-value" --repo owner/test-gsecret

# 3. Test listing
gsecret list owner/test-gsecret

# 4. Test retrieval
gsecret get owner/test-gsecret TEST_SECRET

# 5. Verify cleanup in Actions tab
```

## Still Having Issues?

If none of the above helps:

1. **Enable debug output**: Run with verbose logging
   ```bash
   gsecret get owner/repo SECRET_NAME 2>&1 | tee debug.log
   ```

2. **Check workflow file was created**: Look for `.github/workflows/gsecret-temp-*.yml` in your repo

3. **Verify GitHub Actions is working**: Create a simple test workflow manually to ensure Actions work

4. **Check GitHub status**: Visit https://www.githubstatus.com to ensure GitHub Actions is operational

5. **Manual cleanup if needed**:
   ```bash
   # Delete any leftover workflow files
   gh api repos/owner/repo/contents/.github/workflows/gsecret-temp-TIMESTAMP.yml \
     --method DELETE \
     -f message="Manual cleanup" \
     -f sha=FILE_SHA
   ```

## Common Questions

**Q: Is it safe to use?**
A: Yes, but only in private repositories. Secrets are briefly in artifacts but cleaned up immediately. Always verify cleanup in the Actions tab after use.

**Q: Can I use it in public repositories?**
A: Not recommended. While it cleans up, secrets would briefly be in Actions artifacts.

**Q: Why does it need workflow write access?**
A: To create and delete the temporary workflow file.

**Q: Can I retrieve secrets without GitHub Actions?**
A: No. GitHub's API intentionally doesn't expose secret values. Actions is the only way to access them programmatically.

**Q: What if cleanup fails?**
A: Manually delete the workflow file and runs from the Actions tab. The tool shows which file it created.
