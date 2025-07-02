package main

import (
	"fmt"
	"os"

	"yankrun/actions"
	"yankrun/services"

	"github.com/urfave/cli"
)

func main() {
	// Instantiate required services
	fs := &services.OsFileSystem{}
	parser := &services.YAMLJSONParser{FileSystem: fs}
	replacer := &services.FileReplacer{FileSystem: fs}
	cloner := &services.GitCloner{FileSystem: fs}

	// Pass them to actions
	templateAction := actions.NewTemplateAction(parser)
	cloneAction := actions.NewCloneAction(fs, parser, replacer, cloner)

	app := cli.NewApp()
	app.Name = "yankrun"
	app.Usage = "It templates values and repos"
	app.Commands = []cli.Command{
		{
			Name:    "template",
			Aliases: []string{"t"},
			Usage:   "Template values",
			Flags:   []cli.Flag{inputFlag},
			Action:  templateAction.Execute,
		},
		{
			Name:    "clone",
			Aliases: []string{"r"},
			Usage:   "Clone a repo with template file replacements",
			Flags:   []cli.Flag{repoFlag, inputFlag, outputDirFlag, verboseFlag, fileSizeLimitFlag},
			Action:  cloneAction.Execute,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
