package list

type ListExecutionParams struct {
	Project         string
	Org             string
	Component       string
	Env             string
	DeploymentTrack string
	Offset          int
	Limit           int
	OutputFlag      string // --output / -o: output format (table or json)
}
