package services

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/brasa-ai/yankrun/domain"
)

type Replacer interface {
	ReplaceInDir(dir string, replacements domain.InputReplacement, config *domain.Config, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error
	AnalyzeDir(dir string, fileSizeLimit string, startDelim string, endDelim string, config *domain.Config) (map[string]int, error)
}

type FileReplacer struct {
	FileSystem FileSystem
}

func (fr *FileReplacer) ReplaceInDir(dir string, replacements domain.InputReplacement, config *domain.Config, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error {
	fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
	if err != nil {
		return err
	}

	return fr.replacePatterns(dir, replacements, config, fileSizeInBytes, startDelim, endDelim, verbose)
}

// AnalyzeDir returns a map of placeholder -> count discovered in files within size limit
func (fr *FileReplacer) AnalyzeDir(dir string, fileSizeLimit string, startDelim string, endDelim string, config *domain.Config) (map[string]int, error) {
	result := map[string]int{}
	fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
	if err != nil {
		return result, err
	}
	var cfgIgnores []string
	if config != nil && len(config.IgnorePath) > 0 {
		cfgIgnores = append(cfgIgnores, config.IgnorePath...)
	}
	err = fr.walkAndAnalyze(dir, fileSizeInBytes, startDelim, endDelim, cfgIgnores, result)
	return result, err
}

func (fr *FileReplacer) walkAndAnalyze(dir string, fileSizeInBytes int64, startDelim string, endDelim string, ignorePatterns []string, result map[string]int) error {
	files, err := fr.FileSystem.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := fr.FileSystem.Join(dir, file.Name())
		if shouldIgnorePath(path, ignorePatterns) {
			continue
		}
		info, err := fr.FileSystem.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip common directories
			switch file.Name() {
			case ".git", "node_modules", "vendor", "dist", "build", "bin":
				continue
			}
			if err := fr.walkAndAnalyze(path, fileSizeInBytes, startDelim, endDelim, ignorePatterns, result); err != nil {
				return err
			}
			continue
		}
		if !fr.checkFileSize(info, fileSizeInBytes, false) {
			continue
		}
		content, err := fr.FileSystem.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(content) || isBinaryByExt(path) {
			continue
		}
		text := string(content)
		// simple scan for startDelim ... endDelim occurrences
		for {
			start := strings.Index(text, startDelim)
			if start == -1 {
				break
			}
			text = text[start+len(startDelim):]
			end := strings.Index(text, endDelim)
			if end == -1 {
				break
			}
			key := text[:end]
			result[key] = result[key] + 1
			text = text[end+len(endDelim):]
		}
	}
	return nil
}

func (fr *FileReplacer) replacePatterns(dir string, replacements domain.InputReplacement, config *domain.Config, fileSizeInBytes int64, startDelim string, endDelim string, verbose bool) error {
	files, err := fr.FileSystem.ReadDir(dir)
	if err != nil {
		return err
	}

	// build ignore patterns: config + input replacements
	var ignorePatterns []string
	if config != nil && len(config.IgnorePath) > 0 {
		ignorePatterns = append(ignorePatterns, config.IgnorePath...)
	}
	if len(replacements.IgnorePath) > 0 {
		ignorePatterns = append(ignorePatterns, replacements.IgnorePath...)
	}

	for _, file := range files {
		path := fr.FileSystem.Join(dir, file.Name())
		if shouldIgnorePath(path, ignorePatterns) {
			if verbose {
				fmt.Printf("Skipping path by ignore_patterns: %s\n", path)
			}
			continue
		}
		info, err := fr.FileSystem.Stat(path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip common directories
			switch file.Name() {
			case ".git", "node_modules", "vendor", "dist", "build", "bin":
				continue
			}
			err := fr.replacePatterns(path, replacements, config, fileSizeInBytes, startDelim, endDelim, verbose)
			if err != nil {
				return err
			}
			continue
		}

		if !fr.checkFileSize(info, fileSizeInBytes, verbose) {
			continue
		}

		content, err := fr.FileSystem.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(content) || isBinaryByExt(path) {
			continue
		}

		newContent := string(content)
		numReplacements := 0
		for _, replacement := range replacements.Variables {
			// Determine the token to search for: explicit key or wrapped with delimiters
			token := replacement.Key
			if !strings.Contains(token, startDelim) && !strings.Contains(token, endDelim) && startDelim != "" && endDelim != "" {
				token = startDelim + replacement.Key + endDelim
			}

			if token != "" {
				count := strings.Count(newContent, token)
				if count > 0 {
					// Apply smart transformations based on template name
					transformedValue := fr.applySmartTransformations(replacement.Key, replacement.Value, replacements.Functions, config)
					newContent = strings.ReplaceAll(newContent, token, transformedValue)
					numReplacements += count
				}
			}
		}

		err = fr.FileSystem.WriteFile(path, []byte(newContent), 0644)
		if err != nil {
			return err
		}

		if verbose && numReplacements != 0 {
			fmt.Printf("Replaced %d instances in %s\n", numReplacements, file.Name())
		}
	}

	return nil
}

// applySmartTransformations applies smart transformations based on template name and functions
func (fr *FileReplacer) applySmartTransformations(templateName, value string, inputFunctions domain.Functions, config *domain.Config) string {
	result := value

	// Check for built-in transformations
	if strings.Contains(templateName, "APPLY_UPPERCASE") {
		result = strings.ToUpper(result)
	}
	if strings.Contains(templateName, "APPLY_DOWNCASE") {
		result = strings.ToLower(result)
	}

	// Apply custom replace functions only if template name contains APPLY_REPLACE
	if strings.Contains(templateName, "APPLY_REPLACE") {
		// Merge config functions with input functions (input functions take precedence)
		mergedFunctions := make(map[string]string)

		// Add config functions first
		if config != nil && config.Functions.APPLY_REPLACE != nil {
			for from, to := range config.Functions.APPLY_REPLACE {
				mergedFunctions[from] = to
			}
		}

		// Override with input functions
		if inputFunctions.APPLY_REPLACE != nil {
			for from, to := range inputFunctions.APPLY_REPLACE {
				mergedFunctions[from] = to
			}
		}

		// Apply merged functions
		for from, to := range mergedFunctions {
			result = strings.ReplaceAll(result, from, to)
		}
	}

	return result
}

func isBinaryByExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".pdf", ".zip", ".gz", ".tar", ".tgz", ".xz", ".rar", ".7z", ".exe", ".dll", ".so":
		return true
	}
	return false
}

func isBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if bytes.IndexByte(data, 0x00) >= 0 {
		return true
	}
	if utf8.Valid(data) {
		nonPrintable := 0
		for _, b := range data {
			if b == '\n' || b == '\r' || b == '\t' {
				continue
			}
			if b < 0x20 || b == 0x7f {
				nonPrintable++
				if nonPrintable > 8 {
					return true
				}
			}
		}
		return false
	}
	highBytes := 0
	for _, b := range data {
		if b >= 0x80 {
			highBytes++
			if highBytes > 16 {
				return true
			}
		}
	}
	return false
}

func (fr *FileReplacer) stringToBytes(size string) (int64, error) {
	size = strings.TrimSpace(size)

	var numStr, unitStr string
	for i, r := range size {
		if unicode.IsLetter(r) {
			numStr = size[:i]
			unitStr = size[i:]
			break
		}
	}

	if numStr == "" || unitStr == "" {
		return 0, fmt.Errorf("invalid format: %s", size)
	}

	num, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
	if err != nil {
		return 0, err
	}

	var multiplier int64 = 1
	switch strings.ToLower(strings.TrimSpace(unitStr)) {
	case "kb":
		multiplier = 1024
	case "mb":
		multiplier = 1024 * 1024
	case "gb":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid unit: %s", unitStr)
	}

	return num * multiplier, nil
}

func (fr *FileReplacer) checkFileSize(fileInfo interface {
	Size() int64
	Name() string
}, fileSizeLimit int64, verbose bool) bool {
	if fileInfo.Size() > fileSizeLimit {
		if verbose {
			fmt.Printf("Skipping file %s because its size (%d) exceeds the limit (%d)\n", fileInfo.Name(), fileInfo.Size(), fileSizeLimit)
		}
		return false
	}
	return true
}

func shouldIgnorePath(path string, ignorePatterns []string) bool {
	if len(ignorePatterns) == 0 {
		return false
	}
	lowerPath := strings.ToLower(path)
	for _, p := range ignorePatterns {
		pp := strings.ToLower(strings.TrimSpace(p))
		if pp == "" {
			continue
		}
		// simple contains match to honor common patterns like "node_modules", "dist", or any substring
		if strings.Contains(lowerPath, pp) {
			return true
		}
	}
	return false
}
