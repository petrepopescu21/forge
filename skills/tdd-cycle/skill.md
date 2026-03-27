---
name: tdd-cycle
description: Drive test-driven development with strict red/green/refactor discipline. Use when implementing any feature, fixing bugs, or writing new code. Trigger on phrases like "implement this", "write the code", "build this feature", "fix this bug", or when transitioning from BDD scenarios to implementation. This skill enforces that failing tests exist before implementation code is written.
---

# TDD Cycle

Drives implementation through strict red/green/refactor discipline.

**Announce:** "Using TDD cycle — I'll write failing tests first, then implement."

## The Iron Law

**No implementation code without a failing test first.**

This is not optional. Every piece of functionality gets a test that fails before the implementation exists, then passes after. If you find yourself writing implementation code without a red test, stop and write the test first.

## Process

### Starting Point

The TDD cycle can start from two places:

1. **From BDD step stubs** — the `bdd-feature` skill has generated `.feature` files and step definition stubs. Each unimplemented step is a failing test. Pick the first one.
2. **Standalone** — no BDD context. Identify the first unit of behavior to implement and write a test for it.

### The Loop

For each unit of behavior:

#### 1. Red — Write the Failing Test

Write a test that describes what the code should do. The test must fail because the implementation doesn't exist yet.

**Go pattern:**
```go
func TestLeadAssignment_AssignsToRep(t *testing.T) {
    t.Parallel()
    svc := NewLeadService(stubRepo{}, stubRBAC{})
    err := svc.Assign(ctx, leadID, repID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**React pattern:**
```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

it('assigns lead when rep is selected', async () => {
  render(<AssignmentDialog leadId="123" />)
  await userEvent.click(screen.getByRole('button', { name: 'Assign' }))
  expect(screen.getByText('Assigned')).toBeInTheDocument()
})
```

#### 2. Run — Verify It Fails

```bash
# Go
go test ./internal/leads/... -run TestLeadAssignment_AssignsToRep -v

# Frontend
cd web && bun run test -- --reporter=verbose AssignmentDialog
```

Confirm the test fails for the expected reason (missing function, wrong return value, etc.) — not for an unrelated error like a syntax mistake in the test itself.

#### 3. Green — Minimum Implementation

Write the smallest amount of code that makes the test pass. Do not add error handling, validation, or features that the test doesn't require. If you're tempted to add something "while you're here," resist — write a test for it first.

#### 4. Run — Verify It Passes

```bash
make test
```

All tests must pass — not just the new one. If an existing test breaks, fix the implementation, not the test (unless the test was wrong).

#### 5. Refactor — Clean Up

With green tests as your safety net, clean up:
- Remove duplication
- Improve naming
- Extract helpers if a pattern repeats 3+ times
- Simplify logic

Run `make test` again after refactoring to confirm nothing broke.

#### 6. Commit

```bash
git add -A
git commit -m "feat: <what this test+implementation achieves>"
```

Commit after each red→green→refactor cycle, not after batching multiple features.

#### 7. Repeat

Pick the next unimplemented behavior and go back to step 1.

### Completion

After all behaviors are implemented and all tests pass:

1. Run the full quality check: `make lint && make typecheck && make test`
2. If any gate fails, fix the issue and re-run
3. Summarize what was implemented and how many tests were added

## Test Design Guidelines

### Go Tests
- Use `_test` package suffix (e.g., `package leads_test`) to test the public API
- Use `t.Parallel()` on every test and subtest
- Use manual stubs over mocking frameworks — implement the interface with the behavior you need
- Use `t.Helper()` on test helper functions
- Wrap test setup in helper functions like `newTestService()` that return configured dependencies

### Frontend Tests
- Use React Testing Library — query by role, label, or text, never by CSS class or test ID
- Use `userEvent` (not `fireEvent`) for user interactions
- Test behavior, not implementation — don't assert on internal state
- Use `vi.fn()` for callback mocking
- Use `renderHook` from `@testing-library/react` for hook tests

## Red Flags

If you notice any of these, stop and correct course:

| Symptom | Problem |
|---------|---------|
| Writing implementation before a test | Violates the iron law — write the test first |
| Test passes immediately when written | The test isn't testing anything new, or the implementation already exists |
| Large implementation between test runs | Break it into smaller steps — each step should be one test |
| Fixing tests to match implementation | Tests define behavior — if the test is right, fix the implementation |
| Skipping refactor phase | Technical debt accumulates — take 30 seconds to clean up |
