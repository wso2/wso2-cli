package describe

type ProjectDescribeFlags struct {
	Project    string
	Org        string
	outputFlag string
}

// ProjectDescribeOutput is the JSON shape for `wso2-integration-platform describe project`.
// Kept separate from models.Project so the on-the-wire API type can evolve
// without breaking the CLI's machine-readable contract.
type ProjectDescribeOutput struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	CreatedDate   string                   `json:"createdDate"`
	Handler       string                   `json:"handler"`
	Organization  string                   `json:"organization"`
	Type          string                   `json:"type"`
	RepositoryURL string                   `json:"repositoryUrl,omitempty"`
	Components    []ProjectComponentOutput `json:"components"`
}

type ProjectComponentOutput struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	LastBuildDate string `json:"lastBuildDate"`
}
