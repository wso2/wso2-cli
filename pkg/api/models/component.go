package models

import "fmt"

type Component struct {
	ProjectId         string            `json:"projectId" yaml:"projectId"`
	Id                string            `json:"id" yaml:"id"`
	Description       string            `json:"description" yaml:"description"`
	Name              string            `json:"name" yaml:"name"`
	Handler           string            `json:"handler" yaml:"handler"`
	DisplayName       string            `json:"displayName" yaml:"displayName"`
	DisplayType       string            `json:"displayType" yaml:"displayType"`
	ComponentSubType  string            `json:"componentSubType" yaml:"componentSubType"`
	Version           string            `json:"version" yaml:"version"`
	CreatedAt         string            `json:"createdAt" yaml:"createdAt"`
	LastBuildDate     string            `json:"lastBuildDate" yaml:"lastBuildDate"`
	OrgHandler        string            `json:"orgHandler" yaml:"orgHandler"`
	ApiVersions       []ApiVersion      `json:"apiVersions" yaml:"apiVersions"`
	DeploymentTracks  []DeploymentTrack `json:"deploymentTracks" yaml:"deploymentTracks"`
	UpdatedAt         string            `json:"updatedAt" yaml:"updatedAt"`
	Project           Project           `json:"project" yaml:"project"`
	Repository        ComponentRepo     `json:"repository,omitempty" yaml:"repository,omitempty"`
	OriginCloud       string            `json:"originCloud" yaml:"originCloud"`
	IsSystemComponent string            `json:"isSystemComponent" yaml:"isSystemComponent"`
}

type ComponentRepo struct {
	BitbucketServerUrl    string                          `json:"bitbucketServerUrl"`
	RepoCredRef           string                          `json:"repoCredRef"`
	ServerUrl             string                          `json:"serverUrl"`
	GitProvider           string                          `json:"gitProvider"`
	NameApp               string                          `json:"nameApp"`
	NameConfig            string                          `json:"nameConfig"`
	Branch                string                          `json:"branch"`
	BranchApp             string                          `json:"branchApp"`
	OrganizationApp       string                          `json:"organizationApp"`
	OrganizationConfig    string                          `json:"organizationConfig"`
	IsUserManage          bool                            `json:"isUserManage"`
	AppSubPath            string                          `json:"appSubPath"`
	ByocWebAppBuildConfig *ComponentRepoWebAppBuildConfig `json:"byocWebAppBuildConfig,omitempty"`
	ByocBuildConfig       *ComponentRepoByocBuildConfig   `json:"byocBuildConfig,omitempty"`
	BuildPackConfig       []ComponentRepoBuildPackConfig  `json:"buildpackConfig,omitempty"`
}

type ComponentRepoBuildPackConfig struct {
	VersionId       string `json:"versionId"`
	BuildContext    string `json:"buildContext"`
	LanguageVersion string `json:"languageVersion"`
	Buildpack       struct {
		ID       string `json:"id"`
		Language string `json:"language"`
	} `json:"buildpack"`
}

type ComponentRepoByocBuildConfig struct {
	ID              string `json:"id"`
	IsMainContainer bool   `json:"isMainContainer"`
	ContainerID     string `json:"containerId"`
	ComponentID     string `json:"componentId"`
	RepositoryID    string `json:"repositoryId"`
	DockerContext   string `json:"dockerContext"`
	DockerfilePath  string `json:"dockerfilePath"`
	OASFilePath     string `json:"oasFilePath"`
}

type ComponentRepoWebAppBuildConfig struct {
	ID                    string `json:"id"`
	ContainerID           string `json:"containerId"`
	ComponentID           string `json:"componentId"`
	RepositoryID          string `json:"repositoryId"`
	DockerContext         string `json:"dockerContext"`
	WebAppType            string `json:"webAppType"`
	BuildCommand          string `json:"buildCommand"`
	PackageManagerVersion string `json:"packageManagerVersion"`
	OutputDirectory       string `json:"outputDirectory"`
}

type ApiVersion struct {
	ApiVersion     string           `yaml:"apiVersion" json:"apiVersion"`
	ProxyName      string           `yaml:"proxyName" json:"proxyName"`
	ProxyUrl       string           `yaml:"proxyUrl" json:"proxyUrl"`
	ProxyId        string           `yaml:"proxyId" json:"proxyId"`
	Id             string           `yaml:"id" json:"id"`
	State          string           `yaml:"state" json:"state"`
	Latest         bool             `yaml:"latest" json:"latest"`
	Branch         string           `yaml:"branch" json:"branch"`
	Accessibility  string           `yaml:"accessibility" json:"accessibility"`
	AppEnvVersions []AppEnvVersions `yaml:"appEnvVersions" json:"appEnvVersions"`
	VersionId      string           `yaml:"versionId" json:"versionId"`
}

type AppEnvVersions struct {
	EnvironmentId string `yaml:"environmentId" json:"environmentId"`
	ReleaseId     string `yaml:"releaseId" json:"releaseId"`
}

type DeploymentTrack struct {
	Id              string `yaml:"id" json:"id"`
	CreatedAt       string `yaml:"createdAt" json:"createdAt"`
	UpdatedAt       string `yaml:"updatedAt" json:"updatedAt"`
	ApiVersion      string `yaml:"apiVersion" json:"apiVersion"`
	Branch          string `yaml:"branch" json:"branch"`
	Description     string `yaml:"description" json:"description"`
	ComponentId     string `yaml:"componentId" json:"componentId"`
	Latest          bool   `yaml:"latest" json:"latest"`
	VersionStrategy string `yaml:"versionStrategy" json:"versionStrategy"`
}

// return the all component details as a string
func (c *Component) String() string {
	return fmt.Sprintf("ProjectId: %s\nId: %s\nDescription: %s\nName: %s\nHandler: %s\nDisplayName: %s\nDisplayType: %s\nVersion: %s\nCreatedAt: %s\nLastBuildDate: %s\nOrgHandler: %s\nApiVersions: %v\nDeploymentTracks: %v\n",
		c.ProjectId,
		c.Id,
		c.Description,
		c.Name,
		c.Handler,
		c.DisplayName,
		c.DisplayType,
		c.Version,
		c.CreatedAt,
		c.LastBuildDate,
		c.OrgHandler,
		c.ApiVersions,
		c.DeploymentTracks)
}
