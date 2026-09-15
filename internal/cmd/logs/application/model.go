package application

type ApplicationLogsOpts struct {
	org             string
	project         string
	component       string
	deploymentTrack string
	env             string

	// log type specific
	followFlag bool
	queryFlag  string
	limitFlag  uint
}
