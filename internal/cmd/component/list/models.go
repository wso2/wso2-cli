package list

// CmpListOptions holds the flag values for the `list components` command.
// cobra binds each flag to a field here, and the handler reads them.
type CmpListOptions struct {
	OrgFlag     string // --org: organization name, ID, or handle
	ProjectFlag string // --project: project name
	OutputFlag  string // --output / -o: output format (table or json)
}
