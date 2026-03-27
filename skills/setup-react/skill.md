---
name: setup-react
description: Scaffold a React + TypeScript frontend with Vite, Bun, TanStack Query/Table, Vitest, and React Testing Library. Use when bootstrapping a frontend or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up frontend", "create React app", "add frontend", or as part of project bootstrapping.
---

# Setup React

Setting up React + TypeScript frontend with Vite and Bun.

## Prerequisites

- Bun 1.1 or higher

## Process

### 1. Create Vite Project with React Template

```bash
bun create vite web --template react-ts
cd web
bun install
```

### 2. Install Dependencies

Install runtime dependencies:

```bash
bun add @tanstack/react-query @tanstack/react-table
```

Install development dependencies:

```bash
bun add -D @testing-library/react @testing-library/jest-dom @testing-library/user-event @vitejs/plugin-react vitest jsdom @vitest/coverage-v8
```

### 3. Configure Vite

Create or update `vite.config.ts`:

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

### 4. Configure Vitest

Create `vitest.config.ts`:

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    exclude: ['e2e', 'node_modules'],
    passWithNoTests: true,
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

### 5. Create Test Setup File

Create `src/test/setup.ts`:

```typescript
import '@testing-library/jest-dom'
```

### 6. Configure TypeScript

Update `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "resolveJsonModule": true,
    "jsx": "react-jsx",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

### 7. Add Package.json Scripts

Update the `scripts` section in `package.json`:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc --noEmit && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:e2e": "vitest run --include '**/*.e2e.ts'",
    "test:e2e:headed": "vitest run --include '**/*.e2e.ts' --reporter=verbose",
    "test:e2e:integration": "vitest run --include '**/*.integration.ts'",
    "lint": "eslint src --ext .ts,.tsx",
    "typecheck": "tsc --noEmit"
  }
}
```

### 8. Verify Setup

Run the verification commands:

```bash
bun run build
bun run test
bun run typecheck
```

All commands must succeed before proceeding.

### 9. Commit

```bash
git add -A
git commit -m "feat: scaffold React + TypeScript frontend with Vite, Bun, TanStack Query/Table, Vitest"
```
