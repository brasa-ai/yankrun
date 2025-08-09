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

type TemplateAction struct {
    fs       services.FileSystem
    parser   services.ReplacementParser
    replacer services.Replacer
}

func NewTemplateAction(fs services.FileSystem, parser services.ReplacementParser, replacer services.Replacer) *TemplateAction {
    return &TemplateAction{fs: fs, parser: parser, replacer: replacer}
}

func (t *TemplateAction) Execute(c *cli.Context) error {
    inputFile := c.String("input")
    dir := c.String("dir")
    verbose := c.Bool("verbose")
    interactive := c.Bool("interactive")
    startDelim := c.String("startDelim")
    endDelim := c.String("endDelim")
    fileSizeLimit := c.String("fileSizeLimit")

    if dir == "" {
        return fmt.Errorf("--dir is required for template command")
    }

    // Load defaults from config
    cfg, _ := services.Load()
    if cfg == nil {
        cfg = &domain.Config{}
    }
    if startDelim == "" && cfg.StartDelim != "" { startDelim = cfg.StartDelim }
    if endDelim == "" && cfg.EndDelim != "" { endDelim = cfg.EndDelim }
    if fileSizeLimit == "" && cfg.FileSizeLimit != "" { fileSizeLimit = cfg.FileSizeLimit }
    if startDelim == "" { startDelim = "[[" }
    if endDelim == "" { endDelim = "]]" }
    if fileSizeLimit == "" { fileSizeLimit = "3 mb" }

    var parsed domain.InputReplacement
    var err error
    if inputFile != "" {
        parsed, err = t.parser.Parse(inputFile)
        if err != nil { return err }
    }

    // Analyze placeholders in dir
    counts, err := t.replacer.AnalyzeDir(dir, fileSizeLimit, startDelim, endDelim)
    if err != nil { return err }
    if len(counts) == 0 {
        helpers.Log.Info().Msg("No placeholders found.")
        return nil
    }

    // Merge existing values from parsed file
    values := map[string]string{}
    for _, r := range parsed.Variables { values[r.Key] = r.Value }

    // Pretty print summary
    keys := make([]string, 0, len(counts))
    for k := range counts { keys = append(keys, k) }
    sort.Strings(keys)
    helpers.Log.Info().Msg("Discovered placeholders:")
    for _, k := range keys {
        v := values[k]
        if v == "" { v = "(unset)" }
        fmt.Printf("  %-24s  matches=%-6d  value=%s\n", k, counts[k], v)
    }

    // Interactive prompt for missing values
    if interactive {
        r := bufio.NewReader(os.Stdin)
        for _, k := range keys {
            def := values[k]
            fmt.Printf("Enter value for %s [%s]: ", k, def)
            s, _ := r.ReadString('\n')
            s = strings.TrimSpace(s)
            if s != "" { values[k] = s }
        }
        fmt.Println()
    }

    // Build replacements with final values (use only discovered keys)
    final := domain.InputReplacement{}
    for _, k := range keys {
        if v, ok := values[k]; ok && v != "" {
            final.Variables = append(final.Variables, domain.Replacement{Key: k, Value: v})
        }
    }

    if len(final.Variables) == 0 {
        helpers.Log.Info().Msg("No values provided; nothing to replace.")
        return nil
    }

    if err := t.replacer.ReplaceInDir(dir, final, fileSizeLimit, startDelim, endDelim, verbose); err != nil {
        return err
    }
    helpers.Log.Info().Msg("Templating complete ✔")
    return nil
}
