package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTemplateProcessingIntegration(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	// Create test .tpl files
	tplContent1 := `# [[PROJECT_NAME]]

Welcome to [[PROJECT_NAME]]!

## Configuration
- App Name: [[APP_NAME:toLowerCase]]
- Version: [[VERSION:toUpperCase]]
- Database: [[DB_NAME]]`

	tplContent2 := `package main

import "fmt"

func main() {
    fmt.Println("Hello from [[APP_NAME]]!")
    fmt.Println("Version: [[VERSION]]")
}`

	// Write test files
	tplFile1 := filepath.Join(workDir, "README.md.tpl")
	tplFile2 := filepath.Join(workDir, "main.tpl")
	regularFile := filepath.Join(workDir, "config.txt")

	if err := os.WriteFile(tplFile1, []byte(tplContent1), 0644); err != nil {
		t.Fatalf("Failed to create README.md.tpl: %v", err)
	}
	if err := os.WriteFile(tplFile2, []byte(tplContent2), 0644); err != nil {
		t.Fatalf("Failed to create main.tpl: %v", err)
	}
	if err := os.WriteFile(regularFile, []byte("This is a regular file with [[PLACEHOLDER]]"), 0644); err != nil {
		t.Fatalf("Failed to create config.txt: %v", err)
	}

	// Create values file
	vals := `variables:
  - key: PROJECT_NAME
    value: TestProject
  - key: APP_NAME
    value: MyApp
  - key: VERSION
    value: 1.0.0
  - key: DB_NAME
    value: testdb
  - key: PLACEHOLDER
    value: replaced_value`

	valsPath := writeFile(t, t.TempDir(), "values.yaml", vals)

	// Run template command with processTemplates flag
	cmd := exec.Command(bin,
		"template",
		"--dir", workDir,
		"--input", valsPath,
		"--startDelim", "[[",
		"--endDelim", "]]",
		"--processTemplates",
		"--verbose",
	)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("template command failed: %v\n%s", err, string(out))
	}

	// Check that .tpl files were removed
	if _, err := os.Stat(tplFile1); err == nil {
		t.Error("README.md.tpl should have been removed")
	}
	if _, err := os.Stat(tplFile2); err == nil {
		t.Error("main.tpl should have been removed")
	}

	// Check that new files were created without .tpl suffix
	readmeFile := filepath.Join(workDir, "README.md")
	mainFile := filepath.Join(workDir, "main")

	if _, err := os.Stat(readmeFile); err != nil {
		t.Error("README file should have been created")
	}
	if _, err := os.Stat(mainFile); err != nil {
		t.Error("main file should have been created")
	}

	// Check content of processed files
	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		t.Fatalf("Failed to read README file: %v", err)
	}

	expectedReadme := `# TestProject

Welcome to TestProject!

## Configuration
- App Name: myapp
- Version: 1.0.0
- Database: testdb`

	if string(readmeContent) != expectedReadme {
		t.Errorf("README content mismatch. Expected:\n%s\nGot:\n%s", expectedReadme, string(readmeContent))
	}

	mainContent, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	expectedMain := `package main

import "fmt"

func main() {
    fmt.Println("Hello from MyApp!")
    fmt.Println("Version: 1.0.0")
}`

	if string(mainContent) != expectedMain {
		t.Errorf("main content mismatch. Expected:\n%s\nGot:\n%s", expectedMain, string(mainContent))
	}

	// Check that regular file was also processed (placeholders replaced)
	configContent, err := os.ReadFile(regularFile)
	if err != nil {
		t.Fatalf("Failed to read config.txt: %v", err)
	}

	if !bytes.Contains(configContent, []byte("replaced_value")) {
		t.Error("Regular file should have had placeholders replaced")
	}
	if bytes.Contains(configContent, []byte("[[PLACEHOLDER]]")) {
		t.Error("Regular file should not contain unreplaced placeholders")
	}
}

func TestTemplateProcessingWithoutFlag(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	// Create test .tpl file
	tplContent := `Hello [[NAME]]!`
	tplFile := filepath.Join(workDir, "test.tpl")
	if err := os.WriteFile(tplFile, []byte(tplContent), 0644); err != nil {
		t.Fatalf("Failed to create test.tpl: %v", err)
	}

	// Create values file
	vals := `variables:
  - key: NAME
    value: World`

	valsPath := writeFile(t, t.TempDir(), "values.yaml", vals)

	// Run template command WITHOUT processTemplates flag
	cmd := exec.Command(bin,
		"template",
		"--dir", workDir,
		"--input", valsPath,
		"--startDelim", "[[",
		"--endDelim", "]]",
		"--verbose",
	)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("template command failed: %v\n%s", err, string(out))
	}

	// Check that .tpl file was NOT processed (should still exist)
	if _, err := os.Stat(tplFile); err != nil {
		t.Error("test.tpl should NOT have been removed when processTemplates flag is not set")
	}

	// Check that no new file was created
	newFile := filepath.Join(workDir, "test")
	if _, err := os.Stat(newFile); err == nil {
		t.Error("test file should NOT have been created when processTemplates flag is not set")
	}

	// Check that the .tpl file content was still processed (placeholders replaced)
	tplContentAfter, err := os.ReadFile(tplFile)
	if err != nil {
		t.Fatalf("Failed to read test.tpl: %v", err)
	}

	if !bytes.Contains(tplContentAfter, []byte("Hello World!")) {
		t.Error("test.tpl should have had placeholders replaced even without processTemplates flag")
	}
}

func TestCloneWithTemplateProcessing(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()

	// Create a test repository structure
	testRepoDir := t.TempDir()

	// Create .tpl files in the test repo
	tplContent1 := `# [[PROJECT_NAME]]
Version: [[VERSION]]`
	tplContent2 := `package main

func main() {
    fmt.Println("[[GREETING]]")
}`

	tplFile1 := filepath.Join(testRepoDir, "README.md.tpl")
	tplFile2 := filepath.Join(testRepoDir, "main.tpl")

	if err := os.WriteFile(tplFile1, []byte(tplContent1), 0644); err != nil {
		t.Fatalf("Failed to create README.md.tpl: %v", err)
	}
	if err := os.WriteFile(tplFile2, []byte(tplContent2), 0644); err != nil {
		t.Fatalf("Failed to create main.tpl: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = testRepoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, string(out))
	}

	// Add files and commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = testRepoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, string(out))
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = testRepoDir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, string(out))
	}

	// Create values file
	vals := `variables:
  - key: PROJECT_NAME
    value: ClonedProject
  - key: VERSION
    value: 2.0.0
  - key: GREETING
    value: Hello from cloned project!`

	valsPath := writeFile(t, t.TempDir(), "values.yaml", vals)

	// Clone with template processing
	cmd = exec.Command(bin,
		"clone",
		"--repo", testRepoDir,
		"--input", valsPath,
		"--outputDir", outDir,
		"--startDelim", "[[",
		"--endDelim", "]]",
		"--processTemplates",
		"--verbose",
	)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("clone command failed: %v\n%s", err, string(out))
	}

	// Check that .tpl files were processed
	readmeFile := filepath.Join(outDir, "README.md")
	mainFile := filepath.Join(outDir, "main")

	if _, err := os.Stat(readmeFile); err != nil {
		t.Error("README file should have been created from README.md.tpl")
	}
	if _, err := os.Stat(mainFile); err != nil {
		t.Error("main file should have been created from main.tpl")
	}

	// Check content
	readmeContent, err := os.ReadFile(readmeFile)
	if err != nil {
		t.Fatalf("Failed to read README: %v", err)
	}

	expectedReadme := `# ClonedProject
Version: 2.0.0`

	if string(readmeContent) != expectedReadme {
		t.Errorf("README content mismatch. Expected:\n%s\nGot:\n%s", expectedReadme, string(readmeContent))
	}

	mainContent, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main: %v", err)
	}

	expectedMain := `package main

func main() {
    fmt.Println("Hello from cloned project!")
}`

	if string(mainContent) != expectedMain {
		t.Errorf("main content mismatch. Expected:\n%s\nGot:\n%s", expectedMain, string(mainContent))
	}
}
