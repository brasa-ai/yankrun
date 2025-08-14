package domain

type Config struct {
    StartDelim    string `yaml:"start_delim"`
    EndDelim      string `yaml:"end_delim"`
    FileSizeLimit string `yaml:"file_size_limit"`
    Templates     []TemplateRepo `yaml:"templates"`
    GitHub        GitHubConfig   `yaml:"github"`
    Functions    Functions      `yaml:"functions,omitempty"`
    IgnorePath   []string       `yaml:"ignore_patterns,omitempty"`
}

type TemplateRepo struct {
    Name        string `yaml:"name"`
    URL         string `yaml:"url"`
    Description string `yaml:"description"`
    DefaultBranch string `yaml:"default_branch"`
}

type GitHubConfig struct {
    User           string   `yaml:"user"`
    Orgs           []string `yaml:"orgs"`
    Token          string   `yaml:"token"`
    Topic          string   `yaml:"topic"`   // optional: required topic in repo
    Prefix         string   `yaml:"prefix"`  // optional: repo name prefix filter
    IncludePrivate bool     `yaml:"include_private"`
}


