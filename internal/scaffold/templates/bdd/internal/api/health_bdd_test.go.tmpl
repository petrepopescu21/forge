package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cucumber/godog"
)

type apiContext struct {
	server   *httptest.Server
	response *http.Response
	body     []byte
	err      error
}

func (ac *apiContext) theAPIServerIsRunning() error {
	// In a real application, you would initialize your actual API router here
	// For now, we'll create a minimal health check handler
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
	})

	ac.server = httptest.NewServer(mux)
	return nil
}

func (ac *apiContext) iMakeAGETRequestToEndpoint(endpoint string) error {
	url := ac.server.URL + endpoint
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("making GET request: %w", err)
	}

	ac.response = resp
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	defer resp.Body.Close()
	ac.body = body
	return nil
}

func (ac *apiContext) theResponseStatusShouldBe(status int) error {
	if ac.response.StatusCode != status {
		return fmt.Errorf("expected status %d, got %d", status, ac.response.StatusCode)
	}
	return nil
}

func (ac *apiContext) theResponseBodyShouldContain(expected string) error {
	if !bytes.Contains(ac.body, []byte(expected)) {
		return fmt.Errorf("expected response to contain %q, got %s", expected, string(ac.body))
	}
	return nil
}

func (ac *apiContext) closeServer() {
	if ac.server != nil {
		ac.server.Close()
	}
}

// InitializeScenario is called before each scenario to reset context
func (ac *apiContext) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Given(`the API server is running`, ac.theAPIServerIsRunning)
	ctx.When(`I make a GET request to "([^"]*)"`, ac.iMakeAGETRequestToEndpoint)
	ctx.Then(`the response status should be (\d+)`, ac.theResponseStatusShouldBe)
	ctx.Then(`the response body should contain "([^"]*)"`, ac.theResponseBodyShouldContain)
	ctx.After(func(ctx context.Context, scenario *godog.Scenario, err error) context.Context {
		ac.closeServer()
		return ctx
	})
}

// TestHealthFeatures runs all feature tests for health checks
func TestHealthFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ac := &apiContext{}
			ac.InitializeScenario(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog scenarios failed")
	}
}
