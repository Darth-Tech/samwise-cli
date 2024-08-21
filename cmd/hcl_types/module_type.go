package hcltypes

type ModuleConfig struct {
	Source  string `hcl:"source,label"`
	Version string `hcl:"version,label"`
}
type Module struct {
	Module ModuleConfig `hcl:"service,block"`
}
