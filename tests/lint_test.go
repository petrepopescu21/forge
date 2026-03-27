package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	skillsDir  = "../skills"
	pluginFile = "../plugin.json"
	depsFile   = "../skills/dependencies.yaml"
)

type DepsManifest struct {
	Skills map[string]SkillEntry `yaml:"skills"`
}

type SkillEntry struct {
	Type    string   `yaml:"type"`
	Depends []string `yaml:"depends"`
}

type PluginJSON struct {
	Skills []string `json:"skills"`
}

func loadDepsManifest(t *testing.T) DepsManifest {
	t.Helper()
	data, err := os.ReadFile(depsFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", depsFile, err)
	}
	var m DepsManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to parse %s: %v", depsFile, err)
	}
	return m
}

func loadPluginJSON(t *testing.T) PluginJSON {
	t.Helper()
	data, err := os.ReadFile(pluginFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", pluginFile, err)
	}
	var p PluginJSON
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("failed to parse %s: %v", pluginFile, err)
	}
	return p
}

func listSkillDirs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("failed to read %s: %v", skillsDir, err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

func TestFrontmatterValid(t *testing.T) {
	frontmatterRe := regexp.MustCompile(`(?s)^---\n(.+?)\n---`)

	for _, dir := range listSkillDirs(t) {
		skillMd := filepath.Join(skillsDir, dir, "skill.md")
		data, err := os.ReadFile(skillMd)
		if err != nil {
			t.Errorf("skill %s: cannot read skill.md: %v", dir, err)
			continue
		}

		match := frontmatterRe.FindSubmatch(data)
		if match == nil {
			t.Errorf("skill %s: missing YAML frontmatter (---...---)", dir)
			continue
		}

		var fm map[string]interface{}
		if err := yaml.Unmarshal(match[1], &fm); err != nil {
			t.Errorf("skill %s: invalid YAML in frontmatter: %v", dir, err)
			continue
		}

		if _, ok := fm["name"]; !ok {
			t.Errorf("skill %s: frontmatter missing 'name' field", dir)
		}
		if _, ok := fm["description"]; !ok {
			t.Errorf("skill %s: frontmatter missing 'description' field", dir)
		}
	}
}

func TestPluginJsonConsistency(t *testing.T) {
	plugin := loadPluginJSON(t)
	dirs := listSkillDirs(t)

	pluginSkills := make(map[string]bool)
	for _, s := range plugin.Skills {
		name := strings.TrimPrefix(s, "skills/")
		pluginSkills[name] = true
	}

	dirSet := make(map[string]bool)
	for _, d := range dirs {
		dirSet[d] = true
	}

	for _, d := range dirs {
		if !pluginSkills[d] {
			t.Errorf("skill directory %q exists but is not listed in plugin.json", d)
		}
	}

	for name := range pluginSkills {
		if !dirSet[name] {
			t.Errorf("plugin.json lists %q but no skill directory exists", name)
		}
	}
}

func TestDependencyManifestConsistency(t *testing.T) {
	manifest := loadDepsManifest(t)
	dirs := listSkillDirs(t)
	dirSet := make(map[string]bool)
	for _, d := range dirs {
		dirSet[d] = true
	}

	for skill, entry := range manifest.Skills {
		for _, dep := range entry.Depends {
			if !dirSet[dep] {
				t.Errorf("skill %q depends on %q, but no skill directory exists for it", skill, dep)
			}
			if _, ok := manifest.Skills[dep]; !ok {
				t.Errorf("skill %q depends on %q, but %q is not in dependencies.yaml", skill, dep, dep)
			}
		}
	}

	forgeRef := regexp.MustCompile(`forge:([a-z][-a-z0-9]*)`)
	for _, dir := range dirs {
		skillMd := filepath.Join(skillsDir, dir, "skill.md")
		data, err := os.ReadFile(skillMd)
		if err != nil {
			continue
		}
		matches := forgeRef.FindAllSubmatch(data, -1)
		for _, m := range matches {
			ref := string(m[1])
			if _, ok := manifest.Skills[ref]; !ok {
				t.Errorf("skill %q references forge:%s, but %s is not in dependencies.yaml", dir, ref, ref)
			}
		}
	}
}

func TestNoOrphanSkills(t *testing.T) {
	manifest := loadDepsManifest(t)
	plugin := loadPluginJSON(t)
	dirs := listSkillDirs(t)

	pluginSkills := make(map[string]bool)
	for _, s := range plugin.Skills {
		pluginSkills[strings.TrimPrefix(s, "skills/")] = true
	}

	dirSet := make(map[string]bool)
	for _, d := range dirs {
		dirSet[d] = true
	}

	for name := range manifest.Skills {
		if !pluginSkills[name] {
			t.Errorf("dependencies.yaml has %q but it is not in plugin.json", name)
		}
		if !dirSet[name] {
			t.Errorf("dependencies.yaml has %q but no skill directory exists", name)
		}
	}
}

func TestNoCyclicDependencies(t *testing.T) {
	manifest := loadDepsManifest(t)

	inDegree := make(map[string]int)
	for name := range manifest.Skills {
		inDegree[name] = 0
	}
	for _, entry := range manifest.Skills {
		for _, dep := range entry.Depends {
			inDegree[dep]++
		}
	}

	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, dep := range manifest.Skills[node].Depends {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if visited != len(manifest.Skills) {
		var cycleNodes []string
		for name, deg := range inDegree {
			if deg > 0 {
				cycleNodes = append(cycleNodes, name)
			}
		}
		t.Errorf("dependency cycle detected involving: %s", strings.Join(cycleNodes, ", "))
	}
}
