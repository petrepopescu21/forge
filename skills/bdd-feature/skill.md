---
name: bdd-feature
description: Transform a rough feature prompt into expert Gherkin scenarios and generate step definition stubs. Use when starting any new feature, adding behavior, or when the user provides a feature description (even vague ones). Trigger on phrases like "add feature", "I want users to be able to", "new capability", "implement X", or any description of desired behavior. This skill acts as a BDD analyst — it proposes concrete scenarios rather than asking clarifying questions.
---

# BDD Feature

Transforms rough prompts into expert Gherkin scenarios with step definition stubs.

**Announce:** "Using BDD feature skill — I'll propose Gherkin scenarios for your approval before any implementation."

## Why This Matters

Starting with Gherkin scenarios forces you to define behavior precisely before writing code. It catches ambiguity early, documents acceptance criteria, and produces tests that verify the feature works end-to-end. The scenarios become the single source of truth for what the feature does.

## Process

### Step 1: Analyze the Prompt

Read the user's prompt and extract:
- **Actor(s):** Who is performing the action? (e.g., rep, manager, admin, system)
- **Action(s):** What are they trying to do?
- **Outcome(s):** What should happen when they succeed?
- **Edge cases:** What can go wrong? (unauthorized, not found, invalid state, conflicts)

Also read:
- `CLAUDE.md` for domain context, roles, and existing conventions
- Existing `.feature` files for style consistency and to avoid duplicating scenarios

### Step 2: Propose Scenarios

Do NOT ask clarifying questions. Instead, propose what the feature should look like based on domain context. Present a complete `.feature` file for approval.

**The analyst behavior:** If the prompt is vague (e.g., "add assignments"), propose a fully-formed feature with your best interpretation of what it should do. The user will correct what's wrong — this is faster than a Q&A loop.

Structure the feature file like this:

```gherkin
Feature: Lead Assignment
  As a manager
  I want to assign leads to reps
  So that each lead has a responsible rep for follow-up

  Background:
    Given the system has the following reps:
      | name       | role    |
      | Alice      | rep     |
      | Bob        | rep     |
    And the following unassigned leads exist:
      | company    | status     |
      | Acme Corp  | new        |

  Scenario: Manager assigns a lead to a rep
    When the manager assigns "Acme Corp" to "Alice"
    Then the lead "Acme Corp" should be assigned to "Alice"
    And a "lead_assigned" event should be recorded

  Scenario: Cannot assign to a non-existent rep
    When the manager assigns "Acme Corp" to "Unknown"
    Then the assignment should fail with "rep not found"

  Scenario: Rep cannot assign leads
    Given the current user is a rep
    When the rep tries to assign "Acme Corp" to "Bob"
    Then the assignment should fail with "forbidden"

  Scenario Outline: Assignment updates lead status
    Given a lead with status "<initial_status>"
    When the manager assigns the lead to a rep
    Then the lead status should be "<final_status>"

    Examples:
      | initial_status | final_status |
      | new            | assigned     |
      | contacted      | assigned     |
```

**Include:**
- Happy path scenario(s)
- Error/edge cases (unauthorized, not found, invalid state, conflicts)
- Scenario Outlines for parameterized cases where 3+ scenarios differ only by data
- Background section for shared setup (when 2+ scenarios share Given steps)

**Feature file conventions:**
- Use domain language from `CLAUDE.md` (e.g., "rep", "manager", "lead", not "user", "item")
- Steps should be readable by a non-developer
- Avoid implementation details in steps (no "sends a POST request", no "queries the database")
- One Feature per file, focused on one capability

### Step 3: Get Approval

Present the proposed `.feature` file and ask:

> "Here are the scenarios I've drafted for this feature. Do these cover the right behaviors? Want me to add, remove, or change any scenarios?"

Wait for approval. Adjust if the user provides feedback.

### Step 4: Write the Feature File

After approval, write the `.feature` file to the correct location:
- **Backend behavior** → `features/<feature-name>.feature`
- **Frontend behavior** → `web/features/<feature-name>.feature`

Use kebab-case for filenames (e.g., `lead-assignment.feature`).

### Step 5: Generate Step Definition Stubs

#### Go (godog)

Create a test file with step definitions wired to the feature:

```go
// internal/leads/assignment_test.go
package leads_test

import (
    "testing"
    "github.com/cucumber/godog"
)

func TestAssignmentFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeAssignmentScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../features/lead-assignment.feature"},
            TestingT: t,
        },
    }
    if suite.Run() != 0 {
        t.Fatal("non-zero exit from godog")
    }
}

func InitializeAssignmentScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^the system has the following reps:$`, theSystemHasTheFollowingReps)
    ctx.Step(`^the manager assigns "([^"]*)" to "([^"]*)"$`, theManagerAssigns)
    ctx.Step(`^the lead "([^"]*)" should be assigned to "([^"]*)"$`, theLeadShouldBeAssigned)
    // ... remaining steps
}

func theSystemHasTheFollowingReps(table *godog.Table) error {
    return godog.ErrPending
}

func theManagerAssigns(lead, rep string) error {
    return godog.ErrPending
}

func theLeadShouldBeAssigned(lead, rep string) error {
    return godog.ErrPending
}
```

#### Frontend (playwright-bdd)

Create step definitions for playwright-bdd:

```typescript
// web/e2e/steps/lead-assignment.steps.ts
import { createBdd } from 'playwright-bdd'
import { test } from '../fixtures'

const { Given, When, Then } = createBdd(test)

Given('the system has the following reps:', async ({ page }, table) => {
  // TODO: implement — mock API or seed data
})

When('the manager assigns {string} to {string}', async ({ page }, lead: string, rep: string) => {
  // TODO: implement
})

Then('the lead {string} should be assigned to {string}', async ({ page }, lead: string, rep: string) => {
  // TODO: implement
})
```

### Step 6: Verify Stubs Fail

Run the tests to confirm the stubs are wired correctly and fail with "pending" status:

```bash
# Go
go test ./internal/leads/... -run TestAssignmentFeatures -v

# Frontend
cd web && bun run test:e2e
```

The tests should fail because steps return `ErrPending` or have `TODO` implementations. This confirms the BDD scaffolding is wired correctly and ready for TDD implementation.

## After This Skill

The next step is `forge:tdd-cycle` to implement each step definition using red/green/refactor. The step stubs are the failing tests — implement them one at a time.
