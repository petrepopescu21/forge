---
name: add-feature
description: Orchestrate the full feature development cycle — from rough prompt through BDD scenarios, TDD implementation, to quality verification. Use when adding any new feature, capability, or behavior. This is the primary development workflow enforced by CLAUDE.md. Trigger on any feature request, behavior description, or when the user says "add", "implement", "build", "create" followed by a feature description. Even for "simple" features, use this workflow.
---

# Add Feature

Using the add-feature workflow — I'll start with BDD scenarios, then implement with TDD.

## The Workflow

This orchestrator drives a complete feature development cycle through four sequential phases:

```
graph TD
    A["Receive Feature Prompt"] --> B["Phase 1: BDD"]
    B --> B1["forge:bdd-feature"]
    B1 --> B2[".feature file + step stubs"]
    B2 --> B3{"Scenarios Approved?"}
    B3 -->|Refine| B1
    B3 -->|Proceed| C["Phase 2: TDD"]
    C --> C1["forge:tdd-cycle"]
    C1 --> C2["Red/Green/Refactor Steps"]
    C2 --> C3{"All Steps Done?"}
    C3 -->|More steps| C1
    C3 -->|Proceed| D["Phase 3: Quality"]
    D --> D1["make lint && make typecheck && make test"]
    D1 --> D2["All Gates Pass?"]
    D2 --> D3{"All Pass?"}
    D3 -->|Fix issues| D1
    D3 -->|Proceed| E["Phase 4: Summary"]
    E --> F["Complete"]

    style A fill:#e1f5ff
    style B fill:#fff3e0
    style C fill:#f3e5f5
    style D fill:#e8f5e9
    style E fill:#fce4ec
    style F fill:#c8e6c9
```

## Phase 1: BDD — Capture Scenarios

**Goal:** Write executable specifications before any code.

- Invoke `forge:bdd-feature` with your feature prompt
- Outputs: `.feature` file (Gherkin) and step definition stubs
- **Do NOT write implementation code yet**
- Review scenarios with user; refine if needed
- Proceed to Phase 2 only when scenarios are approved

## Phase 2: TDD — Red/Green/Refactor

**Goal:** Implement each BDD step with test-driven discipline.

- Invoke `forge:tdd-cycle` for each step stub from Phase 1
- Process per step:
  1. **Red:** Write failing test from the step
  2. **Green:** Implement minimal code to pass
  3. **Refactor:** Clean up, extract, improve
  4. **Commit:** One commit per completed step
- Continue until all steps are passing
- All tests must pass before proceeding to Phase 3

## Phase 3: Quality — Verification Gates

**Goal:** Ensure code meets all quality standards.

- Run all quality gates directly:
  - `make test` (unit + integration tests)
  - `make lint` (style + best practices)
  - `make typecheck` (TypeScript strict mode)
  - `make coverage` (if applicable)
- If any gate fails, fix and re-run; do not skip
- All gates must pass before Phase 4

## Phase 4: Summary

**Goal:** Document what was delivered.

Provide a concise summary with:
- **Feature name:** What capability was added?
- **Scenarios implemented:** List the `.feature` scenarios
- **Tests added:** Count of new test cases (unit/integration)
- **Files changed:** Modified files and new files created
- **Design decisions:** Key choices made during implementation

## Important Constraints

- **Do NOT skip BDD.** Always write scenarios first.
- **Do NOT write implementation before scenarios are approved.** BDD unblocks TDD.
- **Do NOT skip quality check.** Every gate must pass.
- **Each step gets its own commit.** No squashing during development; clear history matters.
- **Pass context between phases.** Each phase builds on outputs from the previous one.

## When to Use This Skill

Trigger this orchestrator whenever you:
- Request a new feature or capability
- Ask to "add", "implement", "build", or "create" followed by a description
- Describe desired behavior or outcomes
- Need to extend existing functionality

Even for features that seem simple or small, this workflow ensures quality and traceability.
