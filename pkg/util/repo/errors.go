package repo

import "errors"

var (
	ErrGitCredentialsNotFound = errors.New("git credentials not found")
	ErrGitRepoNotInitialized  = errors.New("git repository not initialized")
	ErrCannotCreateTargetDir  = errors.New("cannot create target directory")
	ErrNotWithinGitRepo       = errors.New("not within a git repo")
)
