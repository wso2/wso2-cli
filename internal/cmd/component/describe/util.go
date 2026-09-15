package describe

import (
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// emitComponentJSON gathers the same data the text path prints (general info,
// repository, and per-environment deployment details) and renders it as one
// JSON object. It is only called when --output=json. UX affordances (spinners
// excluded — those go to stderr and stay) are otherwise unchanged; decorative
// text and section headers are dropped because the payload describes the
// resource, not the CLI's layout.
func emitComponentJSON(cmp *models.Component, org *api.Organization, proj models.Project, envs *[]project.ProjectEnvironment, isProxy bool, format common.OutputFormat) error {
	out := ComponentDescribeOutput{
		ID:           cmp.Id,
		Name:         cmp.DisplayName,
		Handle:       cmp.Handler,
		Type:         component.GetTypeForDisplayType(cmp.DisplayType),
		Buildpack:    getLangForComponent(*cmp),
		Version:      cmp.Version,
		Organization: org.Name,
		Project:      proj.Name,
		Repository:   buildRepositoryOutput(cmp),
		Environments: make([]EnvironmentDeploymentOutput, 0),
	}

	for i := range *envs {
		env := (*envs)[i]
		var envOut EnvironmentDeploymentOutput
		var err error
		if isProxy {
			envOut, err = BuildProxyEnvOutput(org, cmp, &env)
		} else {
			envOut, err = BuildDeploymentEnvOutput(cmp, org, &env)
		}
		if err != nil {
			return err
		}
		out.Environments = append(out.Environments, envOut)
	}

	rendered, err := common.RenderStructured(format, out)
	if err != nil {
		return err
	}
	fmt.Fprintln(utils.IO.Out, rendered)
	return nil
}

// buildRepositoryOutput mirrors the resolution logic in PrintRepoInfo. It
// returns nil when no repository URL can be resolved, matching the text path
// which prints nothing in that case.
func buildRepositoryOutput(cmp *models.Component) *RepositoryOutput {
	gitRepo, branch := "", "N/A"
	if cmp.Repository.NameApp != "" {
		gitRepo = common.GenRepoUrl(cmp.Repository.GitProvider, cmp.Repository.OrganizationApp, cmp.Repository.NameApp, cmp.Repository.ServerUrl)
	} else if cmp.Project.Repository != "" {
		gitRepo = cmp.Project.Repository
	}

	if cmp.Repository.BranchApp != "" {
		branch = cmp.Repository.BranchApp
	} else if cmp.Project.Branch != "" {
		branch = cmp.Project.Branch
	}

	if gitRepo == "" {
		return nil
	}
	return &RepositoryOutput{Repository: gitRepo, Branch: branch}
}

// BuildDeploymentEnvOutput mirrors PrintDeploymentInfoForEnv for the JSON path.
func BuildDeploymentEnvOutput(cmp *models.Component, org *api.Organization, env *project.ProjectEnvironment) (EnvironmentDeploymentOutput, error) {
	out := EnvironmentDeploymentOutput{Name: env.Name}

	var latestVersion models.ApiVersion
	for _, versionItem := range cmp.ApiVersions {
		if versionItem.Latest {
			latestVersion = versionItem
			break
		}
	}

	deploymentStatusSpinner := utils.CreateSpinner(
		fmt.Sprintf(i18n.T(" Fetching deployment status for the %s environment"), env.Name),
		"")
	deploymentStatusSpinner.Start()

	componentDeployment, _ := auth.ComponentClient.GetComponentDeployment(
		org.Handle,
		org.UUID,
		org.ID,
		cmp.Id,
		latestVersion.Id,
		env.ID)

	deploymentStatusSpinner.Stop()

	if componentDeployment == nil {
		out.Deployed = false
		out.DeploymentStatus = "Not Deployed"
		return out, nil
	}

	out.Deployed = true

	deploymentStatus := "Not Deployed"
	switch componentDeployment.DeploymentStatusV2 {
	case "ERROR":
		deploymentStatus = "Runtime error"
	default:
		deploymentStatus = componentDeployment.DeploymentStatusV2
	}
	out.DeploymentStatus = cases.Title(language.AmericanEnglish).String(deploymentStatus)
	out.LastDeployed = fmt.Sprintf("%s ago", utils.GetRelativeTime(componentDeployment.Build.DeployedAt))

	if componentDeployment.Build.Commit.Sha != "" {
		out.Commit = &CommitOutput{
			Committed: fmt.Sprintf("%s ago", utils.GetRelativeTime(componentDeployment.Build.Commit.Author.Date)),
			Message:   componentDeployment.Build.Commit.Message,
			CreatedBy: componentDeployment.Build.Commit.Author.Name,
		}
	}

	if componentDeployment.DeploymentStatusV2 != "ERROR" {
		if componentDeployment.InvokeUrl != "" {
			out.DeployedURL = componentDeployment.InvokeUrl
		}

		// if component type is service or webhook
		if strings.HasSuffix(cmp.DisplayType, "Service") || strings.HasSuffix(cmp.DisplayType, "Webhook") {
			endpointsSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints"), "")
			endpointsSpinner.Start()
			endpoints, err := auth.ProjectClient.GetComponentEndpoints(cmp.Id, latestVersion.Id, org.ID)
			endpointsSpinner.Stop()
			if err != nil {
				return out, fmt.Errorf(i18n.T("failed to get component endpoints: %w"), err)
			}

			for _, endpoint := range *endpoints {
				accessURL := ""
				if endpoint.Visibility == "Organization" {
					accessURL = endpoint.OrganizationURL
				} else if endpoint.Visibility == "Public" {
					accessURL = endpoint.PublicURL
				}

				out.Endpoints = append(out.Endpoints, EndpointOutput{
					Name:       endpoint.DisplayName,
					Port:       endpoint.Port,
					Status:     endpoint.State,
					Type:       endpoint.Type,
					Context:    endpoint.APIContext,
					Visibility: endpoint.Visibility,
					ProjectURL: endpoint.ProjectURL,
					AccessURL:  accessURL,
				})
			}
		}
	}

	return out, nil
}

// BuildProxyEnvOutput mirrors DisplayProxyDeploymentInfo for the JSON path.
func BuildProxyEnvOutput(org *api.Organization, cmp *models.Component, env *project.ProjectEnvironment) (EnvironmentDeploymentOutput, error) {
	out := EnvironmentDeploymentOutput{Name: env.Name}

	var apiV *models.ApiVersion
	for i := range cmp.ApiVersions {
		if cmp.ApiVersions[i].Latest {
			apiV = &cmp.ApiVersions[i]
			break
		}
	}

	if apiV == nil {
		return out, fmt.Errorf("failed to resolve api version.")
	}

	proxyDeploymentSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints"), "")
	proxyDeploymentSpinner.Start()
	depInfo, err := auth.DeploymentBuildClient.GetProxyDeploymentInfo(
		org.ID,
		org.Handle,
		org.UUID,
		cmp.Id,
		apiV.Id,
		env.ID,
	)
	proxyDeploymentSpinner.Stop()

	if err != nil {
		// Text path treats a fetch error here as "No Deployments found".
		out.Deployed = false
		return out, nil
	}

	out.Deployed = true
	out.DeploymentStatus = cases.Title(language.AmericanEnglish).String(depInfo.LifecycleStatus)
	out.LastDeployed = fmt.Sprintf("%s ago", utils.GetRelativeTimeUnix(time.UnixMilli(depInfo.DeployedTime)))
	out.DeployedURL = depInfo.InvokeURL

	return out, nil
}

func PrintGeneralComponentInfo(cmpWithRepoData *models.Component, org *api.Organization, project models.Project) error {
	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			Component details:
			
			ID:              %s
			Name:            %s
			Handle:          %s
			Type:            %s
			Buildpack:       %s
			Version:         %s
			Organization:    %s
			Project:         %s
		`),
		cmpWithRepoData.Id,
		cmpWithRepoData.DisplayName,
		cmpWithRepoData.Handler,
		component.GetTypeForDisplayType(cmpWithRepoData.DisplayType),
		getLangForComponent(*cmpWithRepoData),
		cmpWithRepoData.Version,
		org.Name,
		project.Name,
	))

	return nil
}

func PrintRepoInfo(cmpWithRepoData *models.Component) error {
	gitRepo, branch := "", "N/A"
	if cmpWithRepoData.Repository.NameApp != "" {
		gitRepo = common.GenRepoUrl(cmpWithRepoData.Repository.GitProvider, cmpWithRepoData.Repository.OrganizationApp, cmpWithRepoData.Repository.NameApp, cmpWithRepoData.Repository.ServerUrl)
	} else if cmpWithRepoData.Project.Repository != "" {
		gitRepo = cmpWithRepoData.Project.Repository
	}

	if cmpWithRepoData.Repository.BranchApp != "" {
		branch = cmpWithRepoData.Repository.BranchApp
	} else if cmpWithRepoData.Project.Branch != "" {
		branch = cmpWithRepoData.Project.Branch
	}

	if gitRepo != "" {
		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			Repository Details:
			Repository: %s
			Branch: %s
		`),
			gitRepo,
			branch,
		))
	}

	return nil
}

func PrintDeploymentInfoForEnv(cmp *models.Component, org *api.Organization, env *project.ProjectEnvironment) error {
	var latestVersion models.ApiVersion
	for _, versionItem := range cmp.ApiVersions {
		if versionItem.Latest {
			latestVersion = versionItem
			break
		}
	}

	deploymentStatusSpinner := utils.CreateSpinner(
		fmt.Sprintf(i18n.T(" Fetching deployment status for the %s environment"), env.Name),
		"")
	deploymentStatusSpinner.Start()

	componentDeployment, _ := auth.ComponentClient.GetComponentDeployment(
		org.Handle,
		org.UUID,
		org.ID,
		cmp.Id,
		latestVersion.Id,
		env.ID)

	deploymentStatusSpinner.Stop()

	if componentDeployment != nil {
		deploymentStatus := "Not Deployed"

		switch componentDeployment.DeploymentStatusV2 {
		case "ERROR":
			deploymentStatus = "Runtime error"
		default:
			deploymentStatus = componentDeployment.DeploymentStatusV2
		}

		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

				%s
				Deployment status: %s
				Last deployed: %s ago

				`),
			utils.CS.Bold(fmt.Sprintf("%s Environment:", env.Name)),
			cases.Title(language.AmericanEnglish).String(deploymentStatus),
			utils.GetRelativeTime(componentDeployment.Build.DeployedAt),
		))
		if componentDeployment.Build.Commit.Sha != "" {
			utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

					Commit details:
						Committed: %s ago
						Commit message: %s
						Created by: %s

					`),
				utils.GetRelativeTime(componentDeployment.Build.Commit.Author.Date),
				componentDeployment.Build.Commit.Message,
				componentDeployment.Build.Commit.Author.Name,
			))
		}

		if componentDeployment.DeploymentStatusV2 != "ERROR" {
			if componentDeployment.InvokeUrl != "" {
				utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
					Deployed URL: %s
				`),
					componentDeployment.InvokeUrl,
				))
			}

			// if component type is service or webhook
			if strings.HasSuffix(cmp.DisplayType, "Service") || strings.HasSuffix(cmp.DisplayType, "Webhook") {
				endpointsEndpoints := utils.CreateSpinner(i18n.T(" Fetching endpoints"), "")
				endpointsEndpoints.Start()
				endpoints, err := auth.ProjectClient.GetComponentEndpoints(cmp.Id, latestVersion.Id, org.ID)
				endpointsEndpoints.Stop()
				if err != nil {
					return fmt.Errorf(i18n.T("failed to get component endpoints: %w"), err)
				}

				for _, endpoint := range *endpoints {
					endpointUrl := ""
					if endpoint.Visibility == "Organization" {
						endpointUrl = heredoc.Docf(i18n.T(`Organization URL: %s`), endpoint.OrganizationURL)
					} else if endpoint.Visibility == "Public" {
						endpointUrl = heredoc.Docf(i18n.T(`Public URL: %s`), endpoint.PublicURL)
					}

					utils.PrintInfo("%s", heredoc.Docf(i18n.T(`
						Endpoint: %s
							Port: %d
							Status: %s
							Type: %s
							Context: %s
							Visibility: %s
							Project URL: %s
							%s
					`),
						endpoint.DisplayName,
						endpoint.Port,
						endpoint.State,
						endpoint.Type,
						endpoint.APIContext,
						endpoint.Visibility,
						endpoint.ProjectURL,
						endpointUrl,
					))
				}
			}

		}
	} else {
		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

				%s: Not Deployed
			`),
			utils.CS.Bold(fmt.Sprintf("%s Environment", env.Name)),
		))
	}

	return nil
}

func getLangForComponent(comp models.Component) string {
	lang := component.ComponentBuildPackBallerina
	if comp.DisplayType == component.ComponentTypeProxyGH {
		lang = "N/A"
	}

	if len(comp.Repository.BuildPackConfig) > 0 {
		lang = comp.Repository.BuildPackConfig[0].Buildpack.Language
	} else if comp.Repository.ByocWebAppBuildConfig != nil {
		lang = comp.Repository.ByocWebAppBuildConfig.WebAppType
	} else if strings.HasPrefix(comp.DisplayType, "byoi") {
		lang = "Container Image"
	} else if strings.HasPrefix(comp.DisplayType, "byoc") {
		lang = component.ComponentBuildPackDocker
	} else if strings.HasPrefix(comp.DisplayType, "mi") {
		lang = component.ComponentBuildPackMI
	}

	return strings.ToLower(lang)
}

func DisplayProxyDeploymentInfo(org *api.Organization, cmp *models.Component, env *project.ProjectEnvironment) error {
	var apiV *models.ApiVersion

	for _, entry := range cmp.ApiVersions {
		if entry.Latest {
			apiV = &entry
			break
		}
	}

	if apiV == nil {
		return fmt.Errorf("failed to resolve api version.")
	}

	if env != nil {
		proxyDeploymentSpinner := utils.CreateSpinner(i18n.T(" Fetching endpoints"), "")
		proxyDeploymentSpinner.Start()
		depInfo, err := auth.DeploymentBuildClient.GetProxyDeploymentInfo(
			org.ID,
			org.Handle,
			org.UUID,
			cmp.Id,
			apiV.Id,
			env.ID,
		)
		proxyDeploymentSpinner.Stop()

		if err != nil {
			utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

                %s No Deployments found
			`),
				utils.CS.Bold(fmt.Sprintf("%s Environment:", env.Name)),
			))
			return nil
		}

		utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

                %s
                Lifecycle status: %s
                Last deployed: %s ago
                Deployed URL: %s
			`),
			utils.CS.Bold(fmt.Sprintf("%s Environment:", env.Name)),
			cases.Title(language.AmericanEnglish).String(depInfo.LifecycleStatus),
			utils.GetRelativeTimeUnix(time.UnixMilli(depInfo.DeployedTime)),
			depInfo.InvokeURL,
		))

	}

	return nil
}

func displayProxyInsights(org *api.Organization, projectId string, cmp *models.Component) error {
	tNow := time.Now().Format("2006-01-02T15:04:05.000-07:00")
	tFrom := time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05.000-07:00")
	envStr := ""

	var apiV *models.ApiVersion

	for _, entry := range cmp.ApiVersions {
		if entry.Latest {
			apiV = &entry
			break
		}
	}

	if apiV == nil {
		return fmt.Errorf("failed to resolve api version.")
	}

	envData, err := auth.ProjectClient.GetEnvList(org.ID, org.UUID, projectId)
	if err != nil {
		return err
	}

	for i, env := range envData {
		envStr += env.ID
		if i < len(envData)-1 {
			envStr += ","
		}
	}

	// insights, err := auth.DeploymentBuildClient.GetProxyInsights(tFrom, tNow, org.UUID)

	fmt.Println(tNow, tFrom, envStr)
	fmt.Println(envData)

	return nil
}
