package actions

import (
	"fmt"

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

	if err := a.fs.EnsureDir(outputDir); err != nil {
		return err
	}

	if err := a.cloner.CloneRepository(repoURL, outputDir); err != nil {
		return err
	}

	fmt.Printf("Repository cloned successfully into %s\n", outputDir)

	replacements, err := a.parser.Parse(input)
	if err != nil {
		return err
	}

	err = a.replacer.ReplaceInDir(outputDir, replacements, fileSizeLimit, verbose)
	if err != nil {
		return err
	}

	return nil
}
