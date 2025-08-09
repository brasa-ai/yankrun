package services

import (
    "os"
    "path/filepath"

    "github.com/mitchellh/go-homedir"
    "gopkg.in/yaml.v3"

    "yankrun/domain"
)

func configPath() (string, error) {
    home, err := homedir.Dir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".yankrun", "config.yaml"), nil
}

func Load() (*domain.Config, error) {
    cfg := &domain.Config{}
    path, err := configPath()
    if err != nil {
        return cfg, err
    }
    f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0600)
    if err != nil {
        return cfg, err
    }
    defer f.Close()
    _ = yaml.NewDecoder(f).Decode(cfg)
    return cfg, nil
}

func Save(cfg *domain.Config) error {
    path, err := configPath()
    if err != nil {
        return err
    }
    if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
        return err
    }
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    enc := yaml.NewEncoder(f)
    enc.SetIndent(2)
    return enc.Encode(cfg)
}


