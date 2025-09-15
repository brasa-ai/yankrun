package main

import (
	"fmt"
	"os"

	"github.com/brasa-ai/yankrun/actions"
	"github.com/brasa-ai/yankrun/helpers"
	"github.com/brasa-ai/yankrun/services"

	"github.com/urfave/cli"
)

// Version information - these will be set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Instantiate required services
	fs := &services.OsFileSystem{}
	parser := &services.YAMLJSONParser{FileSystem: fs}
	replacer := &services.FileReplacer{FileSystem: fs}
	cloner := &services.GitCloner{FileSystem: fs}

	// Pass them to actions
	templateAction := actions.NewTemplateAction(fs, parser, replacer)
	cloneAction := actions.NewCloneAction(fs, parser, replacer, cloner)
	generateAction := actions.NewGenerateAction(fs, cloner, parser, replacer)

	app := cli.NewApp()
	app.Name = "yankrun"
	app.Usage = "It templates values and repos"
	app.Version = Version
	// Setup logger using config defaults
	if cfg, err := services.Load(); err == nil {
		helpers.SetupLogger("info")
		_ = cfg // reserved for future loglevel in yankrun
	} else {
		helpers.SetupLogger("info")
	}
	app.Commands = []cli.Command{
		{
			Name:    "template",
			Aliases: []string{"t"},
			Usage:   "Template values",
			Flags:   []cli.Flag{inputFlag, dirFlag, verboseFlag, fileSizeLimitFlag, startDelimFlag, endDelimFlag, interactiveFlag, processTemplatesFlag},
			Action:  templateAction.Execute,
		},
		{
			Name:    "clone",
			Aliases: []string{"r"},
			Usage:   "Clone a repo with template file replacements",
			Flags:   []cli.Flag{repoFlag, inputFlag, outputDirFlag, verboseFlag, fileSizeLimitFlag, startDelimFlag, endDelimFlag, interactiveFlag, branchFlag, processTemplatesFlag},
			Action:  cloneAction.Execute,
		},
		{
			Name:   "generate",
			Usage:  "Interactively choose a template repo/branch and clone it as a new repo (removes .git)",
			Flags:  []cli.Flag{inputFlag, outputDirFlag, verboseFlag, fileSizeLimitFlag, startDelimFlag, endDelimFlag, interactiveFlag, templateNameFlag, branchFlag, processTemplatesFlag},
			Action: generateAction.Execute,
		},
		{
			Name:  "setup",
			Usage: "create or update ~/.yankrun/config.yaml (use --show to display, --reset to delete)",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "show", Usage: "show current configuration"},
				&cli.BoolFlag{Name: "reset", Usage: "delete ~/.yankrun/config.yaml"},
			},
			Action: func(c *cli.Context) error {
				return actions.RunSetup(os.Args[2:])
			},
		},
		{
			Name:  "version",
			Usage: "show version information",
			Action: func(c *cli.Context) error {
				fmt.Printf("yankrun version %s\n", Version)
				fmt.Printf("Build time: %s\n", BuildTime)
				fmt.Printf("Git commit: %s\n", GitCommit)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
