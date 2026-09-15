package common

type KeyValOpt struct {
	Key string `json:"key"`
	Val string `json:"value"`
}

type KeyMultipleValOpt struct {
	Key string   `json:"key"`
	Val []string `json:"value"`
}

type CreateConfigParams struct {
	ConfigName       string
	ConfigType       string
	MountType        string
	EnvVars          []KeyValOpt
	FileMountPath    string
	FileMountContent string
	OrgId            string
	OrgUuid          string
	ProjectId        string
	EnvId            string
	AppEnvId         string
	ComponentId      string
}

type LinkData struct {
	Project   string `yaml:"project"`
	Org       string `yaml:"org"`
	Component string `yaml:"component"`
}
