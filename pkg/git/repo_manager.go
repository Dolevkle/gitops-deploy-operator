package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
	"path/filepath"
)

// RepoManager Manages Git repository operations.
type RepoManager struct {
	RepoURL string
	Branch  string
	RepoDir string
}

func NewRepoManager(url, branch, name string) *RepoManager {
	return &RepoManager{
		RepoURL: url,
		Branch:  branch,
		RepoDir: filepath.Join("/tmp", name),
	}
}

// CloneOrPull Ensures the local repository is up to date.
func (rm *RepoManager) CloneOrPull() error {
	if _, err := os.Stat(rm.RepoDir); os.IsNotExist(err) {
		return rm.cloneRepo()
	}
	return rm.pullRepo()
}

// Clones the repository to the local directory.
func (rm *RepoManager) cloneRepo() error {
	_, err := git.PlainClone(rm.RepoDir, false, &git.CloneOptions{
		URL:           rm.RepoURL,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", rm.Branch)),
	})
	return err
}

// Pulls the latest changes from the remote repository.
func (rm *RepoManager) pullRepo() error {
	repo, err := git.PlainOpen(rm.RepoDir)
	if err != nil {
		return err
	}
	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}
	return workTree.Pull(&git.PullOptions{RemoteName: "origin"})
}

// GetManifestsPath Constructs the full path to the manifests directory within the repository.
func (rm *RepoManager) GetManifestsPath(subPath string) string {
	return filepath.Join(rm.RepoDir, subPath)
}

func GetRepoPath(name, path string) string {
	return filepath.Join("/tmp", name, path)
}

// DeleteRepo Deletes the local repository clone.
func DeleteRepo(name string) error {
	repoPath := filepath.Join("/tmp", name)
	return os.RemoveAll(repoPath)
}
