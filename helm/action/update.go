package action

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Masterminds/vcs"

	"github.com/deis/helm/helm/log"
)

// Update fetches the remote repo into the home directory.
func Update(repo, home string) {
	home, err := filepath.Abs(home)
	if err != nil {
		log.Die("Could not generate absolute path for %q: %s", home, err)
	}

	// Basically, install if this is the first run.
	ensurePrereqs()
	ensureHome(home)
	gitrepo := filepath.Join(home, CachePath)
	git := ensureRepo(repo, gitrepo)

	if err := gitUpdate(git); err != nil {
		log.Die("Failed to update from Git: %s", err)
	}
}

// gitUpdate updates a Git repo.
func gitUpdate(git *vcs.GitRepo) error {
	if err := git.Update(); err != nil {
		return err
	}

	// TODO: We should make this pretty.
	log.Info("Updated")
	return nil
}

// ensurePrereqs verifies that Git and Kubectl are both available.
func ensurePrereqs() {
	if _, err := exec.LookPath("git"); err != nil {
		log.Die("Could not find 'git' on $PATH: %s", err)
	}
	if _, err := exec.LookPath("kubectl"); err != nil {
		log.Die("Could not find 'kubectl' on $PATH: %s", err)
	}
}

// ensureRepo ensures that the repo exists and is checked out.
func ensureRepo(repo, home string) *vcs.GitRepo {
	if err := os.Chdir(home); err != nil {
		log.Die("Could not change to directory %q: %s", home, err)
	}
	git, err := vcs.NewGitRepo(repo, home)
	if err != nil {
		log.Die("Could not get repository %q: %s", repo, err)
	}

	if !git.CheckLocal() {
		log.Info("Cloning repo into %q. Please wait.", home)
		if err := git.Get(); err != nil {
			log.Die("Could not create repository in %q: %s", home, err)
		}
	}

	return git
}

// ensureHome ensures that a HELM_HOME exists.
func ensureHome(home string) {
	if fi, err := os.Stat(home); err != nil {
		log.Info("Creating %s", home)
		for _, p := range helmpaths {
			pp := filepath.Join(home, p)
			if err := os.MkdirAll(pp, 0755); err != nil {
				log.Die("Could not create %q: %s", pp, err)
			}
		}
	} else if !fi.IsDir() {
		log.Die("%s must be a directory.", home)
	}

	if err := os.Chdir(home); err != nil {
		log.Die("Could not change to directory %q: %s", home, err)
	}
}
