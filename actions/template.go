package actions

import (
	"fmt"

	"yankrun/services"

	"github.com/urfave/cli"
)

type TemplateAction struct {
	parser services.ReplacementParser
}

func NewTemplateAction(parser services.ReplacementParser) *TemplateAction {
	return &TemplateAction{parser: parser}
}

func (t *TemplateAction) Execute(c *cli.Context) error {
	inputFile := c.String("input")
	fmt.Printf("Templating with input: %s\n", inputFile)

	// Here we could parse the file just to ensure correctness if needed:
	_, err := t.parser.Parse(inputFile)
	if err != nil {
		return err
	}

	// Additional logic could be added if needed.
	return nil
}
