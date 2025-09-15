package actions

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/brasa-ai/yankrun/domain"
    "github.com/brasa-ai/yankrun/helpers"
    "github.com/brasa-ai/yankrun/services"
    "github.com/urfave/cli"
)

type GenerateAction struct {
    fs       services.FileSystem
    cloner   services.Cloner
    parser   services.ReplacementParser
    replacer services.Replacer
}

func NewGenerateAction(fs services.FileSystem, cloner services.Cloner, parser services.ReplacementParser, replacer services.Replacer) *GenerateAction {
    return &GenerateAction{fs: fs, cloner: cloner, parser: parser, replacer: replacer}
}

// Execute: choose template repo/branch, clone, remove .git, then optionally prompt and apply replacements
func (a *GenerateAction) Execute(c *cli.Context) error {
    // parse flags first for non-interactive allowance
    interactivePrompt := c.Bool("interactive")
    input := c.String("input")
    startDelim := c.String("startDelim")
    endDelim := c.String("endDelim")
    fileSizeLimit := c.String("fileSizeLimit")
    verbose := c.Bool("verbose")
    outputDir := c.String("outputDir")
    templateFilter := c.String("template")
    branchFlag := c.String("branch")
    processTemplates := c.Bool("processTemplates")

    cfg, err := services.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    if len(cfg.Templates) == 0 && cfg.GitHub.User == "" && len(cfg.GitHub.Orgs) == 0 && templateFilter == "" {
        // Ask minimal discovery setup inline
        r := bufio.NewReader(os.Stdin)
        fmt.Println("No templates configured. Let's set where to search:")
        u := strings.TrimSpace(promptString(r, "GitHub user (optional, Enter to skip)", ""))
        orgsCSV := strings.TrimSpace(promptString(r, "GitHub orgs (comma-separated, optional)", ""))
        var orgs []string
        if orgsCSV != "" {
            for _, p := range strings.Split(orgsCSV, ",") { if s := strings.TrimSpace(p); s != "" { orgs = append(orgs, s) } }
        }
        cfg.GitHub.User = u
        cfg.GitHub.Orgs = orgs
        _ = services.Save(cfg)
    }

    // flags already parsed above

    // Fill defaults from config
    if startDelim == "" && cfg.StartDelim != "" { startDelim = cfg.StartDelim }
    if endDelim == "" && cfg.EndDelim != "" { endDelim = cfg.EndDelim }
    if fileSizeLimit == "" && cfg.FileSizeLimit != "" { fileSizeLimit = cfg.FileSizeLimit }
    if startDelim == "" { startDelim = "[[" }
    if endDelim == "" { endDelim = "]]" }
    if fileSizeLimit == "" { fileSizeLimit = "3 mb" }

    r := bufio.NewReader(os.Stdin)

    // Aggregate configured repos + discovered GitHub repos
    repos := cfg.Templates
    // Allow direct URL via --template for non-interactive shortcut
    if templateFilter != "" && (strings.Contains(templateFilter, "://") || strings.HasPrefix(templateFilter, "git@")) {
        repos = append(repos, domain.TemplateRepo{Name: templateFilter, URL: templateFilter, DefaultBranch: "main"})
    }
    if cfg.GitHub.User != "" || len(cfg.GitHub.Orgs) > 0 {
        ghClient := services.NewGitHubClient()
        found, _ := ghClient.ListRepos(context.Background(), cfg.GitHub)
        for _, fr := range found {
            repos = append(repos, domain.TemplateRepo{
                Name: fr.FullName, URL: fr.SSHURL, Description: fr.Description, DefaultBranch: fr.DefaultBranch,
            })
        }
    }
    if len(repos) == 0 {
        return fmt.Errorf("no templates configured or found")
    }

    // Build filtered set non-interactively first
    var filtered []domain.TemplateRepo
    if templateFilter != "" {
        for _, t := range repos {
            if strings.Contains(strings.ToLower(t.Name), strings.ToLower(templateFilter)) || strings.Contains(strings.ToLower(t.URL), strings.ToLower(templateFilter)) {
                filtered = append(filtered, t)
            }
        }
    } else {
        filtered = repos
    }
    if len(filtered) == 0 {
        return fmt.Errorf("no templates matched filter")
    }

    var chosen domain.TemplateRepo
    if !interactivePrompt && (templateFilter != "" && len(filtered) >= 1) {
        chosen = filtered[0]
    } else {
        helpers.Log.Info().Msg("Available templates:")
        for i, t := range repos { fmt.Printf("  [%d] %s  (%s)\n", i+1, t.Name, t.URL) }
        fmt.Printf("Filter templates by substring (press Enter to skip): ")
        filter, _ := r.ReadString('\n')
        filter = strings.TrimSpace(filter)
        filtered = nil
        if filter == "" { filtered = repos } else {
            for _, t := range repos {
                if strings.Contains(strings.ToLower(t.Name), strings.ToLower(filter)) || strings.Contains(strings.ToLower(t.URL), strings.ToLower(filter)) {
                    filtered = append(filtered, t)
                }
            }
            if len(filtered) == 0 { filtered = repos }
        }
        for i, t := range filtered { fmt.Printf("  [%d] %s  (%s)\n", i+1, t.Name, t.URL) }
        fmt.Printf("Select template [1-%d]: ", len(filtered))
        selStr, _ := r.ReadString('\n')
        selStr = strings.TrimSpace(selStr)
        idx := 0
        if selStr != "" { fmt.Sscanf(selStr, "%d", &idx); idx-- }
        if idx < 0 || idx >= len(filtered) { idx = 0 }
        chosen = filtered[idx]
    }

    // Retrieve branches from remote and allow filtering by substring as user types
    branches, _ := a.cloner.ListRemoteBranches(chosen.URL)
    if len(branches) == 0 && chosen.DefaultBranch != "" {
        branches = []string{chosen.DefaultBranch}
    }
    // Non-interactive branch selection via flag
    var br string
    if !interactivePrompt {
        if branchFlag != "" {
            br = branchFlag
        } else if chosen.DefaultBranch != "" {
            br = chosen.DefaultBranch
        } else {
            br = "main"
        }
    } else {
        fmt.Printf("Type to filter branches (Enter to accept default [%s]): ", chosen.DefaultBranch)
        branchFilter, _ := r.ReadString('\n')
        branchFilter = strings.TrimSpace(branchFilter)
        var candidates []string
        if branchFilter == "" { candidates = branches } else {
            for _, b := range branches {
                if strings.Contains(strings.ToLower(b), strings.ToLower(branchFilter)) { candidates = append(candidates, b) }
            }
            if len(candidates) == 0 { candidates = branches }
        }
        fmt.Println("Available branches:")
        for i, b := range candidates { fmt.Printf("  [%d] %s\n", i+1, b) }
        fmt.Printf("Select branch [1-%d] (Enter for default): ", len(candidates))
        pick, _ := r.ReadString('\n')
        pick = strings.TrimSpace(pick)
        br = chosen.DefaultBranch
        if pick != "" {
            var idx int
            if _, err := fmt.Sscanf(pick, "%d", &idx); err == nil && idx >= 1 && idx <= len(candidates) { br = candidates[idx-1] }
        }
    }

    if outputDir == "" {
        fmt.Printf("Output directory [./new-project]: ")
        out, _ := r.ReadString('\n')
        out = strings.TrimSpace(out)
        if out == "" { out = "./new-project" }
        outputDir = out
    }

    if err := a.fs.EnsureDir(outputDir); err != nil { return err }

    if err := a.cloner.CloneRepositoryBranch(chosen.URL, br, outputDir); err != nil {
        return err
    }
    helpers.Log.Info().Msgf("Cloned %s@%s into %s", chosen.Name, br, outputDir)

    // Remove .git directory to make it a fresh repo
    gitDir := filepath.Join(outputDir, ".git")
    if err := os.RemoveAll(gitDir); err != nil {
        return fmt.Errorf("failed to remove %s: %w", gitDir, err)
    }
    helpers.Log.Info().Msg("Removed .git directory (new repo initialized)")

    // Parse provided values if any
    var provided domain.InputReplacement
    if input != "" {
        provided, err = a.parser.Parse(input)
        if err != nil { return err }
    }

    // Analyze placeholders
    counts, err := a.replacer.AnalyzeDir(outputDir, fileSizeLimit, startDelim, endDelim)
    if err != nil { return err }
    if len(counts) == 0 {
        helpers.Log.Info().Msg("No placeholders found.")
        return nil
    }

    // Build values map
    values := map[string]string{}
    for _, rpl := range provided.Variables { values[rpl.Key] = rpl.Value }

    // Show summary
    keys := make([]string, 0, len(counts))
    for k := range counts { keys = append(keys, k) }
    sort.Strings(keys)
    helpers.Log.Info().Msg("Discovered placeholders:")
    for _, k := range keys {
        v := values[k]
        if v == "" { v = "(unset)" }
        fmt.Printf("  %-24s  matches=%-6d  value=%s\n", k, counts[k], v)
    }

    // Prompt if requested
    if interactivePrompt {
        for _, k := range keys {
            def := values[k]
            fmt.Printf("Enter value for %s [%s]: ", k, def)
            s, _ := r.ReadString('\n')
            s = strings.TrimSpace(s)
            if s != "" { values[k] = s }
        }
        fmt.Println()
    }

    // Build final replacements
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

    if err := a.replacer.ReplaceInDir(outputDir, final, fileSizeLimit, startDelim, endDelim, verbose); err != nil {
        return err
    }
    
    // Process .tpl files if requested
    if processTemplates {
        if err := a.replacer.ProcessTemplateFiles(outputDir, final, fileSizeLimit, startDelim, endDelim, verbose); err != nil {
            return err
        }
        helpers.Log.Info().Msg("Template file processing complete ✔")
    }
    
    helpers.Log.Info().Msg("Templating complete ✔")
    return nil
}

// promptString reads a line with a label and returns the trimmed value or default when empty
func promptString(r *bufio.Reader, label string, def string) string {
    if def == "" {
        fmt.Printf("%s: ", label)
    } else {
        fmt.Printf("%s [%s]: ", label, def)
    }
    if v, err := r.ReadString('\n'); err == nil {
        if s := strings.TrimSpace(v); s != "" { return s }
    }
    return def
}


