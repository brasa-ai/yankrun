package services

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Cloner interface {
	CloneRepository(repoURL, outputDir string) error
}

type GitCloner struct {
	FileSystem FileSystem
}

func (gc *GitCloner) CloneRepository(repoURL, outputDir string) error {
	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	if gc.isSSH(repoURL) {
		sshKeyPath, err := gc.getSSHKeyPath()
		if err != nil {
			return fmt.Errorf("failed to get SSH key path: %w", err)
		}
		auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			return fmt.Errorf("failed to create SSH auth method: %v", err)
		}
		cloneOptions.Auth = auth
	}

	_, err := git.PlainClone(outputDir, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone the repository: %v", err)
	}
	return nil
}

func (gc *GitCloner) isSSH(repoURL string) bool {
	return strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://")
}

func (gc *GitCloner) getSSHKeyPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, ".ssh", "id_rsa"), nil
}
