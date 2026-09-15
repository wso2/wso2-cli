package describe

type DescribeConnectionsParams struct {
	projectFlag   string
	orgFlag       string
	componentFlag string
	nameFlag      string
	outputFlag    string
}

// ConnectionDescribeOutput is the JSON shape for `wso2-integration-platform describe connection`.
// Kept separate from connections.Connection so the on-the-wire API type can
// evolve without breaking the CLI's machine-readable contract. The fields
// mirror what the text path prints.
//
// Secret values are NEVER included: the text path redacts ConsumerSecret (see
// GetConnectionConfigString), and the JSON path must not be a way around that.
type ConnectionDescribeOutput struct {
	Name         string                      `json:"name"`
	ConnectingTo string                      `json:"connectingTo"`
	Schema       string                      `json:"schema"`
	Environments []ConnectionEnvConfigOutput `json:"environments"`
}

// ConnectionEnvConfigOutput is one entry per project environment, mirroring the
// per-env config block in the text output. Fields that aren't present for a
// given visibility level (e.g. ConsumerKey/TokenURL only exist for public or
// org-level connections) are simply omitted.
type ConnectionEnvConfigOutput struct {
	Name string `json:"name"`
	// ServiceURLFound is false when the connection has no ServiceURL — usually
	// because the target service isn't deployed. The text path prints
	// "Service URL not found." in that case.
	ServiceURLFound bool   `json:"serviceUrlFound"`
	ServiceURL      string `json:"serviceUrl,omitempty"`
	ConsumerKey     string `json:"consumerKey,omitempty"`
	// ConsumerSecret is always redacted, never the real secret material —
	// matching the text path.
	ConsumerSecret string `json:"consumerSecret,omitempty"`
	TokenURL       string `json:"tokenUrl,omitempty"`
}
