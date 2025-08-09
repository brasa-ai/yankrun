package actions

import (
    "yankrun/domain"
    "yankrun/helpers"
    "yankrun/services"

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
        startDelim = "[[{["
    }
    if endDelim == "" {
        endDelim = "]}]]"
    }
    if fileSizeLimit == "" {
        fileSizeLimit = "3 mb"
    }

	if err := a.fs.EnsureDir(outputDir); err != nil {
		return err
	}

    if err := a.cloner.CloneRepository(repoURL, outputDir); err != nil {
        return err
    }

    helpers.Log.Info().Msgf("Cloned into %s", outputDir)

	replacements, err := a.parser.Parse(input)
	if err != nil {
		return err
	}

    err = a.replacer.ReplaceInDir(outputDir, replacements, fileSizeLimit, startDelim, endDelim, verbose)
    if err != nil {
        return err
    }
    if verbose {
        helpers.Log.Info().Msg("Templating complete ✔")
    }

	return nil
}
