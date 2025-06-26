#!/bin/bash
# check-commit.sh - Comprehensive pre-commit checker for dnd-bot-discord
# 
# This script runs all the checks that should pass before committing code.
# It can be run manually or called from git hooks.
#
# Usage:
#   ./scripts/check-commit.sh        # Run all checks
#   ./scripts/check-commit.sh --fix  # Auto-fix what we can

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse arguments
AUTO_FIX=false
if [[ "$1" == "--fix" ]]; then
    AUTO_FIX=true
fi

echo -e "${GREEN}🔍 D&D Bot Pre-Commit Checker${NC}"
echo "================================"

# Track if we fixed anything
FIXED_SOMETHING=false

# 1. Code Formatting
echo -e "\n${YELLOW}📝 Checking code formatting...${NC}"
if ! gofmt -l . | grep -q .; then
    echo -e "${GREEN}✓ Code formatting looks good${NC}"
else
    if $AUTO_FIX; then
        echo "  Auto-formatting code..."
        make fmt
        FIXED_SOMETHING=true
        echo -e "${GREEN}✓ Code formatted${NC}"
    else
        echo -e "${RED}✗ Code needs formatting${NC}"
        echo "  Run: make fmt"
        exit 1
    fi
fi

# 2. Module Tidiness
echo -e "\n${YELLOW}📦 Checking module dependencies...${NC}"
cp go.mod go.mod.backup
cp go.sum go.sum.backup
go mod tidy
if ! diff -q go.mod go.mod.backup > /dev/null || ! diff -q go.sum go.sum.backup > /dev/null; then
    if $AUTO_FIX; then
        rm go.mod.backup go.sum.backup
        FIXED_SOMETHING=true
        echo -e "${GREEN}✓ Modules tidied${NC}"
    else
        mv go.mod.backup go.mod
        mv go.sum.backup go.sum
        echo -e "${RED}✗ Modules need tidying${NC}"
        echo "  Run: go mod tidy"
        exit 1
    fi
else
    rm go.mod.backup go.sum.backup
    echo -e "${GREEN}✓ Modules are tidy${NC}"
fi

# 3. Run Tests
echo -e "\n${YELLOW}🧪 Running tests...${NC}"
if go test ./... -short -timeout 30s > /tmp/test-output.txt 2>&1; then
    echo -e "${GREEN}✓ All tests passed${NC}"
else
    echo -e "${RED}✗ Tests failed${NC}"
    cat /tmp/test-output.txt
    exit 1
fi

# 4. Linting (if available)
echo -e "\n${YELLOW}🔎 Running linter...${NC}"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./... > /tmp/lint-output.txt 2>&1; then
        echo -e "${GREEN}✓ No lint issues${NC}"
    else
        echo -e "${YELLOW}⚠ Lint issues found (not blocking)${NC}"
        # Show first 20 lines of lint output
        head -20 /tmp/lint-output.txt
        echo "  ... (run 'golangci-lint run ./...' to see all)"
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not found, skipping${NC}"
fi

# 5. Check for common issues
echo -e "\n${YELLOW}🔍 Checking for common issues...${NC}"

# Check for debugging prints
if grep -r "fmt.Println" --include="*.go" . | grep -v "_test.go" | grep -v "cmd/"; then
    echo -e "${YELLOW}⚠ Found fmt.Println (consider using proper logging)${NC}"
fi

# Check for TODO/FIXME
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" . | wc -l)
if [ $TODO_COUNT -gt 0 ]; then
    echo -e "${YELLOW}ℹ Found $TODO_COUNT TODO/FIXME comments${NC}"
fi

# Summary
echo -e "\n================================"
if $FIXED_SOMETHING; then
    echo -e "${GREEN}✅ Auto-fixed some issues!${NC}"
    echo -e "Please review the changes and stage them:"
    echo -e "  ${YELLOW}git add -A${NC}"
    echo -e "  ${YELLOW}git commit${NC}"
else
    echo -e "${GREEN}✅ All checks passed!${NC}"
    echo -e "Your code is ready to commit! 🎉"
fi