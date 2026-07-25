# Contributing to NOANT

## Branching Strategy

```
main          ← production-ready, always deployable
  └── feat/*  ← new features
  └── fix/*   ← bug fixes
  └── hotfix/*← emergency production fixes
```

### Rules

1. **Never commit directly to `main`** — always use a branch + merge
2. **`main` must always pass CI** — lint, build, tests
3. **Merge via fast-forward or squash** — keep history clean

### Workflow

```bash
# Start work
git checkout main && git pull origin main
git checkout -b feat/whatsapp-webhook-v2

# Work, commit with conventional commits
git add .
git commit -m "feat: add webhook retry logic"

# Push and create PR
git push -u origin feat/whatsapp-webhook-v2
gh pr create --title "feat: WhatsApp webhook v2" --body "Closes #42"

# After review/CI passes, merge
gh pr merge --squash

# Clean up
git checkout main && git pull origin main
git branch -d feat/whatsapp-webhook-v2
```

### Branch Naming

| Prefix     | Use for                        | Example                    |
|------------|--------------------------------|----------------------------|
| `feat/`    | New features                   | `feat/campaign-analytics`  |
| `fix/`     | Bug fixes                      | `fix/websocket-mem-leak`   |
| `hotfix/`  | Emergency production fixes     | `hotfix/auth-crash`        |
| `chore/`   | Tooling, deps, CI              | `chore/upgrade-golang`     |
| `docs/`    | Documentation only             | `docs/api-reference`       |

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
[optional footer]
```

### Types

| Type     | When                                    | Bumps     |
|----------|-----------------------------------------|-----------|
| `feat`   | New feature                             | minor     |
| `fix`    | Bug fix                                 | patch     |
| `docs`   | Documentation only                      | —         |
| `chore`  | Build, CI, deps, tooling                | —         |
| `refactor` | Code restructure (no behavior change) | —         |
| `perf`   | Performance improvement                 | patch     |
| `test`   | Adding/fixing tests                     | —         |

### Examples

```
feat(chat): add message search endpoint
fix(auth): prevent token refresh race condition
docs: update API reference for /leads
chore: upgrade golangci-lint to v2.12
```

## Releases & Tags

Tags follow [Semantic Versioning](https://semver.org/): `v<major>.<minor>.<patch>`

```bash
# Tag a release (after merging to main)
git tag -a v1.0.0 -m "Release v1.0.0: WhatsApp integration + billing"
git push origin v1.0.0
git push enterprise v1.0.0
```

| Version bump | When                                         |
|-------------|----------------------------------------------|
| `major`     | Breaking API changes, DB migrations that break backward compat |
| `minor`     | New features, non-breaking improvements      |
| `patch`     | Bug fixes, security patches                  |

### Rollback

```bash
# See all tags
git tag -l

# Roll back to a known-good version
git checkout v0.9.0
# or reset main (destructive — use with caution)
git reset --hard v0.9.0 && git push --force origin main
```

## Git Hooks

Pre-commit hooks run automatically:

- **pre-commit**: Runs `golangci-lint` on staged Go files
- **commit-msg**: Validates conventional commit format

To install hooks after cloning:

```bash
git config core.hooksPath .githooks
```

## Protected Branches (Recommended)

When ready, enable on GitHub:

- **`main`**: Require PR reviews, require CI to pass, no force-push
- **Tags**: Protect `v*` tags from deletion
