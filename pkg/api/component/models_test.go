package component

import "testing"

func TestIsIntegrationDisplayType(t *testing.T) {
	integrationTypes := []string{
		DisplayTypeService,
		DisplayTypeByocService,
		DisplayTypeBuildpackService,
		DisplayTypeRestApi,
		DisplayTypeScheduledTask,
		DisplayTypeByocCronjob,
		DisplayTypeBuildpackCronJob,
		DisplayTypeMiEventHandler,
		DisplayTypeBallerinaEventHandler,
	}
	for _, displayType := range integrationTypes {
		if !IsIntegrationDisplayType(displayType) {
			t.Errorf("IsIntegrationDisplayType(%q) = false, want true", displayType)
		}
	}

	nonIntegrationTypes := []string{
		DisplayTypeByocWebApp,
		DisplayTypeByoiWebApp,
		DisplayTypeBuildpackWebApp,
		DisplayTypeByocWebAppDockerLess,
		DisplayTypeManualTrigger,
		DisplayTypeWebhook,
		DisplayTypeBuildpackWebhook,
		DisplayTypeProxy,
		DisplayTypeGitProxy,
		DisplayTypeBuildpackTestRunner,
		"unknown-display-type",
		"",
	}
	for _, displayType := range nonIntegrationTypes {
		if IsIntegrationDisplayType(displayType) {
			t.Errorf("IsIntegrationDisplayType(%q) = true, want false", displayType)
		}
	}
}
