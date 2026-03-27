package tests

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSetupHelm(t *testing.T) {
	skipIfNoAPI(t)
	if testing.Short() {
		t.Skip("skipping infrastructure test in short mode")
	}
	if fixtureDir == "" {
		t.Skip("bootstrap fixture not available")
	}

	dir := t.TempDir()
	copyDir(t, fixtureDir, filepath.Join(dir, "project"))
	projectDir := filepath.Join(dir, "project")

	// L2: file structure
	assertFileExists(t, projectDir, "deploy/helm/testapp/Chart.yaml")
	assertFileContains(t, projectDir, "deploy/helm/testapp/Chart.yaml", "testapp")
	runCmd(t, projectDir, "helm", "lint", "deploy/helm/testapp")

	// L3: Kind cluster + Helm deploy
	runMake(t, projectDir, "cluster-create")
	t.Cleanup(func() {
		runCmd(t, projectDir, "make", "cluster-delete")
	})

	runMake(t, projectDir, "build")
	runCmd(t, projectDir, "docker", "build", "-t", "testapp:latest", ".")
	runCmd(t, projectDir, "kind", "load", "docker-image", "testapp:latest", "--name", "testapp")

	runCmd(t, projectDir, "helm", "install", "testapp", "deploy/helm/testapp",
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
