---
name: setup-go-module
description: Scaffold a Go backend with module init, cmd/ entrypoint, and internal/ package structure. Use when bootstrapping a new Go project or when the bootstrap-project orchestrator invokes this skill. Trigger on "set up Go", "init Go module", "create Go backend", or as part of project bootstrapping.
---

# Setup Go Module

Setting up Go module with cmd/ and internal/ structure.

## Prerequisites

- Go 1.22 or later installed
- Empty or new project directory

## Process

### Step 1: Determine Module Path

If not provided, ask for the module path (e.g., `github.com/petrepopescu21/myproject`).

```bash
MODULE_PATH="$1"
if [ -z "$MODULE_PATH" ]; then
  read -p "Enter module path (e.g., github.com/petrepopescu21/myproject): " MODULE_PATH
fi
PROJECT_NAME=$(basename "$MODULE_PATH")
```

### Step 2: Initialize Go Module

```bash
go mod init "$MODULE_PATH"
```

### Step 3: Create Directory Structure

Create the following directories:

```bash
mkdir -p cmd/"$PROJECT_NAME"
mkdir -p internal/api
mkdir -p internal/domain
mkdir -p internal/store
```

### Step 4: Write Main Entrypoint

Create `cmd/$PROJECT_NAME/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/petrepopescu21/myproject/internal/api"
)

func main() {
	// Create a router with all endpoints
	router := api.NewRouter()

	// Create HTTP server with required timeouts
	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("starting server on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown with timeout
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v\n", err)
	}

	log.Println("server stopped")
}
```

### Step 5: Write API Router

Create `internal/api/router.go`:

```go
package api

import "net/http"

// NewRouter returns a configured HTTP router with all API endpoints.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /api/v1/health", handleHealth)

	return mux
}

// handleHealth returns a simple health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
```

### Step 6: Verify Build and Lint

Run verification commands:

```bash
go build ./...
go vet ./...
```

Expected output:
- No build errors
- No vet warnings

### Step 7: Commit

Commit the scaffold:

```bash
git add -A
git commit -m "scaffold: initialize Go module with cmd/ and internal/ structure"
```

## Summary

The Go module is now set up with:

- **cmd/$PROJECT_NAME/main.go** — HTTP server entrypoint with graceful shutdown and signal handling
- **internal/api/router.go** — Router factory with `/api/v1/health` endpoint
- **internal/domain/** — Business entity package (ready for models)
- **internal/store/** — Repository package (ready for database implementations)

All code uses the standard library only. The server listens on `:8080` and responds to health checks at `/api/v1/health`.
