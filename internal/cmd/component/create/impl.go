package create

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/huh"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/cmd/context/set"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/component"
	"github.com/wso2/integration-platform-tools/pkg/api/models"
	"github.com/wso2/integration-platform-tools/pkg/util/repo"
)

func HandleComponentCreate(params *CreateComponentParams, buildFlagsGiven bool, buildConfigs *[]string) error {
	err := promptComponentName(params)

	if err != nil {
		return err
	}

	if err = validateComponentName(params.ComponentName, []models.Component{}); err != nil {
		return err
	}

	orgId, projectId, err := common.ResolveContext(params.OrgFlag, params.ProjectFlag)

	if err == nil {
		if params.OrgFlag == "" {
			params.OrgFlag = orgId
		}

		if params.ProjectFlag == "" {
			params.ProjectFlag = projectId
		}
	}

	org, err := common.ResolveTargetOrganization(params.OrgFlag)
	if err != nil {
		return err
	}

	projInfo, err := common.ResolveTargetProject(org, params.ProjectFlag)
	if err != nil {
		return err
	}

	existingComps, err := common.GetComponentsWithSystemCompsForProject(org, projInfo)
	if err != nil {
		utils.HandleErr(err)
	}

	repoRootPath, _, _ := common.GetRepoRootAndSubPath()
	// Ignore error if user not within a git repo

	currUser, err := auth.GetCurrentUser()

	if err != nil {
		return err
	}

	// check if the it's a project by in a different org other than current user's org
	isDifferentOrg := currUser.IDPId != org.Owner.IDPId

	remoteUrl, detectedRepo, err := common.HandleGitRemoteURLs(&repoRootPath, &params.Repo)
	if err != nil {
		return err
	}

	credentialRef := ""

	if !strings.Contains(remoteUrl, "github.com") {
		if strings.Contains(remoteUrl, "bitbucket.org") {
			params.RepoProvider = BIT_BUCKET
		} else {
			// NOTE:
			// if there are more than one Git providers with custom URLS,
			// we might need to prompt from the user and let them choose the provider
			params.RepoProvider = GIT_LAB_SERVER
		}
		creds, err := auth.CredentialClient.GetCommonCredentials(org.ID, org.UUID)
		if err != nil {
			return err
		}

		if params.GitCredName == "" {
			opts := make([]huh.Option[string], 0)

			for _, cred := range creds {
				if cred.Type == params.RepoProvider.String() {
					opts = append(opts, huh.Option[string]{Key: cred.Name, Value: cred.ID})
				}
			}

			if len(opts) == 0 {
				return api.NoGitCredentialsFound
			}

			err = prompt.NewPromptSelectMessage(prompt.PromptSelectOpts[string]{
				Message:     "Select Git credential to be used",
				Description: "",
				Default:     "",
				Options:     opts,
			}, &credentialRef).Prompt()

			if err != nil {
				return err
			}

		} else {
			for _, cred := range creds {
				if cred.Name == params.GitCredName {
					credentialRef = cred.ID
					break
				}
			}

		}

		if credentialRef == "" {
			return errors.New("Failed resolving git credentials")
		}
	}

	params.GitCredRef = credentialRef

	if projInfo.Repository != "" && projInfo.Branch != "" && projInfo.GitOrganization != "" {
		// Mono repo
		params.Repo = projInfo.Repository
		params.RepoBranch = projInfo.Branch
		params.RepoOrg = projInfo.GitOrganization

		if !isDifferentOrg {
			_, _, err := common.HandleGitRepoAuthorization(
				remoteUrl,
				org.ID,
				org.UUID,
				params.GitCredName,
			)

			if err != nil {
				return err
			}
		}
	} else {
		// Multi repo
		if !isDifferentOrg {
			repoOrg, repoName, err := common.HandleGitRepoAuthorization(remoteUrl, org.ID, org.UUID, credentialRef)
			if err != nil {
				return err
			}
			params.Repo = repoName
			params.RepoOrg = repoOrg
		} else {
			params.RepoOrg, params.Repo, err = repo.ParseGitURL(remoteUrl)

			if err != nil {
				return err
			}
		}

		err = common.HandleComponentBranch(
			&params.RepoOrg,
			&params.Repo,
			&params.RepoBranch,
			org.ID,
			credentialRef,
		)
		if err != nil {
			return err
		}
	}

	// get the relative path if it's given if not prompt
	if params.Subpath == "" {

		// TODO: once component from public repo is allowed,
		// we need to pass isPublicRepo(true) as the input to GetRepoStructure
		resp, _ := common.GetRepoStructure(params.RepoOrg, params.Repo, params.RepoBranch, false, org.ID)

		generatedPaths := []string{"."}
		genDirList(resp, &generatedPaths)

		if len(generatedPaths) > 1 {
			err = prompt.NewPromptSelectMessage(
				prompt.PromptSelectOpts[string]{
					Message:     i18n.T("Directory:"),
					Description: i18n.T("Select a directory to create the component"),
					Values:      generatedPaths,
				},
				&params.Subpath,
			).Prompt()

			if err != nil {
				return err
			}
		} else {
			err = prompt.NewPromptInputMessage(
				prompt.PromptInputOpts{
					Message: i18n.T("Directory:"),
					Default: ".",
				},
				&params.Subpath,
			).Prompt()

			if err != nil {
				return err
			}
		}
	}

	// get the component name input
	err = validateComponentName(params.ComponentName, existingComps)
	if err != nil {
		return err
	}

	// get the component type input
	if params.ComponentType == "" {
		err := prompt.NewPromptSelectMessage(
			prompt.PromptSelectOpts[string]{Message: "Type:", Values: component.ComponentTypes},
			&params.ComponentType,
		).Prompt()
		if err != nil {
			return err
		}
	}

	selectedBuildPack, err := resolveBuildPackInput(org, params)
	if err != nil {
		return err
	}
	params.BuildPack = selectedBuildPack.Language

	if params.Subpath == "." &&
		(params.BuildPack == component.ComponentBuildPackBallerina ||
			params.BuildPack == component.ComponentBuildPackMI) {
		params.Subpath = ""
	}

	configFileExists, err := ConfigFileExists(repoRootPath, params)
	if err != nil {
		return err
	}

	configKeys := getConfigKeys(params.ComponentType, params.BuildPack, configFileExists, repoRootPath)

	params.BuildPackConfigs = make(map[string]string)

	if buildFlagsGiven {
		// get the build pack configurations
		for _, item := range *buildConfigs {
			keyValue := strings.Split(item, "=")
			if len(keyValue) != 2 {
				return errors.New("invalid build config entry")
			}

			params.BuildPackConfigs[keyValue[0]] = keyValue[1]
		}

		err = verifyConfigList(configKeys, params)
		if err != nil {
			return err
		}
	} else {
		err = promptBuildConfigurationInput(configKeys, params, &selectedBuildPack)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(utils.IO.Out, heredoc.Docf(
		`Creating a new component in project %s of %s Organization`,
		utils.CS.Bold(projInfo.Name),
		utils.CS.Bold(org.Name),
	))

	err = handleCreateComponent(params, *projInfo, org.ID, org.UUID, remoteUrl)
	if err != nil {
		return err
	}

	utils.PrintInfo(i18n.T("\nComponent '%s' has been successfully created!\n"), params.ComponentName)

	err = genComponentConfigFile(params, repoRootPath, configFileExists)
	if err != nil {
		return err
	}

	if detectedRepo {
		set.HandleSetCtx(set.CtxSetOpts{Project: projInfo.Handler, Org: org.Handle, ShowMessages: false})
	}

	time.Sleep(1 * time.Second)

	utils.PrintInfo("%s", heredoc.Docf(i18n.T(`

			To view details of the created component :
				%s

			To build the component :
				%s

		`),
		fmt.Sprintf(`$ wso2-integration-platform describe component "%s" --project="%s"`, params.ComponentName, projInfo.Name),
		fmt.Sprintf(`$ wso2-integration-platform create build "%s" --project="%s" --deployment-track="%s"`, params.ComponentName, projInfo.Name, params.RepoBranch),
	))

	return nil
}
