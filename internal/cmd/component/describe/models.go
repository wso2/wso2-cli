package describe

type DescribeComponentParams struct {
	componentFlag string
	projectFlag   string
	orgFlag       string
	outputFlag    string
}

// ComponentDescribeOutput is the JSON shape for `wso2-integration-platform describe component`.
// Kept separate from the on-the-wire API types (models.Component etc.) so the
// CLI's machine-readable contract can stay stable while those evolve. The
// fields mirror what the text path prints, section for section.
type ComponentDescribeOutput struct {
	ID           string                        `json:"id"`
	Name         string                        `json:"name"`
	Handle       string                        `json:"handle"`
	Type         string                        `json:"type"`
	Buildpack    string                        `json:"buildpack"`
	Version      string                        `json:"version"`
	Organization string                        `json:"organization"`
	Project      string                        `json:"project"`
	Repository   *RepositoryOutput             `json:"repository,omitempty"`
	Environments []EnvironmentDeploymentOutput `json:"environments"`
}

// RepositoryOutput mirrors the "Repository Details" text section, which the
// text path only prints when a repo URL is resolved. Omitted (nil) otherwise.
type RepositoryOutput struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
}

// EnvironmentDeploymentOutput is one entry per project environment. The same
// struct covers both the regular-component and git-proxy deployment paths;
// fields that don't apply to a given path are simply left empty/omitted.
type EnvironmentDeploymentOutput struct {
	Name             string           `json:"name"`
	Deployed         bool             `json:"deployed"`
	DeploymentStatus string           `json:"deploymentStatus,omitempty"`
	LastDeployed     string           `json:"lastDeployed,omitempty"`
	DeployedURL      string           `json:"deployedUrl,omitempty"`
	Commit           *CommitOutput    `json:"commit,omitempty"`
	Endpoints        []EndpointOutput `json:"endpoints,omitempty"`
}

type CommitOutput struct {
	Committed string `json:"committed,omitempty"`
	Message   string `json:"message"`
	CreatedBy string `json:"createdBy"`
}

type EndpointOutput struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Context    string `json:"context"`
	Visibility string `json:"visibility"`
	ProjectURL string `json:"projectUrl,omitempty"`
	AccessURL  string `json:"accessUrl,omitempty"`
}
