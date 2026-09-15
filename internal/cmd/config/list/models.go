package list

type ConfigListOptions struct {
	orgFlag             string
	projectFlag         string
	componentFlag       string
	envFlag             string
	deploymentTrackFlag string
	outputFlag          string // --output / -o: output format (table or json)
}
