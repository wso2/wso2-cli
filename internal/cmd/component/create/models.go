package create

import "github.com/wso2/integration-platform-tools/pkg/api/component"

// buildPackConfigs used during interactive component creation for API integrations
const (
	PORT       = "port"
	VISIBILITY = "visibility"
)

type GitProvider int

const (
	GIT_HUB GitProvider = iota + 1
	BIT_BUCKET
	BIT_BUCKET_SERVER
	GIT_LAB_SERVER
)

func (gp GitProvider) String() string {
	return [...]string{"github", "bitbucket", "bitbucket-server", "gitlab-server"}[gp-1]
}

// build pack prompts
var MapBuildPackPrompts = map[string]string{
	PORT:       "Port:",
	VISIBILITY: "Visibility:",
}

// build pack prompt types
var MapBuildPackPromptTypes = map[string]string{
	PORT:       "input",
	VISIBILITY: "select",
}

// build pack prompt options
var MapBuildPackPromptOptions = map[string][]string{
	VISIBILITY: {
		component.ComponentServiceVisibilityPublic,
		component.ComponentServiceVisibilityProject,
		component.ComponentServiceVisibilityOrganization,
	},
}

// build pack prompt optional
var MapBuildPackPromptOptional = map[string]bool{}

// build pack prompt validation
var MapBuildPackPromptValidation = map[string]func(string) error{}

var MapBuildPackDefaultConfigs = map[string]string{
	PORT: "8080",
}

type CreateComponentParams struct {
	// id stuff
	ProjectFlag string
	OrgFlag     string
	// name stuff
	ComponentName    string
	Description      string
	DisplayName      string
	ComponentType    string
	ComponentSubType string
	// multi repo stuff
	RepoOrg      string
	Repo         string
	RepoBranch   string
	Subpath      string
	GitCredName  string
	GitCredRef   string
	RepoProvider GitProvider
	// build pack stuff
	BuildPack        string
	BuildPackConfigs map[string]string
	// origin cloud
	OriginCloud string
	// repo visibility
	IsPublicRepo         bool
	PullLatestSubmodules bool
	// build/deploy automation; nil means "unset", defaults to true in GetComponentKindForCreate
	AutoBuild  *bool
	AutoDeploy *bool
}
