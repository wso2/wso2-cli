package delete

type ConfigDeleteOptions struct {
	orgFlag             string
	projectFlag         string
	componentFlag       string
	envFlag             string
	deploymentTrackFlag string
	nameFlag            string
	skipConfirm         bool
}
