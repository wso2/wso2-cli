package customDomain

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

func getCustomDomains(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve custom domains."), nil
	}

	customDomains, err := utils.GetCustomDomainClient(ctx).GetCustomDomains(
		targetOrg.ID,
	)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve custom domains."), nil
	}

	nextSteps := []string{
		"Check if your domain is already registered for your target environment and component type.",
		"If not found, register it using register_custom_domain.",
	}
	return utils.NewMCPResponse(customDomains, "Custom domains retrieved successfully.", nextSteps)
}

func registerCustomDomain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}

	domainName, err := utils.GetRequiredStringArgument(request, "domain_name")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}
	domainName = strings.TrimPrefix(domainName, "http://")
	domainName = strings.TrimPrefix(domainName, "https://")
	if strings.Contains(domainName, "/") {
		domainName = strings.Split(domainName, "/")[0]
	}

	componentType, err := utils.GetRequiredStringArgument(request, "component_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}

	envId, err := utils.GetRequiredStringArgument(request, "env_template_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}

	tlsProvider, err := utils.GetRequiredStringArgument(request, "tls_provider")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}

	if tlsProvider == "custom" {
		domainRegistrationUrl := fmt.Sprintf("%s/organizations/%s/settings/urlsettings/newdomain", utils.GetChoreConsoleBaseUrl(), targetOrg.Handle)
		return utils.NewMCPErrorResponse(
			fmt.Errorf("domain registration with custom TLS certificates is not supported through the MCP server for security reasons. Please register your custom domain using the following link: %s", domainRegistrationUrl),
			"Failed to register custom domain."), nil
	}

	customDomainClient := utils.GetCustomDomainClient(ctx)

	domainValidity, err := customDomainClient.ValidateCustomDomainForRegistration(targetOrg.ID, domainName, componentType, envId)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}

	if !domainValidity.IsValid {
		if strings.Contains(domainValidity.Message, "CNAME") || strings.Contains(domainValidity.Message, "DNS") {
			cnameVal, err := utils.GetPlatformHostname(componentType)
			if err != nil {
				return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
			}

			dnsInstructions := fmt.Sprintf("DNS Setup Required:\n- Domain Name: %s\n- Record Type: CNAME\n- Target Value: %s\nAfter creating this DNS record with your provider, please wait for DNS propagation and try again",
				domainName, cnameVal)
			return utils.NewMCPErrorResponse(
				fmt.Errorf("%s\n%s", domainValidity.Message, dnsInstructions),
				"Failed to register custom domain."), nil
		}

		return utils.NewMCPErrorResponse(
			fmt.Errorf("%s", domainValidity.Message), "Failed to register custom domain."), nil
	}

	enableAutoApply := false

	customDomain, err := customDomainClient.RegisterCustomDomain(targetOrg.ID, domainName, componentType, envId, tlsProvider, enableAutoApply)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to register custom domain."), nil
	}

	nextSteps := []string{
		fmt.Sprintf("Custom domain has been registered successfully. To create a url mapping for a component with it call create_url_mapping tool with domain_id: '%s'.", customDomain.ID),
	}

	return utils.NewMCPResponse(customDomain, "Custom domain registered successfully.", nextSteps)
}

func createUrlMapping(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	domainId, err := utils.GetRequiredStringArgument(request, "domain_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	componentUuid, err := utils.GetRequiredStringArgument(request, "integration_uuid")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	apiId, err := utils.GetRequiredStringArgument(request, "api_id")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	defaultDomain, err := utils.GetRequiredStringArgument(request, "default_domain")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}
	defaultDomain = strings.TrimPrefix(defaultDomain, "https://")

	component_type, err := utils.GetRequiredStringArgument(request, "component_type")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	customPath, defaultPath := "/", "/"
	if component_type == "api" {
		customPath = utils.GetOptionalStringArgument(request, "custom_path", "/")
		defaultPath = utils.GetOptionalStringArgument(request, "default_path", "/")
	}
	if !strings.HasPrefix(customPath, "/") {
		customPath = "/" + customPath
	}
	if !strings.HasPrefix(defaultPath, "/") {
		defaultPath = "/" + defaultPath
	}
	if customPath != "/" {
		customPath = strings.TrimSuffix(customPath, "/")
	}
	if defaultPath != "/" {
		defaultPath = strings.TrimSuffix(defaultPath, "/")
	}

	workflowMgtClient := utils.GetWorkflowMgtClient(ctx)
	workflowMgtEnabled, err := workflowMgtClient.IsWorkflowEnabled(targetOrg.ID, workflowmgt.URL_CUSTOMIZATION_WORKFLOW_ID)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	metaData := &workflowmgt.RequestUrlCustomizationMetadata{}

	if workflowMgtEnabled {
		targetProject, err := utils.GetTargetProject(ctx, *targetOrg, request)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
		}

		targetComponent, err := utils.GetTargetComponent(ctx, *targetOrg, *targetProject, request)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
		}

		targetEnv, err := utils.GetTargetEnv(ctx, *targetOrg, *targetProject, request)
		if err != nil {
			return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
		}

		metaData.ProjectName = targetProject.Name
		metaData.ComponentName = targetComponent.Name
		metaData.EnvironmentName = targetEnv.Name
		metaData.CustomUrl = strings.TrimPrefix(customPath, "/")
		metaData.Type = component_type
		metaData.ProjectId = targetProject.ID
	}

	urlMapping, err := utils.GetCustomDomainClient(ctx).CreateUrlMapping(targetOrg.ID, domainId, componentUuid, apiId, defaultDomain, customPath, defaultPath, workflowMgtEnabled, metaData)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to create URL mapping."), nil
	}

	if workflowMgtEnabled {
		wfRequestUrl := fmt.Sprintf("%s/organizations/%s/approvals", utils.GetChoreConsoleBaseUrl(), targetOrg.Handle)

		nextSteps := []string{
			"Your organization requires approvals to map custom domains with components. URL mapping request has been submitted.",
			fmt.Sprintf("A reviewer with workflow management permissions should approve the request through following link: %s", wfRequestUrl),
		}
		return utils.NewMCPResponse(nil, "URL mapping request submitted successfully.", nextSteps)
	}
	nextSteps := []string{
		"Custom domain was successfully mapped to your component. Component is now accessible through your custom domain.",
	}
	return utils.NewMCPResponse(urlMapping, "URL mapping created successfully.", nextSteps)
}
