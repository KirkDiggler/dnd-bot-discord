# Branch Protection Setup

This document outlines the recommended branch protection rules for the D&D Discord Bot repository.

## Main Branch Protection

### Settings to Enable

1. **Go to**: Settings → Branches → Add Rule
2. **Branch name pattern**: `main`

### Protection Rules

#### ✅ Require a pull request before merging
- [x] Require approvals: **1**
- [x] Dismiss stale pull request approvals when new commits are pushed
- [x] Require review from CODEOWNERS (if using)

#### ✅ Require status checks to pass before merging
- [x] Require branches to be up to date before merging

**Required status checks:**
- `CI Success` (from ci.yml workflow)
- `Lint`
- `Test`
- `Build`
- `Security Scan`

#### ✅ Require conversation resolution before merging
- Ensures all PR comments are addressed

#### ✅ Require signed commits (optional but recommended)
- Ensures commit authenticity

#### ✅ Include administrators
- Even admins must follow the rules

#### ✅ Restrict who can push to matching branches (optional)
- Limit to specific users/teams if needed

### Additional Settings

#### ❌ Do NOT enable
- Allow force pushes (destroys history)
- Allow deletions (prevents accidental deletion)

## Develop Branch Protection

Similar to main but with relaxed rules:

1. **Branch name pattern**: `develop`
2. **Require pull request**: Yes
3. **Required approvals**: 0 (optional review)
4. **Required status checks**: Same as main
5. **Allow force pushes**: Only by admins

## Feature Branch Naming Convention

Encourage consistent branch naming:

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions/improvements
- `chore/description` - Maintenance tasks

## PR Merge Strategy

### Recommended: Squash and merge
- Keeps main branch history clean
- One commit per feature
- Preserves PR reference

### Alternative: Rebase and merge
- For carefully crafted commit histories
- When commits tell a story

### Avoid: Create a merge commit
- Creates messy history
- Hard to revert features

## Automated PR Rules

The repository includes GitHub Actions that:

1. **Label PRs by size** (XS, S, M, L, XL)
2. **Run comprehensive CI checks**
3. **Provide test coverage reports**
4. **Security scanning**
5. **Build verification for multiple platforms**

## Setting Up Required Checks

1. Make sure CI workflow runs successfully at least once
2. Go to Settings → Branches → Edit rule
3. Search for and select these status checks:
   - `quick-checks`
   - `lint` 
   - `test`
   - `build`
   - `security`
   - `ci-success`

## Review Guidelines

### What CI Checks

✅ **Automatically verified:**
- Code compiles
- Tests pass (unit & integration)
- Code is formatted
- No linting errors
- No security vulnerabilities
- go.mod is tidy

### What Humans Should Review

👀 **Manual review focus:**
- Business logic correctness
- Code readability and maintainability  
- Performance implications
- Security considerations beyond automated scans
- Architecture decisions
- API design
- Error handling
- Test quality and coverage

## Emergency Procedures

### If you need to bypass protections:

1. **Never do this on main branch**
2. For develop branch emergencies:
   - Document why in PR description
   - Get post-merge review
   - Fix any issues immediately

### If CI is broken:

1. Create a PR to fix CI first
2. Admin can temporarily disable specific checks
3. Re-enable immediately after fix is merged

## Monitoring

- Review weekly:
  - Failed PR checks patterns
  - Time to merge metrics
  - Which checks fail most often
  
- Adjust rules if:
  - Checks are too restrictive
  - False positives are common
  - Development velocity is impacted

## CODEOWNERS (Optional)

Create `.github/CODEOWNERS` file:

```
# Global owners
* @username

# Specific areas
/internal/entities/ @domain-expert
/internal/handlers/discord/ @discord-expert
/internal/services/ @services-team
```

This ensures relevant people are automatically requested for review.