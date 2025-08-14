package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brasa-ai/yankrun/domain"
)

func TestApplySmartTransformations(t *testing.T) {
	replacer := &FileReplacer{}

	tests := []struct {
		name      string
		template  string
		value     string
		functions domain.Functions
		expected  string
	}{
		{
			name:      "no transformations",
			template:  "APP_NAME",
			value:     "my-app",
			functions: domain.Functions{},
			expected:  "my-app",
		},
		{
			name:      "APPLY_UPPERCASE transformation",
			template:  "APP_NAME_APPLY_UPPERCASE",
			value:     "my-app",
			functions: domain.Functions{},
			expected:  "MY-APP",
		},
		{
			name:      "APPLY_DOWNCASE transformation",
			template:  "PROJECT_NAME_APPLY_DOWNCASE",
			value:     "MyProject",
			functions: domain.Functions{},
			expected:  "myproject",
		},
		{
			name:     "APPLY_REPLACE transformation",
			template: "SLUG_APPLY_REPLACE",
			value:    "my-project-name",
			functions: domain.Functions{
				APPLY_REPLACE: map[string]string{
					"-": "_",
					" ": "-",
				},
			},
			expected: "my_project_name",
		},
		{
			name:     "multiple transformations",
			template: "APP_NAME_APPLY_UPPERCASE",
			value:    "my-project-name",
			functions: domain.Functions{
				APPLY_REPLACE: map[string]string{
					"-": "_",
				},
			},
			expected: "MY-PROJECT-NAME",
		},
		{
			name:     "APPLY_REPLACE only",
			template: "SLUG",
			value:    "my project name",
			functions: domain.Functions{
				APPLY_REPLACE: map[string]string{
					" ": "-",
				},
			},
			expected: "my project name",
		},
		{
			name:     "both APPLY_UPPERCASE and APPLY_REPLACE",
			template: "APP_NAME_APPLY_UPPERCASE_APPLY_REPLACE",
			value:    "my-project-name",
			functions: domain.Functions{
				APPLY_REPLACE: map[string]string{
					"-": "_",
				},
			},
			expected: "MY_PROJECT_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.applySmartTransformations(tt.template, tt.value, tt.functions, nil)
			if result != tt.expected {
				t.Errorf("applySmartTransformations(%q, %q, %+v) = %q, want %q",
					tt.template, tt.value, tt.functions, result, tt.expected)
			}
		})
	}
}

func TestReplacePatternsWithSmartTransformations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file with placeholders
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := `APP_NAME: MY-APP
PROJECT_NAME: [[PROJECT_NAME_APPLY_DOWNCASE]]
SLUG: [[SLUG_APPLY_REPLACE]]
COMPANY: [[Company]]
TEAM: [[Team]]`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create replacer with mock filesystem
	replacer := &FileReplacer{
		FileSystem: &OsFileSystem{},
	}

	// Define replacements with functions
	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "APP_NAME_APPLY_UPPERCASE", Value: "my-app"},
			{Key: "PROJECT_NAME_APPLY_DOWNCASE", Value: "MyProject"},
			{Key: "SLUG_APPLY_REPLACE", Value: "my-project-name"},
			{Key: "Company", Value: "Your Company"},
			{Key: "Team", Value: "Your Team"},
		},
		Functions: domain.Functions{
			APPLY_REPLACE: map[string]string{
				"-": "_",
				" ": "-",
			},
		},
	}

	// Perform replacement
	err = replacer.ReplaceInDir(tempDir, replacements, nil, "1mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("ReplaceInDir failed: %v", err)
	}

	// Read the file and verify transformations
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	expectedContent := `APP_NAME: MY-APP
PROJECT_NAME: myproject
SLUG: my_project_name
COMPANY: Your Company
TEAM: Your Team`

	if string(content) != expectedContent {
		t.Errorf("File content mismatch.\nExpected: %q\nGot: %q", expectedContent, string(content))
	}
}

func TestReplacePatternsWithoutFunctions(t *testing.T) {
	// Test that replacements work without functions defined
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := `APP_NAME: my-app
PROJECT_NAME: [[PROJECT_NAME]]`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	replacer := &FileReplacer{
		FileSystem: &OsFileSystem{},
	}

	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "APP_NAME", Value: "my-app"},
			{Key: "PROJECT_NAME", Value: "MyProject"},
		},
		// No functions defined
	}

	err = replacer.ReplaceInDir(tempDir, replacements, nil, "1mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("ReplaceInDir failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	expectedContent := `APP_NAME: my-app
PROJECT_NAME: MyProject`

	if string(content) != expectedContent {
		t.Errorf("File content mismatch.\nExpected: %q\nGot: %q", expectedContent, string(content))
	}
}
