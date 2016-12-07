package main

import (
	"bytes"
	"html/template"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Namespace     string       `yaml:"namespace,omitempty"`
	Repositories  []Repository `yaml:"repositories"`
	DefaultBranch string       `yaml:"defaultBranch,omitempty"`
	KubeFolder    string       `yaml:"kubernetesFolder"`
	BaseDir       string       `yaml:"baseDir,omitempty"`
	UpdateRepoVar string       `yaml:"updateRepoVar,omitempty"`
	UpdateRefVar  string       `yaml:"updateRefVar,omitempty"`
}

type Repository struct {
	Name   string `yaml:"name,omitempty"`
	URI    string `yaml:"uri"`
	Commit string `yaml:"commit,omitempty"`
}

func parseConfig(configFile string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("config").Parse(string(configBytes))
	var out bytes.Buffer
	err = tmpl.Execute(&out, envToMap())
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.Unmarshal(out.Bytes(), c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
