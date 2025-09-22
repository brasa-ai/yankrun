package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brasa-ai/yankrun/domain"
)

func TestProcessTemplateFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test .tpl files
	tplContent1 := `Hello [[NAME]]!
This is a template file with [[PROJECT_NAME]].
Version: [[VERSION:toUpperCase]]`

	tplContent2 := `Config for [[APP_NAME:toLowerCase]]:
Database: [[DB_NAME]]
Port: [[PORT]]`

	// Write test files
	tplFile1 := filepath.Join(tempDir, "readme.tpl")
	tplFile2 := filepath.Join(tempDir, "config.tpl")
	regularFile := filepath.Join(tempDir, "regular.txt")

	if err := os.WriteFile(tplFile1, []byte(tplContent1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(tplFile2, []byte(tplContent2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(regularFile, []byte("This is a regular file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create replacer instance
	fs := &OsFileSystem{}
	replacer := &FileReplacer{FileSystem: fs}

	// Define replacements
	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "NAME", Value: "John"},
			{Key: "PROJECT_NAME", Value: "TestProject"},
			{Key: "VERSION", Value: "1.0.0"},
			{Key: "APP_NAME", Value: "MyApp"},
			{Key: "DB_NAME", Value: "testdb"},
			{Key: "PORT", Value: "8080"},
		},
	}

	// Process template files
	err := replacer.ProcessTemplateFiles(tempDir, replacements, "3 mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("ProcessTemplateFiles failed: %v", err)
	}

	// Check that .tpl files were removed
	if _, err := os.Stat(tplFile1); err == nil {
		t.Error("readme.tpl should have been removed")
	}
	if _, err := os.Stat(tplFile2); err == nil {
		t.Error("config.tpl should have been removed")
	}

	// Check that new files were created without .tpl suffix
	readmeFile := filepath.Join(tempDir, "readme")
	configFile := filepath.Join(tempDir, "config")

	if _, err := os.Stat(readmeFile); err != nil {
		t.Error("readme file should have been created")
	}
	if _, err := os.Stat(configFile); err != nil {
		t.Error("config file should have been created")
	}

	// Check content of processed files
	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		t.Fatalf("Failed to read readme file: %v", err)
	}

	expectedReadme := `Hello John!
This is a template file with TestProject.
Version: 1.0.0`

	if string(readmeContent) != expectedReadme {
		t.Errorf("readme content mismatch. Expected:\n%s\nGot:\n%s", expectedReadme, string(readmeContent))
	}

	configContent, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	expectedConfig := `Config for myapp:
Database: testdb
Port: 8080`

	if string(configContent) != expectedConfig {
		t.Errorf("config content mismatch. Expected:\n%s\nGot:\n%s", expectedConfig, string(configContent))
	}

	// Check that regular file was not affected
	regularContent, err := os.ReadFile(regularFile)
	if err != nil {
		t.Fatalf("Failed to read regular file: %v", err)
	}

	if string(regularContent) != "This is a regular file" {
		t.Error("Regular file should not have been modified")
	}
}

func TestProcessTemplateFilesWithSubdirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create subdirectories
	subDir1 := filepath.Join(tempDir, "src")
	subDir2 := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(subDir1, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.MkdirAll(subDir2, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test .tpl files in subdirectories
	tplContent1 := `Source file: [[FILE_NAME]]`
	tplContent2 := `Documentation: [[DOC_TITLE]]`

	tplFile1 := filepath.Join(subDir1, "main.tpl")
	tplFile2 := filepath.Join(subDir2, "readme.tpl")

	if err := os.WriteFile(tplFile1, []byte(tplContent1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(tplFile2, []byte(tplContent2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create replacer instance
	fs := &OsFileSystem{}
	replacer := &FileReplacer{FileSystem: fs}

	// Define replacements
	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "FILE_NAME", Value: "app.go"},
			{Key: "DOC_TITLE", Value: "User Guide"},
		},
	}

	// Process template files
	err := replacer.ProcessTemplateFiles(tempDir, replacements, "3 mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("ProcessTemplateFiles failed: %v", err)
	}

	// Check that .tpl files were removed
	if _, err := os.Stat(tplFile1); err == nil {
		t.Error("main.tpl should have been removed")
	}
	if _, err := os.Stat(tplFile2); err == nil {
		t.Error("readme.tpl should have been removed")
	}

	// Check that new files were created
	mainFile := filepath.Join(subDir1, "main")
	readmeFile := filepath.Join(subDir2, "readme")

	if _, err := os.Stat(mainFile); err != nil {
		t.Error("main file should have been created")
	}
	if _, err := os.Stat(readmeFile); err != nil {
		t.Error("readme file should have been created")
	}

	// Check content
	mainContent, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	if string(mainContent) != "Source file: app.go" {
		t.Errorf("main content mismatch. Expected: 'Source file: app.go', Got: '%s'", string(mainContent))
	}

	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		t.Fatalf("Failed to read readme file: %v", err)
	}

	if string(readmeContent) != "Documentation: User Guide" {
		t.Errorf("readme content mismatch. Expected: 'Documentation: User Guide', Got: '%s'", string(readmeContent))
	}
}

func TestProcessTemplateFilesSkipsIgnoredDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create ignored directories
	gitDir := filepath.Join(tempDir, ".git")
	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	vendorDir := filepath.Join(tempDir, "vendor")

	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("Failed to create node_modules directory: %v", err)
	}
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatalf("Failed to create vendor directory: %v", err)
	}

	// Create .tpl files in ignored directories
	tplContent := `This should not be processed: [[VALUE]]`

	gitTplFile := filepath.Join(gitDir, "config.tpl")
	nodeTplFile := filepath.Join(nodeModulesDir, "package.tpl")
	vendorTplFile := filepath.Join(vendorDir, "lib.tpl")

	if err := os.WriteFile(gitTplFile, []byte(tplContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(nodeTplFile, []byte(tplContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(vendorTplFile, []byte(tplContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create replacer instance
	fs := &OsFileSystem{}
	replacer := &FileReplacer{FileSystem: fs}

	// Define replacements
	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "VALUE", Value: "processed"},
		},
	}

	// Process template files
	err := replacer.ProcessTemplateFiles(tempDir, replacements, "3 mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("ProcessTemplateFiles failed: %v", err)
	}

	// Check that .tpl files in ignored directories were NOT processed
	if _, err := os.Stat(gitTplFile); err != nil {
		t.Error(".git/config.tpl should not have been processed")
	}
	if _, err := os.Stat(nodeTplFile); err != nil {
		t.Error("node_modules/package.tpl should not have been processed")
	}
	if _, err := os.Stat(vendorTplFile); err != nil {
		t.Error("vendor/lib.tpl should not have been processed")
	}
}

func TestProcessTemplateFilesWithTransformations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test .tpl file with transformations
	tplContent := `App: [[APP_NAME:toUpperCase]]
Path: [[PATH:gsub( ,-)]]
Mixed: [[MIXED:toLowerCase:gsub(test,prod)]]`

	tplFile := filepath.Join(tempDir, "app.tpl")
	if err := os.WriteFile(tplFile, []byte(tplContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create replacer instance
	fs := &OsFileSystem{}
	replacer := &FileReplacer{FileSystem: fs}

	// Define replacements
	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "APP_NAME", Value: "myapp"},
			{Key: "PATH", Value: "src main"},
			{Key: "MIXED", Value: "TEST_VALUE"},
		},
	}

	// Process template files
	err := replacer.ProcessTemplateFiles(tempDir, replacements, "3 mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("ProcessTemplateFiles failed: %v", err)
	}

	// Check that .tpl file was removed
	if _, err := os.Stat(tplFile); err == nil {
		t.Error("app.tpl should have been removed")
	}

	// Check that new file was created
	appFile := filepath.Join(tempDir, "app")
	if _, err := os.Stat(appFile); err != nil {
		t.Error("app file should have been created")
	}

	// Check content with transformations
	appContent, err := os.ReadFile(appFile)
	if err != nil {
		t.Fatalf("Failed to read app file: %v", err)
	}

	expectedContent := `App: MYAPP
Path: src-main
Mixed: prod_value`

	if string(appContent) != expectedContent {
		t.Errorf("app content mismatch. Expected:\n%s\nGot:\n%s", expectedContent, string(appContent))
	}
}

func TestOnlyTemplatesFunctionality(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files
	tplContent := `Hello [[NAME]]!
This is a template file with [[PROJECT_NAME]].`
	regularContent := `This is a regular file with [[NAME]] placeholder.`

	// Write test files
	tplFile := filepath.Join(tempDir, "readme.tpl")
	regularFile := filepath.Join(tempDir, "regular.txt")
	anotherRegularFile := filepath.Join(tempDir, "config.json")

	if err := os.WriteFile(tplFile, []byte(tplContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(regularFile, []byte(regularContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(anotherRegularFile, []byte(`{"name": "[[NAME]]"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create replacer instance
	fs := &OsFileSystem{}
	replacer := &FileReplacer{FileSystem: fs}

	// Define replacements
	replacements := domain.InputReplacement{
		Variables: []domain.Replacement{
			{Key: "NAME", Value: "John"},
			{Key: "PROJECT_NAME", Value: "TestProject"},
		},
	}

	// Test 1: Process only .tpl files (simulating onlyTemplates=true)
	err := replacer.ProcessTemplateFiles(tempDir, replacements, "3 mb", "[[", "]]", false)
	if err != nil {
		t.Fatalf("Failed to process template files: %v", err)
	}

	// Check that .tpl file was processed and renamed
	processedFile := filepath.Join(tempDir, "readme")
	if _, err := os.Stat(processedFile); os.IsNotExist(err) {
		t.Error("Processed file should exist")
	}

	// Check that original .tpl file was removed
	if _, err := os.Stat(tplFile); !os.IsNotExist(err) {
		t.Error("Original .tpl file should have been removed")
	}

	// Check that regular files were NOT processed
	regularContentAfter, err := os.ReadFile(regularFile)
	if err != nil {
		t.Fatalf("Failed to read regular file: %v", err)
	}
	if string(regularContentAfter) != regularContent {
		t.Errorf("Regular file should not have been processed. Expected: %s, Got: %s", regularContent, string(regularContentAfter))
	}

	// Check that JSON file was NOT processed
	jsonContentAfter, err := os.ReadFile(anotherRegularFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}
	if string(jsonContentAfter) != `{"name": "[[NAME]]"}` {
		t.Errorf("JSON file should not have been processed. Expected: %s, Got: %s", `{"name": "[[NAME]]"}`, string(jsonContentAfter))
	}

	// Check that the processed .tpl file has correct content
	processedContent, err := os.ReadFile(processedFile)
	if err != nil {
		t.Fatalf("Failed to read processed file: %v", err)
	}
	expectedContent := `Hello John!
This is a template file with TestProject.`
	if string(processedContent) != expectedContent {
		t.Errorf("Processed content mismatch. Expected: %s, Got: %s", expectedContent, string(processedContent))
	}
}
