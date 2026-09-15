package project

import (
	"fmt"
)

type GetProjectMutationRequest struct {
	ID           string
	Name         string
	Description  string
	OrgID        int
	OrgHandler   string
	Version      string
	Region       string
	Repository   string
	Branch       string
	CredentialId string
}

func wrapMutation(mutation string) string {
	return `{"query": "` + mutation + `"}`
}

func GetCreateMonoRepoProjectMutation(req GetProjectMutationRequest) string {
	mutation := fmt.Sprintf(`mutation{ createProject(project: { name: \"%s\", description: \"%s\", orgId: %d, orgHandler: \"%s\",version: \"%s\", region: \"%s\", repository: \"%s\", branch: \"%s\", secretRef: \"%s\" }){ id, orgId, name, version, createdDate, handler, region, description, deploymentPipelineId}}`,
		req.Name,
		req.Description,
		req.OrgID,
		req.OrgHandler,
		req.Version,
		req.Region,
		req.Repository,
		req.Branch,
		req.CredentialId)
	return wrapMutation(mutation)
}

func GetCreateMultiRepoProjectMutation(req GetProjectMutationRequest) string {
	mutation := fmt.Sprintf(`mutation{ createProject(project: { name: \"%s\", description: \"%s\", orgId: %d, orgHandler: \"%s\",version: \"%s\", region: \"%s\" }){ id, orgId, name, version, createdDate, handler, region, description, deploymentPipelineId}}`,
		req.Name,
		req.Description,
		req.OrgID,
		req.OrgHandler,
		req.Version,
		req.Region)
	return wrapMutation(mutation)
}

func GetUpdateProjectMutation(req GetProjectMutationRequest) string {
	mutation := fmt.Sprintf(`mutation{ updateProject(project: { id: \"%s\", name: \"%s\", description: \"%s\",orgId: %d, version: \"%s\" }){ id, name, description, orgId, version, handler, createdDate } }`,
		req.ID,
		req.Name,
		req.Description,
		req.OrgID,
		req.Version)

	return wrapMutation(mutation)
}
