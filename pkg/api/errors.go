package api

import "errors"

var (
	ErrNotLoggedIn              = errors.New("not logged in")
	ErrTokenNotValid            = errors.New("token not valid")
	ErrNoTokenFoundForOrg       = errors.New("token not found for the organization")
	ErrForbidden                = errors.New("forbidden")
	ErrNotFound                 = errors.New("not found")
	ErrNotWithinProject         = errors.New("not within a valid project")
	ErrRefreshToken             = errors.New("error while exchanging the refresh token")
	ErrFailedToResolveComp      = errors.New("failed to resolve component")
	ErrFailedToResolveProj      = errors.New("failed to resolve project")
	ErrProjectNotSynced         = errors.New("project is not synced with WSO2 Integration Platform")
	ErrNotWithinGitRepo         = errors.New("not within a git repo")
	ErrMaxProjectCountReached   = errors.New("maximum number of projects reached within the tier")
	ErrMaxComponentCountReached = errors.New("maximum number of components reached within the tier")
	RepoAccessNeeded            = errors.New("repo access is required by WSO2 Integration Platform")
	ComponentYamlNotFound       = errors.New("component.yaml not found")
	UserNotFound                = errors.New("user not found")
	InvalidSubPath              = errors.New("provided subpath is invalid")
	NoOrgsAvailable             = errors.New("no orgs available for the user")
	FuncionalityNotSupported    = errors.New("functionality not supported")
	NoGitCredentialsFound       = errors.New("No Git credentials found")
)
