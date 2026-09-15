package testuser

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wso2/integration-platform-tools/internal/mcp/utils"
	userstore_mgt "github.com/wso2/integration-platform-tools/pkg/api/user-store-mgt"
)

type UserResponse struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Groups    string `json:"groups,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email"`
}

func manageTestUsers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, err := utils.GetRequiredStringArgument(request, "action")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to manage test users."), nil
	}

	switch action {
	case "add":
		return addTestUser(ctx, request)
	case "list":
		return getTestUsers(ctx, request)
	default:
		return utils.NewMCPErrorResponse(fmt.Errorf("invalid action. Possible values: add, list"), "Failed to manage test users."), nil
	}
}

func getTestUsers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	envTemplateId, err := utils.GetRequiredStringArgument(request, "environment_template_id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("environment_template_id is required", err), nil
	}

	userStoreMgtClient := utils.GetUserStoreMgtClient(ctx)
	users, err := userStoreMgtClient.GetTestUsers(targetOrg.UUID, targetOrg.ID, envTemplateId)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to retrieve test users."), nil
	}

	if len(users) == 0 {
		nextSteps := []string{
			"To add a new test user to this environment, call `manage_test_users` again with `action`: 'add' and provide user details.",
			"If managed auth is enabled, use `create_connection` to connect frontend to backend and `create_configuration` to mount config.js to the frontend.",
		}
		return utils.NewMCPResponse([]UserResponse{}, "No test users found.", nextSteps)
	}

	var responseUsers []UserResponse
	for _, user := range users {
		responseUsers = append(responseUsers, UserResponse{
			Username:  user.Username,
			Password:  user.Password,
			Groups:    user.Groups,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		})
	}

	nextSteps := []string{"To add a new test user to this environment, call `manage_test_users` again with `action`: 'add' and provide user details."}
	return utils.NewMCPResponse(responseUsers, "Test users retrieved successfully.", nextSteps)
}

func addTestUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if authErr := utils.EnsureAuthenticated(ctx); authErr != nil {
		return authErr, nil
	}

	targetOrg, err := utils.GetTargetOrgByUuid(ctx, request)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get target organization", err), nil
	}

	envTemplateId, err := utils.GetRequiredStringArgument(request, "environment_template_id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("environment_template_id is required", err), nil
	}
	username, err := utils.GetRequiredStringArgument(request, "username")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to add test user."), nil
	}
	password, err := utils.GetRequiredStringArgument(request, "password")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to add test user."), nil
	}
	email, err := utils.GetRequiredStringArgument(request, "email")
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to add test user."), nil
	}

	groups := utils.GetOptionalStringArgument(request, "groups", "")
	firstName := utils.GetOptionalStringArgument(request, "first_name", "")
	lastName := utils.GetOptionalStringArgument(request, "last_name", "")

	user := userstore_mgt.User{
		Username:  username,
		Password:  password,
		Groups:    groups,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	userStoreMgtClient := utils.GetUserStoreMgtClient(ctx)
	err = userStoreMgtClient.AddTestUser(targetOrg.UUID, targetOrg.ID, envTemplateId, user)
	if err != nil {
		return utils.NewMCPErrorResponse(err, "Failed to add test user."), nil
	}

	nextSteps := []string{"A new test user has been created. You can use these credentials to log into web applications deployed in this environment."}
	return utils.NewMCPResponse(map[string]string{"username": username}, fmt.Sprintf("Successfully added test user '%s'", username), nextSteps)
}
