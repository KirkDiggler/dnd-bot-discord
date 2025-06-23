#!/bin/bash
# Helper script to create PRs with proper formatting

if [ "$#" -lt 2 ]; then
    echo "Usage: ./scripts/create-pr.sh <issue-number> <pr-title>"
    echo "Example: ./scripts/create-pr.sh 57 'Refactor attack logic to service layer'"
    exit 1
fi

ISSUE_NUMBER=$1
PR_TITLE=$2

# Get issue details
ISSUE_DETAILS=$(gh issue view $ISSUE_NUMBER --json title,body,labels)
if [ -z "$ISSUE_DETAILS" ]; then
    echo "Error: Could not find issue #$ISSUE_NUMBER"
    exit 1
fi

# Create PR body
PR_BODY=$(cat <<EOF
## Summary
$PR_TITLE

## Changes
- TODO: List main changes here

## Test Plan
- [ ] Unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing completed
- [ ] No regressions introduced

Fixes #$ISSUE_NUMBER

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)

echo "📝 Creating PR for issue #$ISSUE_NUMBER..."
echo ""

# Create the PR
gh pr create --title "$PR_TITLE" --body "$PR_BODY" --draft

echo ""
echo "✅ Draft PR created!"
echo "📝 Remember to:"
echo "   - Update the Changes section"
echo "   - Complete the test plan"
echo "   - Mark PR as ready when done"