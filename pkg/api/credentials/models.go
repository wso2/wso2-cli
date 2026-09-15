package credentials

type CredentialEntry struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	CreatedAt        string `json:"createdAt"`
	OrganizationUuid string `json:"organizationUuid"`
	Type             string `json:"type"`
	ReferenceToken   string `json:"referenceToken"`
	ServerUrl        string `json:"serverUrl"`
}
