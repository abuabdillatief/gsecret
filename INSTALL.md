# Installation Guide

## Quick Install (Recommended)

### Using Makefile

The easiest way to install gsecret:

```bash
# 1. Clone the repository
git clone https://github.com/mohammadrendra/gsecret.git
cd gsecret

# 2. Check prerequisites
make check

# 3. Install (choose one)
make install-user    # Install to ~/bin (no sudo, recommended)
# OR
make install         # Install to /usr/local/bin (requires sudo)

# 4. Verify installation
gsecret --version
```

**That's it!** You can now run `gsecret` from anywhere.

---

## Installation Methods

### Method 1: User Install (~/bin) - Recommended

**Pros:** No sudo required, easy to manage
**Cons:** Need to ensure ~/bin is in PATH

```bash
make install-user
```

This installs to `~/bin/gsecret`. If you get "command not found", add this to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/bin:$PATH"
```

Then reload your shell:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### Method 2: System-Wide Install (/usr/local/bin)

**Pros:** Automatically in PATH, available for all users
**Cons:** Requires sudo

```bash
make install
```

This installs to `/usr/local/bin/gsecret` (standard location for user-installed binaries).

### Method 3: Manual Build Only

If you just want the binary in the current directory:

```bash
make build
# Binary is now at ./gsecret
./gsecret --help
```

---

## Uninstallation

Remove gsecret when you're done:

```bash
# If installed to ~/bin
make uninstall-user

# If installed to /usr/local/bin
make uninstall
```

---

## Verification

After installation, verify everything works:

```bash
# Check if gsecret is in PATH
which gsecret

# Check version
gsecret --version

# View help
gsecret --help

# Or use make verify
make verify
```

---

## Prerequisites

### Required

1. **Go 1.21+**
   ```bash
   # Check if installed
   go version
   
   # Install on macOS
   brew install go
   
   # Install on Linux
   # See: https://go.dev/doc/install
   ```

2. **GitHub CLI (gh)**
   ```bash
   # Check if installed
   gh --version
   
   # Install on macOS
   brew install gh
   
   # Install on Linux (Debian/Ubuntu)
   curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
   sudo apt update && sudo apt install gh
   
   # Install on other systems
   # See: https://github.com/cli/cli#installation
   ```

3. **Authentication**
   ```bash
   # Authenticate with GitHub
   gh auth login
   
   # Verify authentication
   gh auth status
   ```

**Check all prerequisites at once:**
```bash
make check
```

---

## Troubleshooting Installation

### "make: command not found"

Install make:
```bash
# macOS (usually pre-installed)
xcode-select --install

# Linux (Debian/Ubuntu)
sudo apt install build-essential

# Linux (RHEL/CentOS)
sudo yum groupinstall "Development Tools"
```

### "gsecret: command not found" after installation

**If installed with `make install-user`:**
1. Check if ~/bin exists: `ls -la ~/bin`
2. Check if gsecret is there: `ls -la ~/bin/gsecret`
3. Check your PATH: `echo $PATH | grep "$HOME/bin"`
4. If ~/bin is not in PATH, add to ~/.bashrc or ~/.zshrc:
   ```bash
   export PATH="$HOME/bin:$PATH"
   ```
5. Reload shell: `source ~/.bashrc` or restart terminal

**If installed with `make install`:**
1. Check if installed: `ls -la /usr/local/bin/gsecret`
2. Check if /usr/local/bin is in PATH: `echo $PATH | grep /usr/local/bin`
3. If not, add to ~/.bashrc or ~/.zshrc:
   ```bash
   export PATH="/usr/local/bin:$PATH"
   ```

### Permission denied during "make install"

The system-wide install requires sudo. Try:
```bash
make install
# Enter your password when prompted
```

Or use user install instead:
```bash
make install-user  # No sudo needed
```

### "go: command not found"

Go is not installed. Install it:
- macOS: `brew install go`
- Linux: Follow https://go.dev/doc/install
- Windows: Download from https://go.dev/dl/

### "gh: command not found"

GitHub CLI is not installed. Install it:
- macOS: `brew install gh`
- Linux: See https://github.com/cli/cli#installation
- Windows: `winget install GitHub.cli`

---

## Alternative Installation Methods

### Go Install (if you have Go already)

```bash
go install github.com/mohammadrendra/gsecret@latest
```

Binary will be in `$GOPATH/bin` (usually `~/go/bin`). Make sure it's in your PATH.

### Download Pre-built Binary

_(Coming soon: releases with pre-built binaries for multiple platforms)_

---

## Platform-Specific Notes

### macOS

- Recommended: `make install-user` (no sudo)
- /usr/local/bin is usually in PATH by default
- You may need to approve the binary in System Preferences → Security & Privacy

### Linux

- Recommended: `make install-user` (no sudo)
- Make sure ~/bin is in your PATH
- On some distros, you may need to install make first

### Windows

- Use WSL (Windows Subsystem for Linux) or Git Bash
- Or build with: `go build -o gsecret.exe .`
- Place binary in a directory in your PATH

---

## Development Installation

For development with debug symbols:

```bash
make dev
```

---

## Quick Reference

```bash
# Prerequisites
make check

# Install (user)
make install-user

# Install (system)
make install

# Build only
make build

# Clean
make clean

# Uninstall
make uninstall-user  # or make uninstall

# Verify
make verify

# Help
make help
```

---

## Next Steps

After installation:
1. Read [docs/QUICKSTART.md](docs/QUICKSTART.md) for usage
2. Read [README.md](README.md) for full documentation
3. Test with a repository: `gsecret list owner/repo`

---

## Support

If you have installation issues:
1. Run `make check` to verify prerequisites
2. Read the troubleshooting section above
3. Check [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for runtime issues
4. Open an issue on GitHub with details from `make check`
