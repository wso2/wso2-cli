package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type GraphQLClient struct {
	Endpoint string
	client   *api.IPHTTPClient
}

func NewGraphQLClient(endpoint string, tokenStore api.ReadOnlyTokenStore) *GraphQLClient {
	return &GraphQLClient{
		Endpoint: endpoint,
		client:   api.NewIPHTTPClient(tokenStore),
	}
}

// Define the shared struct
type CreateComponentResponseData struct {
	ID              string `json:"id"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	Name            string `json:"name"`
	Handle          string `json:"handle"`
	OrganizationID  string `json:"organizationId"`
	ProjectID       string `json:"projectId"`
	OrgHandle       string `json:"orgHandle"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	ImageRegistryID string `json:"imageRegistryId"`
	ImageRegistry   struct {
		ID                  string `json:"id"`
		CreatedAt           string `json:"createdAt"`
		UpdatedAt           string `json:"updatedAt"`
		CloudConnectorID    string `json:"cloudConnectorId"`
		ImageRepositoryName string `json:"imageRepositoryName"`
	} `json:"imageRegistry"`
	ComponentType string `json:"componentType"`
	HTTPBased     bool   `json:"httpBased"`
}

func (c *GraphQLClient) CreateSPAWebAppComponent(
	orgId string,
	orgHandler string,
	projectId string,
	isPublicRepo bool,
	reqBody ComponentKind) (*SPAWebAppComponentResponse, error) {

	q, err := c.getCreateSPAWebAppComponentQuery(
		reqBody.Metadata.Name,
		reqBody.Metadata.DisplayName,
		orgId,
		orgHandler,
		projectId,
		reqBody.Spec.Source.Github.Path,
		reqBody.Spec.Source.Github.Repository,
		reqBody.Spec.Source.Github.Branch,
		reqBody.Spec.Build.Webapp.Type,
		reqBody.Spec.Build.Webapp.BuildCommand,
		reqBody.Spec.Build.Webapp.NodeVersion,
		reqBody.Spec.Build.Webapp.OutputDir,
		isPublicRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating query: %w", err)
	}
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(q)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while creating component: %w, query: %s", err, q)
	}
	defer resp.Body.Close()

	var response SPAWebAppComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil

}

// Use the shared struct in your response types
type SPAWebAppComponentResponse struct {
	Data struct {
		CreateByocComponent CreateComponentResponseData `json:"createByocComponent"`
	} `json:"data"`
}

func (c *GraphQLClient) getCreateSPAWebAppComponentQuery(
	name string,
	displayName string,
	orgId string,
	orgHandler string,
	projectId string,
	dockerContext string,
	srcGitRepoUrl string,
	srcGitRepoBranch string,
	webAppType string,
	webAppBuildCommand string,
	webAppPackageManagerVersion string,
	webAppOutputDirectory string,
	isPublicRepo bool,
) (string, error) {
	orgIdInt, err := strconv.Atoi(orgId)
	if err != nil {
		return "", fmt.Errorf("failed to convert org id to int: %v", err)
	}
	return wrapQuery(fmt.Sprintf(`mutation {
	  createByocComponent(
		component: {
		  name: \"%s\",
		  orgId: %d,
		  orgHandler: \"%s\",
		  projectId: \"%s\",
		  labels: \"\",
		  description: \"\",
		  componentType: \"byocWebAppsDockerfileLess\",
		  displayName: \"%s\",
		  accessibility: \"external\",
		  byocWebAppsConfig: {
			dockerContext:\"%s\",
			srcGitRepoUrl:\"%s\",
			srcGitRepoBranch: \"%s\",
			webAppType: \"%s\",
			webAppBuildCommand: \"%s\",
			webAppPackageManagerVersion: \"%s\",
			webAppOutputDirectory: \"%s\",
			isAppGatewayEnabled: true
		  },
		  secretRef: \"\",
		  isPublicRepo: %t
		}
	  ) 
	  {
		id,
		createdAt,
		updatedAt,
		name,
		handle,
		organizationId,
		projectId,
		orgHandle,
		type,
		description,
		imageRegistryId,
		imageRegistry {
			id,
			createdAt,
			updatedAt,
			cloudConnectorId,
			imageRepositoryName
		},
		componentType,
		httpBased
	  }
	}`,
		name,
		orgIdInt,
		orgHandler,
		projectId,
		displayName,
		dockerContext,
		srcGitRepoUrl,
		srcGitRepoBranch,
		webAppType,
		webAppBuildCommand,
		webAppPackageManagerVersion,
		webAppOutputDirectory,
		isPublicRepo,
	)), nil
}

func (c *GraphQLClient) CreateBuildpackServiceComponent(
	name string,
	displayName string,
	description string,
	orgId string,
	orgHandler string,
	projectId string,
	srcGitRepoUrl string,
	srcGitRepoBranch string,
	srcGitRepoComponentDirectory string,
	languageVersion string,
	buildpackId string,
	isPublicRepo bool,
) (*CreateBuildpackComponentResponse, error) {
	q, err := c.getCreateBuildpackServiceMutation(
		name,
		displayName,
		description,
		orgId,
		orgHandler,
		projectId,
		srcGitRepoUrl,
		srcGitRepoBranch,
		srcGitRepoComponentDirectory,
		languageVersion,
		buildpackId,
		isPublicRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating query: %w", err)
	}
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(q)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while creating component: %w, query: %s", err, q)
	}
	defer resp.Body.Close()

	var response CreateBuildpackComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Use the shared struct in your response types
type CreateBuildpackComponentResponse struct {
	Data struct {
		CreateBuildpackComponent CreateComponentResponseData `json:"createBuildpackComponent"`
	} `json:"data"`
}

func (c *GraphQLClient) getCreateBuildpackServiceMutation(
	name string,
	displayName string,
	description string,
	orgId string,
	orgHandler string,
	projectId string,
	srcGitRepoUrl string,
	srcGitRepoBranch string,
	srcGitRepoComponentDirectory string,
	languageVersion string,
	buildpackId string,
	isPublicRepo bool,
) (string, error) {
	return wrapQuery(fmt.Sprintf(`
			mutation {
				createBuildpackComponent(
					component: {
						name: \"%s\",
						displayName: \"%s\",
						description: \"%s\",
						orgId: %s,
						orgHandler: \"%s\",
						projectId: \"%s\",
						labels: \"\",
						componentType: \"buildpackService\",
						port: null,
						oasFilePath: \"\",
						accessibility: \"external\",
						isAsyncCreationEnabled: true,
						buildpackConfig: {
							srcGitRepoUrl: \"%s\",
							srcGitRepoBranch: \"%s\",
							buildContext: \"%s\",
							languageVersion: \"%s\",
							buildpackId: \"%s\"
						},
						secretRef: \"\",
						isPublicRepo: %t
					}
				) {
					id
					createdAt
					updatedAt
					name
					handle
					organizationId
					projectId
					orgHandle
					type
					description
					imageRegistryId
					imageRegistry {
						id
						createdAt
						updatedAt
						cloudConnectorId
						imageRepositoryName
					}
					componentType
					httpBased
				}
			}`,
		name,
		displayName,
		description,
		orgId,
		orgHandler,
		projectId,
		srcGitRepoUrl,
		srcGitRepoBranch,
		srcGitRepoComponentDirectory,
		languageVersion,
		buildpackId,
		isPublicRepo,
	)), nil
}

func (c *GraphQLClient) CreateBuildpackWebAppComponent(
	orgId string,
	orgHandler string,
	projectId string,
	buildpackId string,
	isPublicRepo bool,
	reqBody ComponentKind,
	runCommand string,
) (*CreateBuildpackComponentResponse, error) {

	q, err := c.getCreateBuildpackComponentQuery(
		reqBody.Metadata.Name,
		reqBody.Metadata.DisplayName,
		"",
		orgId,
		orgHandler,
		projectId,
		"buildpackWebApp",
		reqBody.Spec.Build.Buildpack.Port,
		"",
		"external",
		true,
		reqBody.Spec.Source.Github.Path,
		reqBody.Spec.Source.Github.Repository,
		reqBody.Spec.Source.Github.Branch,
		reqBody.Spec.Build.Buildpack.Version,
		buildpackId,
		runCommand,
		"",
		isPublicRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating buildpack query: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(q)))
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while creating buildpack component: %w, query: %s", err, q)
	}
	defer resp.Body.Close()

	var response CreateBuildpackComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *GraphQLClient) UpdateCodeServer(
	orgId string,
	userId string,
	organizationUuid string,
	projectId string,
	componentId string,
	orgHandle string,
	imageUrl string,
	registryId string,
	sourceCommitHash string,
) error {

	q, err := c.getUpdateCodeServerQuery(
		userId,
		organizationUuid,
		projectId,
		componentId,
		orgHandle,
		imageUrl,
		registryId,
		sourceCommitHash,
	)
	if err != nil {
		return fmt.Errorf("error while creating UpdateCodeServer query: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(q)))
	if err != nil {
		return fmt.Errorf("error while building UpdateCodeServer request: %w", err)
	}

	resp, err := c.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while making UpdateCodeServer request: %w, query: %s", err, q)
	}
	defer resp.Body.Close()

	return nil
}

func (c *GraphQLClient) getCreateBuildpackComponentQuery(
	name string,
	displayName string,
	description string,
	orgId string,
	orgHandler string,
	projectId string,
	componentType string,
	port int,
	oasFilePath string,
	accessibility string,
	isAsyncCreationEnabled bool,
	buildContext string,
	srcGitRepoUrl string,
	srcGitRepoBranch string,
	languageVersion string,
	buildpackId string,
	runCommand string,
	secretRef string,
	isPublicRepo bool,
) (string, error) {

	orgIdInt, err := strconv.Atoi(orgId)
	if err != nil {
		return "", fmt.Errorf("failed to convert org id to int: %v", err)
	}

	query := fmt.Sprintf(`mutation {
  createBuildpackComponent(
	component: {
	  name: \"%s\",
	  displayName: \"%s\",
	  description: \"%s\",
	  orgId: %d,
	  orgHandler: \"%s\",
	  projectId: \"%s\",
	  labels: \"\",
	  componentType: \"%s\",
	  port: %d,
	  oasFilePath: \"%s\",
	  accessibility: \"%s\",
	  isAsyncCreationEnabled: %t,
	  buildpackConfig: {
		buildContext: \"%s\",
		srcGitRepoUrl: \"%s\",
		srcGitRepoBranch: \"%s\",
		languageVersion: \"%s\",
		buildpackId: \"%s\",
		runCommand: \"%s\"
	  },
	  secretRef: \"%s\",
	  isPublicRepo: %t
	}
  ) {
	id,
	createdAt,
	updatedAt,
	name,
	handle,
	organizationId,
	projectId,
	orgHandle,
	type,
	description,
	imageRegistryId,
	imageRegistry {
	  id,
	  createdAt,
	  updatedAt,
	  cloudConnectorId,
	  imageRepositoryName
	},
	componentType,
	httpBased
  }
}`, name, displayName, description, orgIdInt, orgHandler, projectId, componentType, port,
		oasFilePath, accessibility, isAsyncCreationEnabled, buildContext, srcGitRepoUrl,
		srcGitRepoBranch, languageVersion, buildpackId, runCommand, secretRef, isPublicRepo)

	return wrapQuery(query), nil
}
