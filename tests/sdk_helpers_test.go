package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// skillPrompt reads a skill.md and returns its contents.
func skillPrompt(t *testing.T, skillName string) string {
	t.Helper()
	skillPath := filepath.Join(skillsDir, skillName, "skill.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read skill %s: %v", skillName, err)
	}
	return string(data)
}

// claudeCodeTools defines the tools that mimic Claude Code's capabilities.
var claudeCodeTools = []anthropic.ToolUnionParam{
	{OfTool: &anthropic.ToolParam{
		Name:        "Write",
		Description: anthropic.String("Write content to a file at the given path"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"file_path": map[string]interface{}{"type": "string"},
				"content":   map[string]interface{}{"type": "string"},
			},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "Bash",
		Description: anthropic.String("Execute a bash command"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"command": map[string]interface{}{"type": "string"},
			},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "Edit",
		Description: anthropic.String("Edit a file by replacing old_string with new_string"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"file_path":  map[string]interface{}{"type": "string"},
				"old_string": map[string]interface{}{"type": "string"},
				"new_string": map[string]interface{}{"type": "string"},
			},
		},
	}},
}

// toolCall represents a parsed tool use from Claude's response.
type toolCall struct {
	Name  string
	Input map[string]interface{}
}

// runSkillTest sends a skill prompt to Claude and returns the tool calls.
func runSkillTest(t *testing.T, skillName string, userPrompt string) []toolCall {
	t.Helper()
	skipIfNoAPI(t)

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	skill := skillPrompt(t, skillName)

	model := os.Getenv("FORGE_TEST_MODEL")
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}

	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 8192,
		System: []anthropic.TextBlockParam{
			{Text: "You are Claude Code, an AI coding assistant. You have tools: Write, Bash, Edit. " +
				"When asked to set up a project component, use the Write tool to create files. " +
				"Always use the Write tool — do not just describe what to create. " +
				"Here is the skill you must follow:\n\n" + skill},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(userPrompt),
			),
		},
		Tools: claudeCodeTools,
	})
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}

	calls := extractToolCalls(t, resp)
	t.Logf("skill %s produced %d tool calls: %s", skillName, len(calls), summarizeCalls(calls))
	return calls
}

// extractToolCalls pulls tool use blocks from a Claude response.
func extractToolCalls(t *testing.T, resp *anthropic.Message) []toolCall {
	t.Helper()
	var calls []toolCall
	for _, block := range resp.Content {
		tu := block.AsToolUse()
		if tu.Name == "" {
			continue
		}
		input := make(map[string]interface{})
		if err := json.Unmarshal(tu.Input, &input); err != nil {
			t.Logf("warning: failed to parse tool input for %s: %v", tu.Name, err)
			continue
		}
		calls = append(calls, toolCall{
			Name:  tu.Name,
			Input: input,
		})
	}
	return calls
}

// --- Assertion helpers for tool calls ---

// assertToolCallExists checks that any tool call (Write path, Write content, Bash command)
// contains the given substring.
func assertToolCallExists(t *testing.T, calls []toolCall, substr string) {
	t.Helper()
	for _, c := range calls {
		for _, v := range c.Input {
			if s, ok := v.(string); ok && strings.Contains(s, substr) {
				return
			}
		}
	}
	t.Errorf("expected a tool call containing %q, got: %s", substr, summarizeCalls(calls))
}

func assertWritesFile(t *testing.T, calls []toolCall, path string) {
	t.Helper()
	for _, c := range calls {
		if c.Name == "Write" {
			if fp, ok := c.Input["file_path"].(string); ok && strings.Contains(fp, path) {
				return
			}
		}
	}
	t.Errorf("expected a Write tool call for path containing %q, got: %s", path, summarizeCalls(calls))
}

func assertWritesFileContaining(t *testing.T, calls []toolCall, path string, substr string) {
	t.Helper()
	for _, c := range calls {
		if c.Name == "Write" {
			fp, _ := c.Input["file_path"].(string)
			content, _ := c.Input["content"].(string)
			if strings.Contains(fp, path) && strings.Contains(content, substr) {
				return
			}
		}
	}
	t.Errorf("expected Write to %q containing %q, not found in tool calls", path, substr)
}

func summarizeCalls(calls []toolCall) string {
	var parts []string
	for _, c := range calls {
		fp, _ := c.Input["file_path"].(string)
		cmd, _ := c.Input["command"].(string)
		if fp != "" {
			parts = append(parts, c.Name+"("+fp+")")
		} else if cmd != "" {
			short := cmd
			if len(short) > 60 {
				short = short[:60] + "..."
			}
			parts = append(parts, c.Name+"("+short+")")
		} else {
			parts = append(parts, c.Name)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
