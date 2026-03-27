package tests

import (
	"testing"
	"time"
)

func TestSetupHelm(t *testing.T) {
	skipIfNoAPI(t)
	if testing.Short() {
		t.Skip("skipping infrastructure test in short mode")
	}

	dir := t.TempDir()

	runClaude(t, dir, "use forge:setup-go-module with project name testapp and module github.com/test/testapp")
	runClaude(t, dir, "use forge:setup-makefile with project name testapp")
	runClaude(t, dir, "use forge:setup-helm with project name testapp")

	// L2: file structure
	assertFileExists(t, dir, "deploy/helm/testapp/Chart.yaml")
	assertFileContains(t, dir, "deploy/helm/testapp/Chart.yaml", "name: testapp")
	runCmd(t, dir, "helm", "lint", "deploy/helm/testapp")

	// L3: Kind cluster + Helm deploy
	runMake(t, dir, "cluster-create")
	t.Cleanup(func() {
		runCmd(t, dir, "make", "cluster-delete")
	})

	runMake(t, dir, "build")
	runCmd(t, dir, "docker", "build", "-t", "testapp:latest", ".")
	runCmd(t, dir, "kind", "load", "docker-image", "testapp:latest", "--name", "testapp")

	runCmd(t, dir, "helm", "install", "testapp", "deploy/helm/testapp",
		"--set", "image.repository=testapp",
		"--set", "image.tag=latest",
		"--set", "image.pullPolicy=Never",
		"--wait", "--timeout", "120s")

	time.Sleep(10 * time.Second)
	code := httpGet(t, "http://localhost:8080/healthz")
	if code != 200 {
		t.Errorf("expected HTTP 200 from /healthz, got %d", code)
	}
}
