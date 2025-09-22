package services

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/brasa-ai/yankrun/domain"
)

type Replacer interface {
	ReplaceInDir(dir string, replacements domain.InputReplacement, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error
	AnalyzeDir(dir string, fileSizeLimit string, startDelim string, endDelim string, onlyTemplates bool) (map[string]int, error)
	ProcessTemplateFiles(dir string, replacements domain.InputReplacement, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error
}

type FileReplacer struct {
	FileSystem FileSystem
}

func (fr *FileReplacer) ReplaceInDir(dir string, replacements domain.InputReplacement, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error {
	fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
	if err != nil {
		return err
	}

	return fr.replacePatterns(dir, replacements, fileSizeInBytes, startDelim, endDelim, verbose)
}

// AnalyzeDir returns a map of placeholder -> count discovered in files within size limit
func (fr *FileReplacer) AnalyzeDir(dir string, fileSizeLimit string, startDelim string, endDelim string, onlyTemplates bool) (map[string]int, error) {
	result := map[string]int{}
	fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
	if err != nil {
		return result, err
	}
	err = fr.walkAndAnalyze(dir, fileSizeInBytes, startDelim, endDelim, result, onlyTemplates)
	return result, err
}

func (fr *FileReplacer) walkAndAnalyze(dir string, fileSizeInBytes int64, startDelim string, endDelim string, result map[string]int, onlyTemplates bool) error {
	files, err := fr.FileSystem.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := fr.FileSystem.Join(dir, file.Name())
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
			if err := fr.walkAndAnalyze(path, fileSizeInBytes, startDelim, endDelim, result, onlyTemplates); err != nil {
				return err
			}
			continue
		}

		// Skip non-.tpl files when onlyTemplates is true
		if onlyTemplates && !strings.HasSuffix(file.Name(), ".tpl") {
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
			keyWithTransforms := text[:end]
			baseKey, _, err := fr.parsePlaceholder(keyWithTransforms)
			if err != nil {
				// Log error but continue processing other placeholders
				fmt.Printf("Error parsing placeholder '%s': %v\n", keyWithTransforms, err)
				text = text[end+len(endDelim):]
				continue
			}
			result[baseKey] = result[baseKey] + 1
			text = text[end+len(endDelim):]
		}
	}
	return nil
}

// ProcessTemplateFiles processes .tpl files by evaluating templates and removing .tpl suffix
func (fr *FileReplacer) ProcessTemplateFiles(dir string, replacements domain.InputReplacement, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error {
	fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
	if err != nil {
		return err
	}

	return fr.processTemplateFilesRecursive(dir, replacements, fileSizeInBytes, startDelim, endDelim, verbose)
}

func (fr *FileReplacer) processTemplateFilesRecursive(dir string, replacements domain.InputReplacement, fileSizeInBytes int64, startDelim string, endDelim string, verbose bool) error {
	files, err := fr.FileSystem.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := fr.FileSystem.Join(dir, file.Name())
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
			err := fr.processTemplateFilesRecursive(path, replacements, fileSizeInBytes, startDelim, endDelim, verbose)
			if err != nil {
				return err
			}
			continue
		}

		// Only process .tpl files
		if !strings.HasSuffix(file.Name(), ".tpl") {
			continue
		}

		if !fr.checkFileSize(info, fileSizeInBytes, verbose) {
			continue
		}

		content, err := fr.FileSystem.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(content) {
			continue
		}

		// Process the template content
		newContent := string(content)
		numReplacements := 0
		
		// Create a map for quick lookup of replacement values by base key
		replacementValues := make(map[string]string)
		for _, r := range replacements.Variables {
			replacementValues[r.Key] = r.Value
		}

		// Find all placeholders in the content
		placeholderRegex := regexp.MustCompile(regexp.QuoteMeta(startDelim) + `(.*?)` + regexp.QuoteMeta(endDelim))

		// Find all matches
		matches := placeholderRegex.FindAllStringSubmatchIndex(newContent, -1)

		// Process matches in reverse order to avoid issues with index changes
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			fullMatchStart, fullMatchEnd := match[0], match[1]
			placeholderContentStart, placeholderContentEnd := match[2], match[3]

			placeholderWithTransforms := newContent[placeholderContentStart:placeholderContentEnd]

			baseKey, transformations, err := fr.parsePlaceholder(placeholderWithTransforms)
			if err != nil {
				if verbose {
					fmt.Printf("Error parsing placeholder '%s': %v\n", placeholderWithTransforms, err)
				}
				continue
			}

			// Get the base value
			baseValue, ok := replacementValues[baseKey]
			if !ok {
				// If no replacement value is found, skip this placeholder
				continue
			}

			// Apply transformations
			finalValue, err := fr.applyTransformations(baseValue, transformations)
			if err != nil {
				if verbose {
					fmt.Printf("Error applying transformations for '%s': %v\n", placeholderWithTransforms, err)
				}
				continue
			}

			// Replace the full placeholder (including delimiters) with the final value
			newContent = newContent[:fullMatchStart] + finalValue + newContent[fullMatchEnd:]
			numReplacements++
		}

		// Create new filename without .tpl suffix
		newPath := strings.TrimSuffix(path, ".tpl")
		
		// Write the processed content to the new file
		err = fr.FileSystem.WriteFile(newPath, []byte(newContent), 0644)
		if err != nil {
			return err
		}

		// Remove the original .tpl file
		err = fr.FileSystem.Remove(path)
		if err != nil {
			return err
		}

		if verbose && numReplacements != 0 {
			fmt.Printf("Processed template %s -> %s (%d replacements)\n", file.Name(), fr.FileSystem.Base(newPath), numReplacements)
		}
	}

	return nil
}

// parsePlaceholder extracts the base key and transformation functions from a placeholder string.
// Example: "WORLD:gsub(WORLD,galaxy):toUpperCase" -> "WORLD", ["gsub(WORLD,galaxy)", "toUpperCase"]
func (fr *FileReplacer) parsePlaceholder(placeholder string) (string, []string, error) {
	parts := strings.Split(placeholder, ":")
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty placeholder")
	}

	baseKey := parts[0]
	var transformations []string
	if len(parts) > 1 {
		transformations = parts[1:]
	}
	return baseKey, transformations, nil
}

func (fr *FileReplacer) replacePatterns(dir string, replacements domain.InputReplacement, fileSizeInBytes int64, startDelim string, endDelim string, verbose bool) error {
	files, err := fr.FileSystem.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := fr.FileSystem.Join(dir, file.Name())
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
			err := fr.replacePatterns(path, replacements, fileSizeInBytes, startDelim, endDelim, verbose)
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
		// Create a map for quick lookup of replacement values by base key
		replacementValues := make(map[string]string)
		for _, r := range replacements.Variables {
			replacementValues[r.Key] = r.Value
		}

		// Find all placeholders in the content
		// This regex finds content between startDelim and endDelim
		// It's a simple non-greedy match, adjust if delimiters can be nested or contain special regex chars
		placeholderRegex := regexp.MustCompile(regexp.QuoteMeta(startDelim) + `(.*?)` + regexp.QuoteMeta(endDelim))

		// Find all matches
		matches := placeholderRegex.FindAllStringSubmatchIndex(newContent, -1)

		// Process matches in reverse order to avoid issues with index changes
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			fullMatchStart, fullMatchEnd := match[0], match[1]
			placeholderContentStart, placeholderContentEnd := match[2], match[3]

			placeholderWithTransforms := newContent[placeholderContentStart:placeholderContentEnd]

			baseKey, transformations, err := fr.parsePlaceholder(placeholderWithTransforms)
			if err != nil {
				fmt.Printf("Error parsing placeholder '%s': %v\n", placeholderWithTransforms, err)
				continue
			}

			// Get the base value
			baseValue, ok := replacementValues[baseKey]
			if !ok {
				// If no replacement value is found, skip this placeholder
				continue
			}

			// Apply transformations
			finalValue, err := fr.applyTransformations(baseValue, transformations)
			if err != nil {
				fmt.Printf("Error applying transformations for '%s': %v\n", placeholderWithTransforms, err)
				continue
			}

			// Replace the full placeholder (including delimiters) with the final value
			newContent = newContent[:fullMatchStart] + finalValue + newContent[fullMatchEnd:]
			numReplacements++
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

// applyTransformations applies a series of transformation functions to a given value.
func (fr *FileReplacer) applyTransformations(value string, transformations []string) (string, error) {
	transformedValue := value
	for _, t := range transformations {
		var err error
		switch {
		case strings.HasPrefix(t, "toUpperCase"):
			transformedValue = strings.ToUpper(transformedValue)
		case strings.HasPrefix(t, "toLowerCase"), strings.HasPrefix(t, "toDownCase"):
			transformedValue = strings.ToLower(transformedValue)
		case strings.HasPrefix(t, "gsub("):
			transformedValue, err = fr.applyGsub(transformedValue, t)
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("unsupported transformation function: %s", t)
		}
	}
	return transformedValue, nil
}

// applyGsub applies the gsub transformation.
// It parses arguments like "gsub(old,new)" or "gsub( ,new)"
func (fr *FileReplacer) applyGsub(value, transformFunc string) (string, error) {
	// Extract arguments from gsub(arg1,arg2)
	argsStr := transformFunc[len("gsub("):strings.LastIndex(transformFunc, ")")]

	// Split arguments, handling escaped commas or commas within quotes if necessary
	// For now, a simple split by comma, assuming no escaped commas or quotes
	parts := fr.splitGsubArgs(argsStr)

	if len(parts) != 2 {
		return "", fmt.Errorf("invalid gsub arguments: %s. Expected gsub(old,new)", transformFunc)
	}

	old := parts[0]
	new := parts[1]

	// Handle empty string for 'old' argument
	if old == "" {
		// Replace all spaces with 'new'
		return strings.ReplaceAll(value, " ", new), nil
	}

	return strings.ReplaceAll(value, old, new), nil
}

// splitGsubArgs splits the arguments for gsub, handling potential empty strings for the 'old' argument.
func (fr *FileReplacer) splitGsubArgs(argsStr string) []string {
	// Find the first comma that is not inside a quoted string (if we were to support quotes)
	// For now, we'll assume no quotes and just split by the first comma.
	commaIndex := strings.Index(argsStr, ",")
	if commaIndex == -1 {
		return []string{argsStr} // Not enough arguments or malformed
	}

	arg1 := argsStr[:commaIndex]
	arg2 := argsStr[commaIndex+1:]

	return []string{arg1, arg2}
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
