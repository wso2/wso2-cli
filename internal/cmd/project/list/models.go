package list

// ListProjectsParams holds the flag values for the `list projects` command.
// cobra binds each flag to a field here, and the handler reads them.
type ListProjectsParams struct {
	orgFlag    string
	outputFlag string
}
