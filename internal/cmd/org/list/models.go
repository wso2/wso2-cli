package list

// OrgListOptions holds the flag values for the `list organizations` command.
// cobra binds each flag to a field here, and the handler reads them.
type OrgListOptions struct {
	outputFlag string
}
