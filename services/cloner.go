package services

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
    memorystorage "github.com/go-git/go-git/v5/storage/memory"
)

type Cloner interface {
	CloneRepository(repoURL, outputDir string) error
    CloneRepositoryBranch(repoURL, branch, outputDir string) error
    ListRemoteBranches(repoURL string) ([]string, error)
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

func (gc *GitCloner) CloneRepositoryBranch(repoURL, branch, outputDir string) error {
    cloneOptions := &git.CloneOptions{
        URL:      repoURL,
        Progress: os.Stdout,
    }

    if branch != "" {
        cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(branch)
        cloneOptions.SingleBranch = true
        cloneOptions.Depth = 1
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

// ListRemoteBranches lists remote branch names without cloning locally.
func (gc *GitCloner) ListRemoteBranches(repoURL string) ([]string, error) {
    remote := git.NewRemote(memorystorage.NewStorage(), &config.RemoteConfig{URLs: []string{repoURL}})
    listOpts := &git.ListOptions{}

    if gc.isSSH(repoURL) {
        sshKeyPath, err := gc.getSSHKeyPath()
        if err != nil {
            return nil, fmt.Errorf("failed to get SSH key path: %w", err)
        }
        auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
        if err != nil {
            return nil, fmt.Errorf("failed to create SSH auth method: %v", err)
        }
        listOpts.Auth = auth
    }

    refs, err := remote.List(listOpts)
    if err != nil {
        return nil, err
    }
    var branches []string
    seen := map[string]struct{}{}
    for _, r := range refs {
        if r.Name().IsBranch() {
            name := r.Name().Short()
            if _, ok := seen[name]; !ok {
                seen[name] = struct{}{}
                branches = append(branches, name)
            }
        }
    }
    return branches, nil
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
