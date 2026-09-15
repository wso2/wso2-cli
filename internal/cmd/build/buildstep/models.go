package buildstep

type BuildStep int

const (
	CHECKOV BuildStep = iota
	TRIVY
	TRIVY_LIBRARY
	TRIVY_DROPINS
	INTEGRATION_PROJECT_BUILD
	DOCKER_BUILD
	PACK_BUILD
	BALLERINA_BUILD
	POST_BUILD_CHECK
	MI_VERSION_VALIDATION
	MAIN_SEQUENCE_VALIDATION
	GIT_CHECKOUT
	SOURCE_CONFIGURATION_FILE_VALIDATION
	OAS_VALIDATION
	UPDATE_API
	API_GOVERNANCE
	READ_COMOPNENT_CONFIG
)

// humanReadableNames maps human-readable names to BuildStep constants.
var humanReadableNames = map[string]BuildStep{
	"Dockerfile scan":                      CHECKOV,
	"Container (Trivy) vulnerability scan": TRIVY,
	"Library (Trivy) vulnerability scan":   TRIVY_LIBRARY,
	"Dropins (Trivy) vulnerability scan":   TRIVY_DROPINS,
	"Integration Project Build":            INTEGRATION_PROJECT_BUILD,
	"Docker Build":                         DOCKER_BUILD,
	"Build Component":                      PACK_BUILD,
	"Ballerina Build":                      BALLERINA_BUILD,
	"Post-Build Check":                     POST_BUILD_CHECK,
	"MI Version Validation":                MI_VERSION_VALIDATION,
	"Main Sequence Validation":             MAIN_SEQUENCE_VALIDATION,
	"Checkout Source Code":                 GIT_CHECKOUT,
	"Source Configuration File Validation": SOURCE_CONFIGURATION_FILE_VALIDATION,
	"Validate OAS":                         OAS_VALIDATION,
	"Update API":                           UPDATE_API,
	"Validate Against Governance Rules":    API_GOVERNANCE,
	"Read Component Yaml":                  READ_COMOPNENT_CONFIG,
}

// shortFormNames maps short form names to BuildStep constants.
var shortFormNames = map[string]BuildStep{
	"checkov":                   CHECKOV,
	"trivy":                     TRIVY,
	"trivy_library":             TRIVY_LIBRARY,
	"trivy_dropins":             TRIVY_DROPINS,
	"integration_project_build": INTEGRATION_PROJECT_BUILD,
	"docker_build":              DOCKER_BUILD,
	"build":                     PACK_BUILD,
	"ballerina_build":           BALLERINA_BUILD,
	"post_build_check":          POST_BUILD_CHECK,
	"mi_version_val":            MI_VERSION_VALIDATION,
	"main_sequence_val":         MAIN_SEQUENCE_VALIDATION,
	"git_checkout":              GIT_CHECKOUT,
	"src_config_val":            SOURCE_CONFIGURATION_FILE_VALIDATION,
	"oas_validation":            OAS_VALIDATION,
	"update_api":                UPDATE_API,
	"api_governance":            API_GOVERNANCE,
	"read_comp_config":          READ_COMOPNENT_CONFIG,
}

// logFetchKeys maps BuildStep constants to their corresponding log fetch keys.
var logFetchKeys = map[string]BuildStep{
	"libraryTrivyReport":      TRIVY_LIBRARY,
	"dropinsTrivyReport":      TRIVY_DROPINS,
	"integrationProjectBuild": INTEGRATION_PROJECT_BUILD,
	"postBuildCheckLogs":      POST_BUILD_CHECK,
	"mainSequenceValidation":  MAIN_SEQUENCE_VALIDATION,
	"mIVersionValidation":     MI_VERSION_VALIDATION,
	"proxyBuildLogs":          OAS_VALIDATION,
	"governanceLogs":          API_GOVERNANCE,
	"configValidationLogs":    READ_COMOPNENT_CONFIG,
}

func ResolveFromLogFetchKey(key string) BuildStep {
	if resp, ok := logFetchKeys[key]; ok {
		return resp
	}

	return -1
}

func ResolveLogFetchKey(step BuildStep) string {
	for k, s := range logFetchKeys {
		if step == s {
			return k
		}
	}
	return "unknown"
}

// ResolveBuildStep resolves a human-readable name to a BuildStep constant.
func ResolveBuildStep(name string) BuildStep {
	if step, exists := humanReadableNames[name]; exists {
		return step
	}
	return -1 // or some other default value or error handling
}

// GetShortFormStepName returns the short form name for a given BuildStep.
func GetShortFormStepName(step BuildStep) string {
	for shortForm, buildStep := range shortFormNames {
		if buildStep == step {
			return shortForm
		}
	}
	return "unknown"
}

// ResolveBuildStepFromShortForm resolves a short form name to a BuildStep constant.
func ResolveBuildStepFromShortForm(shortForm string) BuildStep {
	if step, exists := shortFormNames[shortForm]; exists {
		return step
	}
	return -1 // or some other default value or error handling
}

// HasLogs determines if a build step has logs based on its conclusion.
func HasLogs(step BuildStep, conclusion string) bool {
	switch step {
	case PACK_BUILD, BALLERINA_BUILD, INTEGRATION_PROJECT_BUILD:
		return conclusion != "skipped"
	case CHECKOV, TRIVY, TRIVY_LIBRARY, POST_BUILD_CHECK, TRIVY_DROPINS, MAIN_SEQUENCE_VALIDATION, MI_VERSION_VALIDATION:
		return conclusion == "failure"
	case DOCKER_BUILD:
		return conclusion != "skipped"
	case GIT_CHECKOUT, SOURCE_CONFIGURATION_FILE_VALIDATION:
		return conclusion == "failure"
	case OAS_VALIDATION, API_GOVERNANCE, READ_COMOPNENT_CONFIG:
		return true
	default:
		return false
	}
}
