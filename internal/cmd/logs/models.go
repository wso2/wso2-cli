package logs

type LogsParams struct {
	typeFlag            string // project, component or build
	componentFlag       string // applicable for component and build logs
	projectFlag         string // applicable for all logs
	orgFlag             string // applicable for all logs
	followFlag          bool   // applicable for component and project logs
	queryFlag           string // applicable for component and project logs
	limitFlag           uint   // applicable for component and project logs
	envFlag             string // applicable for component and project logs
	deploymentTrackFlag string // applicable for all logs
	runIdFlag           int    // applicable for build logs

}
