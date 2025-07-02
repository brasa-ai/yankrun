package services

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"yankrun/domain"
)

type Replacer interface {
	ReplaceInDir(dir string, replacements domain.InputReplacement, fileSizeLimit string, verbose bool) error
}

type FileReplacer struct {
	FileSystem FileSystem
}

func (fr *FileReplacer) ReplaceInDir(dir string, replacements domain.InputReplacement, fileSizeLimit string, verbose bool) error {
	fileSizeInBytes, err := fr.stringToBytes(fileSizeLimit)
	if err != nil {
		return err
	}

	return fr.replacePatterns(dir, replacements, fileSizeInBytes, verbose)
}

func (fr *FileReplacer) replacePatterns(dir string, replacements domain.InputReplacement, fileSizeInBytes int64, verbose bool) error {
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
			err := fr.replacePatterns(path, replacements, fileSizeInBytes, verbose)
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

		newContent := string(content)
		numReplacements := 0
		for _, replacement := range replacements.Variables {
			count := strings.Count(newContent, replacement.Key)
			if count > 0 {
				newContent = strings.ReplaceAll(newContent, replacement.Key, replacement.Value)
				numReplacements += count
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
