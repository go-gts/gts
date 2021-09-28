package flags

type Section struct {
	Name string `yaml:"name"`
	Body string `yaml:"body"`
}

type Command struct {
	Name string `yaml:"name"`
	Info string `yaml:"info"`
	Desc string `yaml:"desc"`
}

type Manpage struct {
	Name string    `yaml:"name"`
	Info string    `yaml:"info"`
	Chpt int       `yaml:"chpt"`
	Sect []Section `yaml:"sect"`
}

type Program struct {
	Version Version   `yaml:"version"`
	Name    string    `yaml:"name"`
	Info    string    `yaml:"info"`
	Desc    string    `yaml:"desc"`
	Sect    []Section `yaml:"sect"`
	Docs    []Manpage `yaml:"docs"`
}
