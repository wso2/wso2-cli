package set

type CtxSetOpts struct {
	Project      string
	Org          string
	ShowMessages bool
}

type CtxModel struct {
	Project string `yaml:"project"`
	Org     string `yaml:"org"`
}
