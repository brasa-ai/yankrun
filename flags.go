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

var dirFlag = cli.StringFlag{
    Name:  "dir, d",
    Value: "",
    Usage: "Target directory for templating (used by template command)",
}

var verboseFlag = cli.BoolFlag{
	Name:  "verbose, v",
	Usage: "Enable verbose mode for detailed logs",
}

var fileSizeLimitFlag = cli.StringFlag{
	Name:  "fileSizeLimit, fl",
    Value: "3 mb",
	Usage: "File size limit to ignore replacements from files that exceed the limit",
}

var startDelimFlag = cli.StringFlag{
    Name:  "startDelim, sd",
    Value: "[[",
    Usage: "Template start delimiter (default [[)",
}

var endDelimFlag = cli.StringFlag{
    Name:  "endDelim, ed",
    Value: "]]",
    Usage: "Template end delimiter (default ]])",
}

var interactiveFlag = cli.BoolFlag{
    Name:  "interactive, prompt, p",
    Usage: "Prompt for values for discovered placeholders before applying",
}
