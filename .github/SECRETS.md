# GitHub Secrets Configuration for Oracle Engine
# =============================================

This document lists all the secrets that need to be configured in your GitHub repository
for the CI/CD pipeline to work correctly.

## How to Add Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add each secret listed below

---

## Required Secrets

### Container Registry (Automatic)

| Secret | Description | Example |
|--------|-------------|---------|
| `GITHUB_TOKEN` | Automatically provided by GitHub Actions | N/A (auto) |

> Note: `GITHUB_TOKEN` is automatically provided and has permissions to push to GitHub Container Registry (ghcr.io).

---

### Deployment Secrets

#### Staging Environment

| Secret | Description | Example |
|--------|-------------|---------|
| `STAGING_HOST` | Hostname or IP of staging server | `staging.api.ifalabs.com` or `192.168.1.100` |
| `STAGING_USER` | SSH username for staging server | `oracle` |
| `STAGING_SSH_KEY` | Private SSH key for staging server | `-----BEGIN OPENSSH PRIVATE KEY-----...` |

#### Production Environment

| Secret | Description | Example |
|--------|-------------|---------|
| `PRODUCTION_HOST` | Hostname or IP of production server | `api.ifalabs.com` or `10.0.0.50` |
| `PRODUCTION_USER` | SSH username for production server | `oracle` |
| `PRODUCTION_SSH_KEY` | Private SSH key for production server | `-----BEGIN OPENSSH PRIVATE KEY-----...` |

---

### Optional Secrets

#### Code Coverage

| Secret | Description | Example |
|--------|-------------|---------|
| `CODECOV_TOKEN` | Token for uploading coverage to Codecov | `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` |

> Get this from [codecov.io](https://codecov.io) after connecting your repository.

#### Notifications

| Secret | Description | Example |
|--------|-------------|---------|
| `SLACK_WEBHOOK_URL` | Slack webhook for deployment notifications | `https://hooks.slack.com/services/T.../B.../xxx` |

> Create a Slack webhook at [api.slack.com/apps](https://api.slack.com/apps) → Your App → Incoming Webhooks

---

## Environment Secrets (Server-side)

These secrets should be configured in the `.env` file on your deployment servers,
NOT in GitHub Secrets (as they contain sensitive runtime configuration).

### Database Configuration

```bash
POSTGRES_USER=oracle
POSTGRES_PASSWORD=<strong-random-password>
POSTGRES_DB=oracle_engine
DB_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@timescale:5432/${POSTGRES_DB}?sslmode=disable
```

### Redis Configuration

```bash
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=<strong-random-password>  # Required for production
```

### Blockchain Configuration

```bash
PRIVATE_KEY=<ethereum-private-key>
ALCHEMY_KEY=<alchemy-api-key>
ALCHEMY_URL=<alchemy-rpc-url>
```

### Data Provider API Keys

```bash
MONIERATE_API_KEY=<api-key>
EXCHANGERATE_API_KEY=<api-key>
TWELVEDATA_API_KEY=<api-key>
FIXER_API_KEY=<api-key>
CURRENCYLAYER_API_KEY=<api-key>
MORALIS_API_KEY=<api-key>
```

---

## GitHub Environments Setup

For proper deployment workflow, set up these environments in GitHub:

### 1. Staging Environment

1. Go to **Settings** → **Environments**
2. Click **New environment**
3. Name: `staging`
4. Configure:
   - **Environment URL**: `https://staging.api.ifalabs.com`
   - **Deployment branches**: `main` branch only

### 2. Production Environment

1. Go to **Settings** → **Environments**
2. Click **New environment**
3. Name: `production`
4. Configure:
   - **Environment URL**: `https://api.ifalabs.com`
   - **Deployment branches**: Tags only (pattern: `v*`)
   - **Required reviewers**: Add team members who must approve production deployments
   - **Wait timer**: Optional delay before deployment (e.g., 5 minutes)

---

## SSH Key Setup

### Generate Deployment Keys

```bash
# Generate a dedicated deployment key (do NOT use your personal key)
ssh-keygen -t ed25519 -C "oracle-engine-deploy" -f ~/.ssh/oracle_deploy

# The private key (~/.ssh/oracle_deploy) goes in GitHub Secrets
# The public key (~/.ssh/oracle_deploy.pub) goes on the server
```

### Add Public Key to Server

```bash
# On your deployment server
echo "ssh-ed25519 AAAA... oracle-engine-deploy" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### Add Private Key to GitHub

1. Copy the contents of `~/.ssh/oracle_deploy` (the private key)
2. Add it as `STAGING_SSH_KEY` or `PRODUCTION_SSH_KEY` in GitHub Secrets
3. Include the full key including `-----BEGIN OPENSSH PRIVATE KEY-----` and `-----END OPENSSH PRIVATE KEY-----`

---

## Security Best Practices

1. **Use dedicated deployment keys** - Don't use personal SSH keys
2. **Rotate secrets regularly** - Update API keys and passwords periodically
3. **Limit key permissions** - Use read-only access where possible
4. **Use environment protection rules** - Require approvals for production
5. **Audit secret access** - Review who has access to repository secrets
6. **Never commit secrets** - Use `.env.example` as a template, never `.env`

---

## Troubleshooting

### SSH Connection Failed

```
Error: Cannot connect to server via SSH
```

**Solutions:**
- Verify the SSH key is correctly formatted in GitHub Secrets
- Check that the public key is in `~/.ssh/authorized_keys` on the server
- Ensure the server's firewall allows SSH connections
- Verify the username and hostname are correct

### Docker Push Failed

```
Error: denied: permission_denied
```

**Solutions:**
- Ensure the repository has package write permissions enabled
- Check that the workflow has `packages: write` permission
- Verify the image name matches the repository

### Deployment Timeout

```
Error: Deployment timed out
```

**Solutions:**
- Check server resources (CPU, memory, disk)
- Review Docker logs: `docker compose logs`
- Verify health check endpoint is responding
