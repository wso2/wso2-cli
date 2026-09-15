package api

// Read only token store
type ReadOnlyTokenStore interface {
	GetTokenForOrg(orgID string, forceSignIn bool) (string, error)
	GetTokenForActiveOrg() (string, error)
}
