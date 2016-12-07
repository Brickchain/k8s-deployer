package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/client/ssh"
)

func getLocalRef(path string) (string, error) {
	repo, err := git.NewFilesystemRepository(path)
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), nil
}

func getLocalRemote(path string) (string, error) {
	repo, err := git.NewFilesystemRepository(path)
	if err != nil {
		return "", err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", err
	}

	return remote.Config().URL, nil
}

func cloneFiles(repoURI string, refName string, fileFunc func(string, string) error) (string, error) {
	repoParts := strings.Split(repoURI, "/")
	repoName := strings.TrimSuffix(repoParts[len(repoParts)-1], ".git")
	repoPath := config.BaseDir + repoName

	os.RemoveAll(repoPath)

	repo, err := git.NewFilesystemRepository(repoPath + "/.git")
	if err != nil {
		return "", err
	}
	auth, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		return "", err
	}

	fmt.Println(repoURI, refName)

	err = repo.Clone(&git.CloneOptions{
		Auth:       auth,
		RemoteName: "origin",
		URL:        repoURI,
	})
	if err != nil {
		return "", err
	}

	iter, err := repo.Commits()
	if err != nil {
		return "", err
	}
	defer iter.Close()

	var commit *git.Commit
	if strings.HasPrefix(refName, "refs/") {
		ref, err := repo.Ref(plumbing.ReferenceName(refName), false)
		if err != nil {
			return "", err
		}
		commit, err = repo.Commit(ref.Hash())
		if err != nil {
			return "", err
		}
	} else {
		iter.ForEach(func(c *git.Commit) error {
			if c.Hash.String() == refName {
				commit = c
				iter.Close()
			}

			return nil
		})
	}
	if commit == nil {
		return "", fmt.Errorf("Could not find commit")
	}

	files, err := commit.Files()
	if err != nil {
		return "", err
	}

	err = files.ForEach(func(f *git.File) error {
		if strings.HasPrefix(f.Name, config.KubeFolder) {
			abs := filepath.Join(repoPath, f.Name)
			dir := filepath.Dir(abs)

			os.MkdirAll(dir, 0777)
			file, err := os.Create(abs)
			if err != nil {
				return err
			}

			defer file.Close()
			r, err := f.Reader()
			if err != nil {
				return err
			}

			defer r.Close()

			if err := file.Chmod(f.Mode); err != nil {
				return err
			}

			_, err = io.Copy(file, r)
			if err != nil {
				return err
			}

			err = fileFunc(repoPath+"/"+f.Name, commit.Hash.String())
			if err != nil {
				return err
			}

			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return commit.Hash.String(), nil

}
