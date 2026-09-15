package models

/*
# +required The configuration file schema version
schemaVersion: 1.2

# +optional Incoming connection details for the component
endpoints:
  # +required Unique name for the endpoint.
  # This name will be used when generating the managed API
  - name: greeter-sample
    # +optional Display name for the endpoint.
    displayName: Go Greeter Sample
    # +required Service section has the user service endpoint details
    service:
      # +optional Context (base path) of the API that gets exposed via the endpoint.
      basePath: /greeting-service
      # +required Numeric port value that gets exposed via the endpoint
      port: 9090
    # +required Type of traffic that the endpoint is accepting.
    # Allowed values: REST, GraphQL, WS, GRPC, TCP, UDP.
    type: REST
    # +optional Network level visibilities of the endpoint.
    # Accepted values: Project|Organization|Public(Default).
    networkVisibilities:
      - Public
      - Organization
    # +optional Path to the schema definition file. Defaults to wild card route if not provided
    # This is only applicable to REST or WS endpoint types.
    # The path should be relative to the docker context.
    schemaFilePath: openapi.yaml

# +optional Outgoing connection details for the component.
dependencies:
  # +optional Defines the connection references from the Internal Marketplace.
  connectionReferences:
    # +required Name of the connection.
    - name: hr-connection
      # +required service identifer of the dependent component.
      resourceRef: service:/HRProject/UserComponent/v1/ad088/PUBLIC

# +optional Defines runtime configurations
configurations:
  # +optional List of environment variables to be injected into the component.
  env:
    # +required Name of the environment variable
    - name: HR_SERVICE_URL
      # +required value source
      # Allowed value sources: connectionRef, configForm
      valueFrom:
        # +required Choreo connection value source
        connectionRef:
          # +required Choreo connection name to refer the value from
          name: hr-connection
          # +required Choreo connection configuration key to refer the value from
          key: ServiceURL
    - name: DB_USER
      # +required value source
      # Allowed value sources: connectionRef, configForm
      valueFrom:
        # +required config form value source
        configForm:
            # +optional display name inside the config form, name will be shown in config form if not specified
            displayName: DB User
            # +optional default value is true if not specified
            required: false
            # +optional default value is string if not specified
            # Allowed types - string, number, boolean, secret
            type: string
  # +optional List of files to be injected into the component from config form
  file:
    # +required name of the file
    - name: application.yaml
      # +required path to mount the file at
      mountPath: /src/resources
      # +required file type
      # Supported types - yaml, json and toml
      type: yaml
      # +required define keys of the file
      values:
        # keys of the file
        # +required at least one key
        - name: version
          # +required value source
          # Allowed value sources: connectionRef, configForm
          valueFrom:
            # +required config form value source
            configForm:
              # +optional display name inside the config form, name will be shown in config form if not specified
              displayName: Version
              # +optional default value is string if not specified
              # Allowed types - string, number, boolean, secret, object, array
              type: number

*/

type ComponentYamlConfig struct {
	SchemaVersion  float64                      `yaml:"schemaVersion"`
	Endpoints      *[]ComponentYamlEndpoint     `yaml:"endpoints,omitempty"`
	Dependencies   *ComponentYamlDependencies   `yaml:"dependencies,omitempty"`
	Configurations *ComponentYamlConfigurations `yaml:"configurations,omitempty"`
	Proxy          *ComponentYamlProxyConfig    `yaml:"proxy,omitempty"`
}

type ComponentYamlProxyConfig struct {
	Type           string `yaml:"type"`
	SchemaFilePath string `yaml:"schemaFilePath"`
	Thumbnail      string `yaml:"thumbnailPath,omitempty"`
	DocPath        string `yaml:"docPath,omitempty"`
}

type ComponentYamlEndpoint struct {
	Name                string               `yaml:"name"`
	DisplayName         string               `yaml:"displayName,omitempty"`
	Service             ComponentYamlService `yaml:"service"`
	Type                string               `yaml:"type"`
	NetworkVisibilities []string             `yaml:"networkVisibilities,omitempty"`
	SchemaFilePath      string               `yaml:"schemaFilePath,omitempty"`
}

type ComponentYamlService struct {
	BasePath string `yaml:"basePath,omitempty"`
	Port     int    `yaml:"port"`
}

type ComponentYamlDependencies struct {
	ConnectionReferences []ComponentYamlConnectionReference `yaml:"connectionReferences,omitempty"`
}

type ComponentYamlConnectionReference struct {
	Name        string `yaml:"name"`
	ResourceRef string `yaml:"resourceRef"`
}

type ComponentYamlConfigurations struct {
	Env  *[]ComponentYamlEnvVar `yaml:"env,omitempty"`
	File []ComponentYamlFileVar `yaml:"file,omitempty"`
}

type ComponentYamlEnvVar struct {
	Name      string                 `yaml:"name"`
	ValueFrom ComponentYamlValueFrom `yaml:"valueFrom"`
}

type ComponentYamlValueFrom struct {
	ConnectionRef *ComponentYamConnectionRefConfigForm `yaml:"connectionRef,omitempty"`
	ConfigForm    *ComponentYamlConfigForm             `yaml:"configForm,omitempty"`
}

type ComponentYamConnectionRefConfigForm struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type ComponentYamlConfigForm struct {
	DisplayName string `yaml:"displayName,omitempty"`
	Required    *bool  `yaml:"required,omitempty"`
	Type        string `yaml:"type,omitempty"`
}

type ComponentYamlFileVar struct {
	Name      string                        `yaml:"name"`
	MountPath string                        `yaml:"mountPath"`
	Type      string                        `yaml:"type"`
	Values    []ComponentYamFileValueConfig `yaml:"values"`
}

type ComponentYamFileValueConfig struct {
	Name      string                 `yaml:"name"`
	ValueFrom ComponentYamlValueFrom `yaml:"valueFrom"`
}
