package actions

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/brasa-ai/yankrun/domain"
	"github.com/brasa-ai/yankrun/helpers"
	"github.com/brasa-ai/yankrun/services"

	"github.com/urfave/cli"
)

type CloneAction struct {
	fs       services.FileSystem
	parser   services.ReplacementParser
	replacer services.Replacer
	cloner   services.Cloner
}

func NewCloneAction(fs services.FileSystem, parser services.ReplacementParser, replacer services.Replacer, cloner services.Cloner) *CloneAction {
	return &CloneAction{
		fs:       fs,
		parser:   parser,
		replacer: replacer,
		cloner:   cloner,
	}
}

func (a *CloneAction) Execute(c *cli.Context) error {
	repoURL := c.String("repo")
	outputDir := c.String("outputDir")
	verbose := c.Bool("verbose")
	input := c.String("input")
	fileSizeLimit := c.String("fileSizeLimit")
	startDelim := c.String("startDelim")
	endDelim := c.String("endDelim")
	interactive := c.Bool("interactive")
	processTemplates := c.Bool("processTemplates")
	onlyTemplates := c.Bool("onlyTemplates")

	// Load defaults from config when flags not provided
	cfg, _ := services.Load()
	if cfg == nil {
		cfg = &domain.Config{}
	}
	if startDelim == "" && cfg.StartDelim != "" {
		startDelim = cfg.StartDelim
	}
	if endDelim == "" && cfg.EndDelim != "" {
		endDelim = cfg.EndDelim
	}
	if fileSizeLimit == "" && cfg.FileSizeLimit != "" {
		fileSizeLimit = cfg.FileSizeLimit
	}
	if startDelim == "" {
		startDelim = "[["
	}
	if endDelim == "" {
		endDelim = "]]"
	}
	if fileSizeLimit == "" {
		fileSizeLimit = "3 mb"
	}

	// Validate flag combination
	if onlyTemplates && !processTemplates {
		return fmt.Errorf("--onlyTemplates requires --processTemplates to be set")
	}

	if err := a.fs.EnsureDir(outputDir); err != nil {
		return err
	}

	if err := a.cloner.CloneRepository(repoURL, outputDir); err != nil {
		return err
	}

	helpers.Log.Info().Msgf("Cloned into %s", outputDir)

	// Parse provided replacements if any
	var provided domain.InputReplacement
	if input != "" {
		var err error
		provided, err = a.parser.Parse(input)
		if err != nil {
			return err
		}
	}

	// Analyze placeholders in cloned directory
	counts, err := a.replacer.AnalyzeDir(outputDir, fileSizeLimit, startDelim, endDelim, onlyTemplates)
	if err != nil {
		return err
	}

	// Build value map from provided input
	values := map[string]string{}
	for _, r := range provided.Variables {
		values[r.Key] = r.Value
	}

	// If interactive, prompt for each discovered key
	final := domain.InputReplacement{}
	if len(counts) > 0 {
		keys := make([]string, 0, len(counts))
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		helpers.Log.Info().Msg("Discovered placeholders:")
		for _, k := range keys {
			v := values[k]
			if v == "" {
				v = "(unset)"
			}
			fmt.Printf("  %-24s  matches=%-6d  value=%s\n", k, counts[k], v)
		}

		if interactive {
			r := bufio.NewReader(os.Stdin)
			for _, k := range keys {
				def := values[k]
				fmt.Printf("Enter value for %s [%s]: ", k, def)
				s, _ := r.ReadString('\n')
				s = strings.TrimSpace(s)
				if s != "" {
					values[k] = s
				}
			}
			fmt.Println()
		}

		for _, k := range keys {
			if v, ok := values[k]; ok && v != "" {
				final.Variables = append(final.Variables, domain.Replacement{Key: k, Value: v})
			}
		}
	} else {
		// No discovered keys; use provided values directly
		final = provided
	}

	// Skip regular templating if onlyTemplates is set
	if !onlyTemplates {
		if err := a.replacer.ReplaceInDir(outputDir, final, fileSizeLimit, startDelim, endDelim, verbose); err != nil {
			return err
		}
	}

	// Process .tpl files if requested
	if processTemplates {
		if err := a.replacer.ProcessTemplateFiles(outputDir, final, fileSizeLimit, startDelim, endDelim, verbose); err != nil {
			return err
		}
		helpers.Log.Info().Msg("Template file processing complete ✔")
	}

	helpers.Log.Info().Msg("Templating complete ✔")

	return nil
}
