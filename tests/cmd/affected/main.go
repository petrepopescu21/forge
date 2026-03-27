package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type DepsManifest struct {
	Skills map[string]SkillEntry `yaml:"skills"`
}

type SkillEntry struct {
	Type    string   `yaml:"type"`
	Depends []string `yaml:"depends"`
}

func main() {
	base := "origin/main"
	if len(os.Args) > 2 && os.Args[1] == "--base" {
		base = os.Args[2]
	}

	out, err := exec.Command("git", "diff", "--name-only", base+"...HEAD").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git diff failed: %v\n", err)
		os.Exit(1)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		os.Exit(0)
	}

	for _, f := range changedFiles {
		if f == "skills/dependencies.yaml" || f == "plugin.json" ||
			strings.HasPrefix(f, "tests/") {
			fmt.Println(".*")
			os.Exit(0)
		}
	}

	data, err := os.ReadFile("skills/dependencies.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read dependencies.yaml: %v\n", err)
		os.Exit(1)
	}
	var manifest DepsManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse dependencies.yaml: %v\n", err)
		os.Exit(1)
	}

	affected := make(map[string]bool)
	for _, f := range changedFiles {
		if !strings.HasPrefix(f, "skills/") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(f, "skills/"), "/", 2)
		if len(parts) == 0 {
			continue
		}
		skillName := parts[0]
		if _, ok := manifest.Skills[skillName]; ok {
			affected[skillName] = true
		}
	}

	if len(affected) == 0 {
		os.Exit(0)
	}

	for name, entry := range manifest.Skills {
		for _, dep := range entry.Depends {
			if affected[dep] {
				affected[name] = true
			}
		}
	}

	var testNames []string
	for name := range affected {
		testName := "Test" + toTestName(name)
		testNames = append(testNames, testName)
	}

	fmt.Println(strings.Join(testNames, "|"))
}

func toTestName(skillName string) string {
	parts := strings.Split(skillName, "-")
	var result strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			result.WriteString(strings.ToUpper(p[:1]))
			result.WriteString(p[1:])
		}
	}
	return result.String()
}
