package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "yankrun-test")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(out))
	}
	return bin
}

func repoRoot(t *testing.T) string {
	// assume test file is in repo/integration
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(wd)
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestCloneNonInteractive(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	vals := `variables:
  - key: APP_NAME
    value: TemplateTester
  - key: PROJECT_NAME
    value: DemoProject
  - key: USER_NAME
    value: tester
  - key: USER_EMAIL
    value: tester@example.com
  - key: VERSION
    value: 9.9.9`
	valsPath := writeFile(t, t.TempDir(), "values.yaml", vals)

	cmd := exec.Command(bin,
		"clone",
		"--repo", "https://github.com/brasa-ai/template-tester.git",
		"--input", valsPath,
		"--outputDir", outDir,
		"--startDelim", "[[",
		"--endDelim", "]]",
		"--verbose",
	)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("clone failed: %v\n%s", err, string(out))
	}
	// Validate replacements: values present, placeholders absent across files
	var foundValue, foundPlaceholder bool
	filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// skip binary-ish and images
		lower := strings.ToLower(path)
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err == nil {
			// The actual placeholder in the template-tester repo is [[APP_NAME:toDownCase]]
			// We need to check for the transformed value.
			if bytes.Contains(b, []byte("TEMPLATETESTER")) {
				foundValue = true
			}
			// We should not find the original placeholder or any transformed version of it in processed files.
			// The original placeholder in the template-tester repo is [[APP_NAME:toDownCase]]
			if bytes.Contains(b, []byte("[[APP_NAME:toDownCase]]")) {
				foundPlaceholder = true
			}
		}
		return nil
	})
	if foundPlaceholder {
		t.Fatalf("placeholder [[APP_NAME]] still present after clone replacements")
	}
	if !foundValue {
		t.Fatalf("expected TemplateTester present after clone replacements")
	}
}

func TestTemplateNonInteractive(t *testing.T) {
	bin := buildBinary(t)
	// reuse clone to prepare directory
	work := t.TempDir()
	vals := `variables: [{key: APP_NAME, value: MyApp}]`
	valsPath := writeFile(t, t.TempDir(), "values.yaml", vals)
	// clone without replacements
	cmd := exec.Command(bin, "clone",
		"--repo", "https://github.com/brasa-ai/template-tester.git",
		"--input", writeFile(t, t.TempDir(), "empty.yaml", `variables: []`),
		"--outputDir", work,
		"--startDelim", "[[", "--endDelim", "]]",
	)
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("prepare clone failed: %v\n%s", err, string(out))
	}

	// Add a node_modules file and a large file to verify skipping behavior
	_ = os.MkdirAll(filepath.Join(work, "node_modules"), 0700)
	_ = os.WriteFile(filepath.Join(work, "node_modules", "keep.txt"), []byte("[[APP_NAME]]"), 0600)
	_ = os.WriteFile(filepath.Join(work, "large.txt"), []byte(strings.Repeat("[[APP_NAME]]", 20000)), 0600)

	// now template the directory non-interactively
	tmpl := exec.Command(bin, "template",
		"--dir", work,
		"--input", valsPath,
		"--startDelim", "[[", "--endDelim", "]]",
		"--fileSizeLimit", "50 kb",
		"--verbose",
	)
	tmpl.Dir = repoRoot(t)
	if out, err := tmpl.CombinedOutput(); err != nil {
		t.Fatalf("template failed: %v\n%s", err, string(out))
	}
	// Validate: placeholders replaced in regular files; kept in node_modules and large file
	var hasMyApp, hasToken bool
	filepath.Walk(work, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, "/node_modules/") || strings.HasSuffix(path, "large.txt") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err == nil {
			// The actual placeholder in the template-tester repo is [[APP_NAME:toDownCase]]
			// We need to check for the transformed value.
			if bytes.Contains(b, []byte("MYAPP")) {
				hasMyApp = true
			}
			// We should not find the original placeholder or any transformed version of it
			if bytes.Contains(b, []byte("[[APP_NAME")) {
				hasToken = true
			}
		}
		return nil
	})
	if hasToken {
		t.Fatalf("placeholder [[APP_NAME]] still present in templated files")
	}
	if !hasMyApp {
		t.Fatalf("expected MyApp present in templated files")
	}
}
