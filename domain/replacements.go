package domain

type Replacement struct {
	Key             string   `json:"key" yaml:"key"`
	Value           string   `json:"value" yaml:"value"`
	BaseKey         string   `json:"-" yaml:"-"` // Not marshalled, used internally
	Transformations []string `json:"-" yaml:"-"` // Not marshalled, used internally
}

type InputReplacement struct {
	Variables  []Replacement `json:"variables" yaml:"variables"`
	IgnorePath []string      `json:"ignore_patterns" yaml:"ignore_patterns"`
}
