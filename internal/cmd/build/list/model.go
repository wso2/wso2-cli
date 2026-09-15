package list

type BuildListFlags struct {
	Component       string
	Project         string
	Org             string
	DeploymentTrack string
	OutputFlag      string // --output / -o: output format (table or json)
}
