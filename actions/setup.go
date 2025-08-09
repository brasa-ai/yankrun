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
    reset := fs.Bool("reset", false, "delete ~/.yankrun/config.yaml")
    _ = fs.Parse(args)

    if *reset {
        if err := services.Reset(); err != nil {
            return errors.New("failed to delete config: " + err.Error())
        }
        helpers.Log.Info().Msg("Configuration removed ✔")
        return nil
    }

    cfg, err := services.Load()
    if err != nil {
        // proceed with empty config if file doesn't exist yet
        cfg = &domain.Config{}
    }

    if *show {
        // Display current config (pretty)
        helpers.Log.Info().Msg("Current configuration:")
        fmt.Printf("\n  start_delim:     %q\n  end_delim:       %q\n  file_size_limit: %s\n", cfg.StartDelim, cfg.EndDelim, cfg.FileSizeLimit)
        // Show templates
        fmt.Printf("  templates:       %d configured\n", len(cfg.Templates))
        for i, t := range cfg.Templates {
            fmt.Printf("    - [%d] %s (%s) default_branch=%s\n", i+1, t.Name, t.URL, t.DefaultBranch)
        }
        // Show GitHub discovery config (print if anything is set)
        if cfg.GitHub.User != "" || len(cfg.GitHub.Orgs) > 0 || cfg.GitHub.Topic != "" || cfg.GitHub.Prefix != "" || cfg.GitHub.IncludePrivate {
            fmt.Printf("  github.user:     %s\n", cfg.GitHub.User)
            fmt.Printf("  github.orgs:     %s\n", strings.Join(cfg.GitHub.Orgs, ", "))
            fmt.Printf("  github.topic:    %s\n", cfg.GitHub.Topic)
            fmt.Printf("  github.prefix:   %s\n", cfg.GitHub.Prefix)
            fmt.Printf("  github.private:  %t\n", cfg.GitHub.IncludePrivate)
        }
        fmt.Println()
        return nil
    }

    r := bufio.NewReader(os.Stdin)
    // Defaults if empty
    if cfg.StartDelim == "" { cfg.StartDelim = "[[" }
    if cfg.EndDelim == "" { cfg.EndDelim = "]]" }
    if cfg.FileSizeLimit == "" {
        cfg.FileSizeLimit = "3 mb"
    }

    // Clear, unambiguous prompts for delimiters and size
    fmt.Printf("Template start delimiter (current: %q, e.g., [[]): ", cfg.StartDelim)
    if v, err := r.ReadString('\n'); err == nil {
        if s := strings.TrimSpace(v); s != "" { cfg.StartDelim = s }
    }
    fmt.Printf("Template end delimiter (current: %q, e.g., ]]): ", cfg.EndDelim)
    if v, err := r.ReadString('\n'); err == nil {
        if s := strings.TrimSpace(v); s != "" { cfg.EndDelim = s }
    }
    fmt.Printf("File size limit (current: %s, e.g., 3 mb): ", cfg.FileSizeLimit)
    if v, err := r.ReadString('\n'); err == nil {
        if s := strings.TrimSpace(v); s != "" { cfg.FileSizeLimit = s }
    }

    // Add or edit templates
    for {
        add := strings.ToLower(prompt(r, "Add a template repo? (y/N)", "n"))
        if add != "y" && add != "yes" { break }
        t := domain.TemplateRepo{}
        fmt.Printf("Template name (label, e.g., 'Go App' or 'org/repo'): ")
        if v, err := r.ReadString('\n'); err == nil { t.Name = strings.TrimSpace(v) }
        fmt.Printf("Template git URL (SSH/HTTPS, e.g., git@github.com:org/repo.git or https://github.com/org/repo.git): ")
        if v, err := r.ReadString('\n'); err == nil { t.URL = strings.TrimSpace(v) }
        fmt.Printf("Description (optional): ")
        if v, err := r.ReadString('\n'); err == nil { t.Description = strings.TrimSpace(v) }
        fmt.Printf("Default branch (e.g., main): ")
        if v, err := r.ReadString('\n'); err == nil {
            if s := strings.TrimSpace(v); s != "" { t.DefaultBranch = s } else { t.DefaultBranch = "main" }
        }
        if t.URL != "" { cfg.Templates = append(cfg.Templates, t) }
    }

    // Configure GitHub discovery (optional)
    if strings.ToLower(prompt(r, "Configure GitHub discovery? (y/N)", "n")) == "y" {
        cfg.GitHub.User = prompt(r, "GitHub user (leave empty to skip)", cfg.GitHub.User)
        orgs := prompt(r, "GitHub orgs (comma-separated)", strings.Join(cfg.GitHub.Orgs, ","))
        if orgs != "" {
            parts := strings.Split(orgs, ",")
            cfg.GitHub.Orgs = cfg.GitHub.Orgs[:0]
            for _, p := range parts {
                p = strings.TrimSpace(p)
                if p != "" { cfg.GitHub.Orgs = append(cfg.GitHub.Orgs, p) }
            }
        }
        cfg.GitHub.Topic = prompt(r, "Filter by topic (optional)", cfg.GitHub.Topic)
        cfg.GitHub.Prefix = prompt(r, "Filter by name prefix (optional)", cfg.GitHub.Prefix)
        if strings.ToLower(prompt(r, "Include private repos? (y/N)", "n")) == "y" {
            cfg.GitHub.IncludePrivate = true
        } else {
            cfg.GitHub.IncludePrivate = false
        }
        // token is optional; prompt but don't echo back in --show
        cfg.GitHub.Token = prompt(r, "GitHub token (optional, for higher rate limits/private)", cfg.GitHub.Token)
    }

    if err := services.Save(cfg); err != nil {
        return errors.New("failed to save config: " + err.Error())
    }
    helpers.Log.Info().Msg("Configuration saved ✔")
    return nil
}


