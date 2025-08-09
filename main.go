package main

import (
    "fmt"
    "os"

    "yankrun/actions"
    "yankrun/helpers"
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
    templateAction := actions.NewTemplateAction(fs, parser, replacer)
	cloneAction := actions.NewCloneAction(fs, parser, replacer, cloner)

    app := cli.NewApp()
	app.Name = "yankrun"
	app.Usage = "It templates values and repos"
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
            Flags:   []cli.Flag{inputFlag, dirFlag, verboseFlag, fileSizeLimitFlag, startDelimFlag, endDelimFlag, interactiveFlag},
			Action:  templateAction.Execute,
		},
		{
			Name:    "clone",
			Aliases: []string{"r"},
			Usage:   "Clone a repo with template file replacements",
            Flags:   []cli.Flag{repoFlag, inputFlag, outputDirFlag, verboseFlag, fileSizeLimitFlag, startDelimFlag, endDelimFlag},
			Action:  cloneAction.Execute,
		},
        {
            Name:  "setup",
            Usage: "create or update ~/.yankrun/config.yaml (use --show to display)",
            Flags: []cli.Flag{&cli.BoolFlag{Name: "show", Usage: "show current configuration"}},
            Action: func(c *cli.Context) error {
                return actions.RunSetup(os.Args[2:])
            },
        },
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
