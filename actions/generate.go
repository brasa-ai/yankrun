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
			for _, p := range strings.Split(orgsCSV, ",") {
				if s := strings.TrimSpace(p); s != "" {
					orgs = append(orgs, s)
				}
			}
		}
		cfg.GitHub.User = u
		cfg.GitHub.Orgs = orgs
		_ = services.Save(cfg)
	}

	// flags already parsed above

	// Fill defaults from config
	if startDelim == "" && cfg.StartDelim != "" {
		startDelim = cfg.StartDelim
	}
	if endDelim == "" && cfg.EndDelim != "" {
		endDelim = cfg.EndDelim
	}
	if fileSizeLimit == "" && cfg.FileSizeLimit != "" {
		fileSizeLimit = cfg.FileSizeLimit
	}
	if startDelim == "" {
		startDelim = "[["
	}
	if endDelim == "" {
		endDelim = "]]"
	}
	if fileSizeLimit == "" {
		fileSizeLimit = "3 mb"
	}

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

	// Choose template
	var chosen domain.TemplateRepo
	if len(filtered) == 1 {
		chosen = filtered[0]
	} else {
		// For simplicity, pick the first one if not interactive
		chosen = filtered[0]
	}

	// Choose branch
	br := branchFlag
	if br == "" {
		br = chosen.DefaultBranch
	}
	if br == "" {
		br = "main"
	}

	if outputDir == "" {
		fmt.Printf("Output directory [./new-project]: ")
		out, _ := r.ReadString('\n')
		out = strings.TrimSpace(out)
		if out == "" {
			out = "./new-project"
		}
		outputDir = out
	}

	if err := a.fs.EnsureDir(outputDir); err != nil {
		return err
	}

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
		if err != nil {
			return err
		}
	}

	// Analyze placeholders
	counts, err := a.replacer.AnalyzeDir(outputDir, fileSizeLimit, startDelim, endDelim, cfg)
	if err != nil {
		return err
	}
	if len(counts) == 0 {
		helpers.Log.Info().Msg("No placeholders found.")
		return nil
	}

	// Build values map
	values := map[string]string{}
	for _, rpl := range provided.Variables {
		values[rpl.Key] = rpl.Value
	}

	// Show summary
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	helpers.Log.Info().Msg("Discovered placeholders:")
	for _, k := range keys {
		v := values[k]
		if v == "" {
			v = "(unset)"
		}
		fmt.Printf("  %-24s  matches=%-6d  value=%s\n", k, counts[k], v)
	}

	// Prompt if requested
	if interactivePrompt {
		for _, k := range keys {
			def := values[k]
			fmt.Printf("Enter value for %s [%s]: ", k, def)
			s, _ := r.ReadString('\n')
			s = strings.TrimSpace(s)
			if s != "" {
				values[k] = s
			}
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

	if err := a.replacer.ReplaceInDir(outputDir, final, cfg, fileSizeLimit, startDelim, endDelim, verbose); err != nil {
		return err
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
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return def
}
