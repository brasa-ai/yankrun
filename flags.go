package main

import "github.com/urfave/cli"

var inputFlag = cli.StringFlag{
	Name:  "input, i",
	Value: "",
	Usage: "Input file with values for replacement",
}

var repoFlag = cli.StringFlag{
	Name:  "repo, r",
	Value: "",
	Usage: "Url to execute clone actions",
}

var outputDirFlag = cli.StringFlag{
	Name:  "outputDir, od",
	Value: "",
	Usage: "Output directory to execute clone actions",
}

var verboseFlag = cli.BoolFlag{
	Name:  "verbose, v",
	Usage: "Enable verbose mode for detailed logs",
}

var fileSizeLimitFlag = cli.StringFlag{
	Name:  "fileSizeLimit, fl",
	Value: "5 mb",
	Usage: "File size limit to ignore replacements from files that exceed the limit",
}
