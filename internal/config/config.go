package config

import (
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Base        BaseConfig      `yaml:"base"`
	Layers      []Layer         `yaml:"layers"`
	Build       BuildConfig     `yaml:"build"`
	Artifacts   []string        `yaml:"artifacts,omitempty"`
	Container   ContainerConfig `yaml:"container,omitempty"`
	Cache       CacheConfig     `yaml:"cache,omitempty"`
}

type BaseConfig struct {
	Provider string `yaml:"provider"`
	Machine  string `yaml:"machine"`
	Distro   string `yaml:"distro"`
	Version  string `yaml:"version"`
}

type Layer struct {
	Name   string `yaml:"name"`
	Git    string `yaml:"git,omitempty"`
	Branch string `yaml:"branch,omitempty"`
	Path   string `yaml:"path,omitempty"`
}

type BuildConfig struct {
	Image           string   `yaml:"image,omitempty"`
	Machine         string   `yaml:"machine,omitempty"`
	ExtraPackages   []string `yaml:"extra_packages,omitempty"`
	ParallelMake    int      `yaml:"parallel_make,omitempty"`
	BBNumberThreads int      `yaml:"bb_number_threads,omitempty"`
}

type ContainerConfig struct {
	BaseImage string `yaml:"base_image,omitempty"`
	Memory    string `yaml:"memory,omitempty"`
	CPUCount  int    `yaml:"cpu_count,omitempty"`
}

type CacheConfig struct {
	Downloads string `yaml:"downloads,omitempty"`
	SState    string `yaml:"sstate,omitempty"`
	Retention int    `yaml:"retention,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
