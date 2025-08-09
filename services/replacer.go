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
    ReplaceInDir(dir string, replacements domain.InputReplacement, fileSizeLimit string, startDelim string, endDelim string, verbose bool) error
    AnalyzeDir(dir string, fileSizeLimit string, startDelim string, endDelim string) (map[string]int, error)
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
func (fr *FileReplacer) AnalyzeDir(dir string, fileSizeLimit string, startDelim string, endDelim string) (map[string]int, error) {
    result := map[string]int{}
    fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
    if err != nil {
        return result, err
    }
    err = fr.walkAndAnalyze(dir, fileSizeInBytes, startDelim, endDelim, result)
    return result, err
}

func (fr *FileReplacer) walkAndAnalyze(dir string, fileSizeInBytes int64, startDelim string, endDelim string, result map[string]int) error {
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
            if err := fr.walkAndAnalyze(path, fileSizeInBytes, startDelim, endDelim, result); err != nil {
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
		for _, replacement := range replacements.Variables {
            // Determine the token to search for: explicit key or wrapped with delimiters
            token := replacement.Key
            if !strings.Contains(token, startDelim) && !strings.Contains(token, endDelim) && startDelim != "" && endDelim != "" {
                token = startDelim + replacement.Key + endDelim
            }

            if token != "" {
                count := strings.Count(newContent, token)
                if count > 0 {
                    newContent = strings.ReplaceAll(newContent, token, replacement.Value)
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

func isBinaryByExt(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".pdf", ".zip", ".gz", ".tar", ".tgz", ".xz", ".rar", ".7z", ".exe", ".dll", ".so":
        return true
    }
    return false
}

func isBinary(data []byte) bool {
    if len(data) == 0 { return false }
    if bytes.IndexByte(data, 0x00) >= 0 { return true }
    if utf8.Valid(data) {
        nonPrintable := 0
        for _, b := range data {
            if b == '\n' || b == '\r' || b == '\t' { continue }
            if b < 0x20 || b == 0x7f {
                nonPrintable++
                if nonPrintable > 8 { return true }
            }
        }
        return false
    }
    highBytes := 0
    for _, b := range data { if b >= 0x80 { highBytes++; if highBytes > 16 { return true } } }
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
