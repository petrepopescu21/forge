---
name: setup-linting
description: Configure golangci-lint for Go and ESLint (flat config) for TypeScript/React with opinionated rulesets. Use when bootstrapping a project or adding linting to an existing one. Trigger on "set up linting", "add linters", "configure ESLint", "configure golangci-lint", or as part of project bootstrapping.
---

# Setup Linting

Announce: "Setting up golangci-lint and ESLint with opinionated configs."

## Process

### 1. Configure golangci-lint

Create `.golangci.yml`:

```yaml
version: "2"

go: "1.24"

linters:
  default: none
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - ineffassign
    - gocritic
    - revive
    - tparallel
    - paralleltest

issues:
  exclude-dirs:
    - web/node_modules
    - test/e2e

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
    disabled-checks:
      - hugeParam

  revive:
    enable-all-rules: false
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: defer
      - name: dot-imports
      - name: early-return
      - name: error-naming
      - name: error-return
      - name: error-strings
      - name: errorf
      - name: exported
      - name: if-return
      - name: increment-decrement
      - name: indent-error-flow
      - name: interface-naming
      - name: package-comments
      - name: range
      - name: receiver-naming
      - name: redefines-builtin-id
      - name: superfluous-else
      - name: time-naming
      - name: unexported-return
      - name: unhandled-error
      - name: unnecessary-stmt
      - name: var-declaration
      - name: var-naming
      - name: void-in-defer

output:
  formats:
    - format: colored-line-number
  sort-results: true
  sort-order:
    - linter
    - severity
    - file
```

### 2. Configure ESLint

Install dependencies:

```bash
cd web
bun install --save-dev \
  eslint \
  @eslint/js \
  typescript-eslint \
  eslint-plugin-react-hooks \
  eslint-plugin-react-refresh \
  globals
```

Create `web/eslint.config.js`:

```javascript
import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';

export default tseslint.config(
  { ignores: ['dist'] },
  {
    extends: [
      js.configs.recommended,
      ...tseslint.configs.recommended,
    ],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
    },
  },
);
```

### 3. Verify Configuration

Run golangci-lint:

```bash
golangci-lint run ./...
```

Run ESLint:

```bash
cd web
bun run lint
```

Both commands should complete without errors.

### 4. Commit

Stage and commit the configuration files:

```bash
git add .golangci.yml web/eslint.config.js web/package.json
git commit -m "Configure golangci-lint and ESLint with opinionated rulesets"
```
