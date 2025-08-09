package domain

type Config struct {
    StartDelim    string `yaml:"start_delim"`
    EndDelim      string `yaml:"end_delim"`
    FileSizeLimit string `yaml:"file_size_limit"`
}


