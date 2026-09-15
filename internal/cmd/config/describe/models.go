package describe

type DescribeConfigParams struct {
	componentFlag       string
	projectFlag         string
	orgFlag             string
	envFlag             string
	deploymentTrackFlag string
	nameFlag            string
	outputFlag          string
}

// ConfigDescribeOutput is the JSON shape for `wso2-integration-platform describe config`.
// Kept separate from devops.ConfigItem so the on-the-wire API type can evolve
// without breaking the CLI's machine-readable contract.
//
// Secret values are NEVER included: the text path already redacts them (see
// printConfigDetails), and the JSON path must not be a way around that. For a
// secret, Data holds the key names with redacted values, mirroring the table.
type ConfigDescribeOutput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`       // "config-map" or "secret"
	ConfigType string `json:"configType"` // "File" or env-var style
	Version    int    `json:"version"`
	MountPath  string `json:"mountPath,omitempty"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	IsSecret   bool   `json:"isSecret"`
	// Data carries config-map values verbatim; for secrets it carries the keys
	// with their values redacted, never the secret material itself.
	Data map[string]string `json:"data"`
}
