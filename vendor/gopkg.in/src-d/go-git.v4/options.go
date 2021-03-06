package git

import (
	"errors"

	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/client/common"
)

const (
	// DefaultRemoteName name of the default Remote, just like git command
	DefaultRemoteName = "origin"
)

var (
	ErrMissingURL     = errors.New("URL field is required")
	ErrInvalidRefSpec = errors.New("invalid refspec")
)

// CloneOptions describe how a clone should be perform
type CloneOptions struct {
	// The (possibly remote) repository URL to clone from
	URL string
	// Auth credentials, if required, to uses with the remote repository
	Auth common.AuthMethod
	// Name of the remote to be added, by default `origin`
	RemoteName string
	// Remote branch to clone
	ReferenceName plumbing.ReferenceName
	// Fetch only ReferenceName if true
	SingleBranch bool
	// Limit fetching to the specified number of commits
	Depth int
}

// Validate validate the fields and set the default values
func (o *CloneOptions) Validate() error {
	if o.URL == "" {
		return ErrMissingURL
	}

	if o.RemoteName == "" {
		o.RemoteName = DefaultRemoteName
	}

	if o.ReferenceName == "" {
		o.ReferenceName = plumbing.HEAD
	}

	return nil
}

// PullOptions describe how a pull should be perform
type PullOptions struct {
	// Name of the remote to be pulled
	RemoteName string
	// Remote branch to clone
	ReferenceName plumbing.ReferenceName
	// Fetch only ReferenceName if true
	SingleBranch bool
	// Limit fetching to the specified number of commits
	Depth int
}

// Validate validate the fields and set the default values
func (o *PullOptions) Validate() error {
	if o.RemoteName == "" {
		o.RemoteName = DefaultRemoteName
	}

	if o.ReferenceName == "" {
		o.ReferenceName = plumbing.HEAD
	}

	return nil
}

// FetchOptions describe how a fetch should be perform
type FetchOptions struct {
	RefSpecs []config.RefSpec
	// Depth limit fetching to the specified number of commits from the tip of
	// each remote branch history.
	Depth int
}

// Validate validate the fields and set the default values
func (o *FetchOptions) Validate() error {
	for _, r := range o.RefSpecs {
		if !r.IsValid() {
			return ErrInvalidRefSpec
		}
	}

	return nil
}
