package actions

import (
    "bufio"
    "errors"
    "flag"
    "fmt"
    "os"
    "strings"

    "github.com/brasa-ai/yankrun/domain"
    "github.com/brasa-ai/yankrun/helpers"
    "github.com/brasa-ai/yankrun/services"
)

func prompt(r *bufio.Reader, label, def string) string {
    fmt.Printf("%s [%s]: ", label, def)
    if v, err := r.ReadString('\n'); err == nil {
        if v = strings.TrimSpace(v); v != "" {
            return v
        }
    }
    return def
}

// RunSetup configures defaults (~/.yankrun/config.yaml). If --show is present, prints current config and exits.
func RunSetup(args []string) error {
    // support --show flag even when invoked from cli.Command Action context
    fs := flag.NewFlagSet("setup", flag.ContinueOnError)
    show := fs.Bool("show", false, "show current configuration")
    _ = fs.Parse(args)

    cfg, err := services.Load()
    if err != nil {
        // proceed with empty config if file doesn't exist yet
        cfg = &domain.Config{}
    }

    if *show {
        // Display current config (pretty)
        helpers.Log.Info().Msg("Current configuration:")
        fmt.Printf("\n  start_delim:     %s\n  end_delim:       %s\n  file_size_limit: %s\n\n", cfg.StartDelim, cfg.EndDelim, cfg.FileSizeLimit)
        return nil
    }

    r := bufio.NewReader(os.Stdin)
    // Defaults if empty
    if cfg.StartDelim == "" {
        cfg.StartDelim = "[[{["
    }
    if cfg.EndDelim == "" {
        cfg.EndDelim = "]}]]"
    }
    if cfg.FileSizeLimit == "" {
        cfg.FileSizeLimit = "3 mb"
    }

    cfg.StartDelim = prompt(r, "Template start delimiter", cfg.StartDelim)
    cfg.EndDelim = prompt(r, "Template end delimiter", cfg.EndDelim)
    cfg.FileSizeLimit = prompt(r, "File size limit (e.g. 3 mb)", cfg.FileSizeLimit)

    if err := services.Save(cfg); err != nil {
        return errors.New("failed to save config: " + err.Error())
    }
    helpers.Log.Info().Msg("Configuration saved ✔")
    return nil
}


