package models

type Project struct {
	ID              string `json:"id" yaml:"id"`
	OrgID           int    `json:"orgId" yaml:"orgId"`
	Name            string `json:"name" yaml:"name"`
	Version         string `json:"version" yaml:"version"`
	CreatedDate     string `json:"createdDate" yaml:"createdDate"`
	Handler         string `json:"handler" yaml:"handler"`
	Region          string `json:"region" yaml:"region"`
	Description     string `json:"description" yaml:"description"`
	Repository      string `json:"repository" yaml:"repository"`
	Branch          string `json:"branch" yaml:"branch"`
	GitProvider     string `json:"gitProvider" yaml:"gitProvider"`
	GitOrganization string `json:"gitOrganization" yaml:"gitOrganization"`
}
