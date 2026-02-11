// Package storage handles everything regarding the persistence layer
package storage

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Request struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
	Auth    *BasicAuth        `yaml:"auth,omitempty"`
}

type Collection struct {
	Name     string    `yaml:"name"`
	BaseURL  string    `yaml:"base_url"`
	Requests []Request `yaml:"requests"`
}

type Environment struct {
	Name      string            `yaml:"name"`
	Variables map[string]string `yaml:"variables"`
}
