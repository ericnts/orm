package orm

type Options struct {
	Master *EntryOptions `yaml:"master"`
	Slave  *EntryOptions `yaml:"slave"`
}

type EntryOptions struct {
	Dialector string `yaml:"dialector"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Host      string `yaml:"host"`
	Path      string `yaml:"path"`
	RawQuery  string `yaml:"rawQuery"`
	MaxOpen   int    `yaml:"maxOpen"`
	MaxIdle   int    `yaml:"maxIdle"`
}
