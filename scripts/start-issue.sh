#!/bin/bash
# Helper script to start working on an issue with proper workflow

if [ "$#" -ne 1 ]; then
    echo "Usage: ./scripts/start-issue.sh <issue-number>"
    echo "Example: ./scripts/start-issue.sh 57"
    exit 1
fi

ISSUE_NUMBER=$1

# Get issue title for branch name
ISSUE_TITLE=$(gh issue view $ISSUE_NUMBER --json title -q .title)
if [ -z "$ISSUE_TITLE" ]; then
    echo "Error: Could not find issue #$ISSUE_NUMBER"
    exit 1
fi

# Convert title to branch-friendly name
BRANCH_NAME=$(echo "$ISSUE_TITLE" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')
BRANCH_NAME="fix-${ISSUE_NUMBER}-${BRANCH_NAME:0:40}"

echo "📋 Working on issue #$ISSUE_NUMBER: $ISSUE_TITLE"

# Add to project if not already added
echo "📊 Adding to project board..."
gh issue edit $ISSUE_NUMBER --add-project "Co Op Dungeon Adventure Start" 2>/dev/null || echo "   (already in project)"

# Assign to self
echo "👤 Assigning to self..."
gh issue edit $ISSUE_NUMBER --assignee @me

# Create and checkout branch
echo "🌿 Creating branch: $BRANCH_NAME"
git checkout main
git pull origin main
git checkout -b "$BRANCH_NAME"

echo ""
echo "✅ Ready to work on issue #$ISSUE_NUMBER!"
echo ""
echo "📝 Remember when creating PR:"
echo "   - Include 'Fixes #$ISSUE_NUMBER' in the PR description"
echo "   - Run 'make pre-commit' before committing"
echo ""
echo "🚀 Quick PR command:"
echo "   gh pr create --title \"<title>\" --body \"Fixes #$ISSUE_NUMBER\""