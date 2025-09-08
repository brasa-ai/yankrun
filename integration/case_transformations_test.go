package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaseTransformations(t *testing.T) {
	bin := buildBinary(t)
	testDir := t.TempDir()

	// Create a test file with placeholders
	testFile := filepath.Join(testDir, "test.txt")
	testContent := `Company: <!Company!>
Company lowercase: <!Company:toLowerCase!>
Company uppercase: <!Company:toUpperCase!>
Team: <!Team!>
Team lowercase: <!Team:toLowerCase!>
Greeting: <!GREETING:gsub(WORLD,HELM):toUpperCase!>
Spaces: <!SPACES:gsub( ,-):toLowerCase!>
Empty Gsub: <!EMPTY_GSUB:gsub( ,_):toUpperCase!>`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create values file with simple key=value format
	valuesFile := filepath.Join(testDir, "values.yaml")
	valuesContent := `variables:
  - key: Company
    value: Your Company
  - key: Team
    value: Your Team
  - key: GREETING
    value: Hello WORLD
  - key: SPACES
    value: Hello World Test
  - key: EMPTY_GSUB
    value: This has spaces`

	if err := os.WriteFile(valuesFile, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("failed to create values file: %v", err)
	}

	// Run template command
	cmd := exec.Command(bin,
		"template",
		"--dir", testDir,
		"--input", valuesFile,
		"--startDelim", "<!",
		"--endDelim", "!>",
		"--verbose",
	)
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("template command failed: %v\n%s", err, string(out))
	}

	// Read the transformed file
	transformedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read transformed file: %v", err)
	}

	content := string(transformedContent)

	// Verify transformations
	tests := []struct {
		name     string
		expected string
		desc     string
	}{
		{"normal", "Your Company", "normal replacement"},
		{"lowercase", "your company", "toLowerCase transformation"},
		{"uppercase", "YOUR COMPANY", "toUpperCase transformation"},
		{"team_normal", "Your Team", "team normal replacement"},
		{"team_lowercase", "your team", "team toLowerCase transformation"},
		{"greeting_gsub_uppercase", "HELLO HELM", "gsub and toUpperCase transformation"},
		{"spaces_gsub_lowercase", "hello-world-test", "gsub with space and toLowerCase transformation"},
		{"empty_gsub_uppercase", "THIS_HAS_SPACES", "gsub with empty string and toUpperCase transformation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(content, tt.expected) {
				t.Errorf("expected %s to contain %q for %s.\nFull content:\n%s", content, tt.expected, tt.desc, content)
			}
		})
	}

	// Verify placeholders are gone
	if strings.Contains(content, "<!Company:toLowerCase!>") {
		t.Error("placeholder <!Company:toLowerCase!> still exists after replacement")
	}
	if strings.Contains(content, "<!Company:toUpperCase!>") {
		t.Error("placeholder <!Company:toUpperCase!> still exists after replacement")
	}
	if strings.Contains(content, "<!GREETING:gsub(WORLD,HELM):toUpperCase!>") {
		t.Error("placeholder <!GREETING:gsub(WORLD,HELM):toUpperCase!> still exists after replacement")
	}
	if strings.Contains(content, "<!SPACES:gsub( ,-):toLowerCase!>") {
		t.Error("placeholder <!SPACES:gsub( ,-):toLowerCase!> still exists after replacement")
	}
	if strings.Contains(content, "<!EMPTY_GSUB:gsub( ,_):toUpperCase!>") {
		t.Error("placeholder <!EMPTY_GSUB:gsub( ,_):toUpperCase!> still exists after replacement")
	}
}
