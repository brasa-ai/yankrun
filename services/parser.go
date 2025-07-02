package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"yankrun/domain"

	"gopkg.in/yaml.v3"
)

type ReplacementParser interface {
	Parse(filePath string) (domain.InputReplacement, error)
}

type YAMLJSONParser struct {
	FileSystem FileSystem
}

func (p *YAMLJSONParser) Parse(filePath string) (domain.InputReplacement, error) {
	var patterns domain.InputReplacement

	data, err := p.FileSystem.ReadFile(filePath)
	if err != nil {
		return patterns, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &patterns)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &patterns)
	default:
		return patterns, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return patterns, err
	}

	return patterns, nil
}
