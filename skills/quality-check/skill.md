---
name: quality-check
description: Run all quality gates (lint, typecheck, test) and report results. Use before every commit, after completing a feature, or when you need to verify the project is clean. Trigger when you see phrases like "check quality", "run checks", "is this clean", "before I commit", or after completing any implementation work.
---

# Quality Check

Pre-commit quality gate. Runs all checks and reports pass/fail for each.

**Announce:** "Running quality checks — lint, typecheck, and tests."

## Why This Exists

Catching lint errors, type errors, and test failures before they hit CI saves time and keeps the commit history clean. This skill runs all three gates in a predictable order and gives you a clear pass/fail report.

## Process

Run each gate in order. Do not stop on first failure — run all three and report everything.

### Step 1: Lint

```bash
make lint
```

This runs `golangci-lint run ./...` for Go and `cd web && bun run lint` for the frontend. Both must pass.

### Step 2: Typecheck

```bash
make typecheck
```

This runs `cd web && bun run typecheck` (tsc --noEmit). Must produce zero errors.

### Step 3: Test

```bash
make test
```

This runs `go test ./...` for Go and `cd web && bun run test` for frontend unit tests.

## Reporting

After all three gates have run, present a summary:

| Gate | Status |
|------|--------|
| Lint | PASS / FAIL |
| Typecheck | PASS / FAIL |
| Test | PASS / FAIL |

If any gate failed, show the relevant error output. Do not attempt to auto-fix — just report what needs attention.

## When NOT to Use

- Do not run this during active TDD red/green cycles (tests are expected to fail during red phase)
- Do not run E2E tests — this is for fast pre-commit checks only
