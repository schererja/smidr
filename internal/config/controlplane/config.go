package config

import (
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Name         string       `yaml:"name"`
	Description  string       `yaml:"description"`
	ServerConfig ServerConfig `yaml:"server_config,omitempty"`
}
type ServerConfig struct {
	Port    string `yaml:"port,omitempty"`
	Address string `yaml:"address,omitempty"`
}

// type DirectoryConfig struct {
// 	Downloads string `yaml:"downloads,omitempty"`
// 	SState    string `yaml:"sstate,omitempty"`
// 	Tmp       string `yaml:"tmp,omitempty"`
// 	Deploy    string `yaml:"deploy,omitempty"`
// 	Source    string `yaml:"source,omitempty"`
// 	Build     string `yaml:"build,omitempty"`
// 	Layers    string `yaml:"layers,omitempty"`
// }

// type PackageConfig struct {
// 	Classes string `yaml:"classes,omitempty"`
// }

// type FeatureConfig struct {
// 	ExtraImageFeatures []string `yaml:"extra_image_features,omitempty"`
// 	UserClasses        []string `yaml:"user_classes,omitempty"`
// 	InheritClasses     []string `yaml:"inherit_classes,omitempty"`
// }

// // CacheConfig removed: Directories.* is canonical for all paths.

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return LoadFromBytes(data)
}

// LoadFromBytes parses configuration from YAML or JSON bytes, performing
// environment variable substitution and full validation.
func LoadFromBytes(data []byte) (*Config, error) {
	// Perform environment variable substitution
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 	// Validate the configuration
// 	if err := cfg.Validate(); err != nil {
// 		return nil, fmt.Errorf("configuration validation failed: %w", err)
// 	}

// 	return &cfg, nil
// }

// // Legacy cache normalization removed

// // substituteEnvVars performs environment variable substitution in YAML content
// // Supports formats: ${VAR}, ${VAR:-default}, ${VAR:+alternative}
// // Handles nested environment variables recursively
// func substituteEnvVars(data []byte) []byte {
// 	content := string(data)

// 	// Keep substituting until no more substitutions are found
// 	for {
// 		// Pattern to match ${VAR}, ${VAR:-default}, ${VAR:+alternative}
// 		// This pattern handles nested braces by counting them
// 		envPattern := regexp.MustCompile(`\$\{([^${}]+)\}`)

// 		originalContent := content
// 		content = envPattern.ReplaceAllStringFunc(content, func(match string) string {
// 			// Extract the variable name and any modifier
// 			matches := envPattern.FindStringSubmatch(match)
// 			if len(matches) < 2 {
// 				return match // Return original if pattern doesn't match
// 			}

// 			inner := matches[1] // Everything inside ${...}

// 			// Parse the inner content for modifiers
// 			// Look for :- or :+ patterns
// 			if colonIndex := strings.Index(inner, ":-"); colonIndex != -1 {
// 				// ${VAR:-default} format
// 				varName := strings.TrimSpace(inner[:colonIndex])
// 				defaultValue := strings.TrimSpace(inner[colonIndex+2:])

// 				envValue := os.Getenv(varName)
// 				if envValue == "" {
// 					return defaultValue
// 				}
// 				return envValue
// 			} else if colonIndex := strings.Index(inner, ":+"); colonIndex != -1 {
// 				// ${VAR:+alternative} format
// 				varName := strings.TrimSpace(inner[:colonIndex])
// 				alternativeValue := strings.TrimSpace(inner[colonIndex+2:])

// 				envValue := os.Getenv(varName)
// 				if envValue != "" {
// 					return alternativeValue
// 				}
// 				return ""
// 			} else {
// 				// ${VAR} format
// 				varName := strings.TrimSpace(inner)
// 				return os.Getenv(varName)
// 			}
// 		})

// 		// If no substitutions were made, we're done
// 		if content == originalContent {
// 			break
// 		}
// 	}

// 	return []byte(content)
// }

// // ValidationError represents a configuration validation error
// type ValidationError struct {
// 	Field   string
// 	Message string
// }

// func (e ValidationError) Error() string {
// 	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
// }

// // Validate performs comprehensive validation of the configuration
// func (c *Config) Validate() error {
// 	var errors []error

// 	// Validate required fields
// 	if c.Name == "" {
// 		errors = append(errors, ValidationError{Field: "name", Message: "project name is required"})
// 	}

// 	if c.Description == "" {
// 		errors = append(errors, ValidationError{Field: "description", Message: "project description is required"})
// 	}

// 	// Validate base configuration
// 	if err := c.Base.Validate(); err != nil {
// 		errors = append(errors, err)
// 	}

// 	// Validate layers
// 	if len(c.Layers) == 0 {
// 		errors = append(errors, ValidationError{Field: "layers", Message: "at least one layer is required"})
// 	}

// 	for i, layer := range c.Layers {
// 		if err := layer.Validate(); err != nil {
// 			errors = append(errors, fmt.Errorf("layer[%d]: %w", i, err))
// 		}
// 	}

// 	// Validate build configuration
// 	if err := c.Build.Validate(); err != nil {
// 		errors = append(errors, err)
// 	}

// 	// Validate container settings
// 	if err := c.Container.Validate(); err != nil {
// 		errors = append(errors, err)
// 	}

// 	// Cache validation removed

// 	if len(errors) > 0 {
// 		var errorMessages []string
// 		for _, err := range errors {
// 			errorMessages = append(errorMessages, err.Error())
// 		}
// 		return fmt.Errorf("validation failed:\n%s", strings.Join(errorMessages, "\n"))
// 	}

// 	return nil
// }

// // Validate validates BaseConfig
// func (b *BaseConfig) Validate() error {
// 	if b.Machine == "" {
// 		return ValidationError{Field: "base.machine", Message: "machine is required"}
// 	}

// 	// Validate machine name format (alphanumeric, hyphens, underscores)
// 	machineRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
// 	if !machineRegex.MatchString(b.Machine) {
// 		return ValidationError{Field: "base.machine", Message: "machine name contains invalid characters (only alphanumeric, hyphens, and underscores allowed)"}
// 	}

// 	if b.Distro == "" {
// 		return ValidationError{Field: "base.distro", Message: "distro is required"}
// 	}

// 	return nil
// }

// // Validate validates Layer
// func (l *Layer) Validate() error {
// 	if l.Name == "" {
// 		return ValidationError{Field: "name", Message: "layer name is required"}
// 	}

// 	// Validate layer name format
// 	layerRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
// 	if !layerRegex.MatchString(l.Name) {
// 		return ValidationError{Field: "name", Message: "layer name contains invalid characters (only alphanumeric, hyphens, and underscores allowed)"}
// 	}

// 	// Either git or path must be specified, but not both
// 	hasGit := l.Git != ""
// 	hasPath := l.Path != ""

// 	if !hasGit && !hasPath {
// 		return ValidationError{Field: "git/path", Message: "either git URL or path must be specified"}
// 	}

// 	if hasGit && hasPath {
// 		return ValidationError{Field: "git/path", Message: "cannot specify both git URL and path"}
// 	}

// 	// Validate git URL format if provided
// 	if hasGit {
// 		gitRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(/[a-zA-Z0-9._/-]*)?$`)
// 		if !gitRegex.MatchString(l.Git) {
// 			return ValidationError{Field: "git", Message: "invalid git URL format"}
// 		}
// 	}

// 	return nil
// }

// // Validate validates BuildConfig
// func (b *BuildConfig) Validate() error {
// 	if b.Image == "" {
// 		return ValidationError{Field: "build.image", Message: "image is required"}
// 	}

// 	// Validate image name format
// 	imageRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
// 	if !imageRegex.MatchString(b.Image) {
// 		return ValidationError{Field: "build.image", Message: "image name contains invalid characters"}
// 	}

// 	// Validate parallel settings
// 	if b.ParallelMake < 0 {
// 		return ValidationError{Field: "build.parallel_make", Message: "parallel_make must be non-negative"}
// 	}

// 	if b.BBNumberThreads < 0 {
// 		return ValidationError{Field: "build.bb_number_threads", Message: "bb_number_threads must be non-negative"}
// 	}

// 	// Validate extra packages
// 	for i, pkg := range b.ExtraPackages {
// 		if pkg == "" {
// 			return ValidationError{Field: fmt.Sprintf("build.extra_packages[%d]", i), Message: "package name cannot be empty"}
// 		}
// 		pkgRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
// 		if !pkgRegex.MatchString(pkg) {
// 			return ValidationError{Field: fmt.Sprintf("build.extra_packages[%d]", i), Message: "package name contains invalid characters"}
// 		}
// 	}

// 	return nil
// }

// // Validate validates DirectoryConfig
// func (d *DirectoryConfig) Validate() error {
// 	// Validate directory paths if provided
// 	// Note: After environment variable substitution, paths may contain various characters
// 	// We'll do basic validation but be more permissive for substituted paths
// 	if d.Downloads != "" && !isValidPathOrEnvVar(d.Downloads) {
// 		return ValidationError{Field: "directories.downloads", Message: "invalid directory path format"}
// 	}

// 	if d.SState != "" && !isValidPathOrEnvVar(d.SState) {
// 		return ValidationError{Field: "directories.sstate", Message: "invalid directory path format"}
// 	}

// 	if d.Tmp != "" && !isValidPathOrEnvVar(d.Tmp) {
// 		return ValidationError{Field: "directories.tmp", Message: "invalid directory path format"}
// 	}

// 	if d.Deploy != "" && !isValidPathOrEnvVar(d.Deploy) {
// 		return ValidationError{Field: "directories.deploy", Message: "invalid directory path format"}
// 	}

// 	return nil
// }

// // Validate validates PackageConfig
// func (p *PackageConfig) Validate() error {
// 	if p.Classes != "" {
// 		validClasses := []string{"package_rpm", "package_deb", "package_ipk"}
// 		if !contains(validClasses, p.Classes) {
// 			return ValidationError{Field: "packages.classes", Message: fmt.Sprintf("invalid package class '%s', must be one of: %s", p.Classes, strings.Join(validClasses, ", "))}
// 		}
// 	}
// 	return nil
// }

// // Validate validates FeatureConfig
// func (f *FeatureConfig) Validate() error {
// 	// Validate extra image features
// 	validFeatures := []string{
// 		"debug-tweaks", "package-management", "tools-sdk", "tools-debug",
// 		"tools-profile", "tools-testapps", "dbg-pkgs", "src-pkgs", "dev-pkgs",
// 		"ptest-pkgs", "eclipse-debug",
// 	}

// 	for i, feature := range f.ExtraImageFeatures {
// 		if !contains(validFeatures, feature) {
// 			return ValidationError{Field: fmt.Sprintf("features.extra_image_features[%d]", i), Message: fmt.Sprintf("invalid feature '%s', must be one of: %s", feature, strings.Join(validFeatures, ", "))}
// 		}
// 	}

// 	// Validate user classes
// 	validUserClasses := []string{"buildstats", "image-mklibs", "image-prelink"}
// 	for i, userClass := range f.UserClasses {
// 		if !contains(validUserClasses, userClass) {
// 			return ValidationError{Field: fmt.Sprintf("features.user_classes[%d]", i), Message: fmt.Sprintf("invalid user class '%s', must be one of: %s", userClass, strings.Join(validUserClasses, ", "))}
// 		}
// 	}

// 	// Validate inherit classes
// 	validInheritClasses := []string{
// 		"rm_work", "toradex-mirrors", "toradex-sanity", "buildstats",
// 		"image-mklibs", "image-prelink", "testimage", "testsdk",
// 	}
// 	for i, inheritClass := range f.InheritClasses {
// 		if !contains(validInheritClasses, inheritClass) {
// 			return ValidationError{Field: fmt.Sprintf("features.inherit_classes[%d]", i), Message: fmt.Sprintf("invalid inherit class '%s', must be one of: %s", inheritClass, strings.Join(validInheritClasses, ", "))}
// 		}
// 	}

// 	return nil
// }

// // Validate validates ContainerConfig
// func (c *ContainerConfig) Validate() error {
// 	// Validate memory format (e.g., "8g", "1024m")
// 	if c.Memory != "" {
// 		memoryRegex := regexp.MustCompile(`^\d+[gGmM]$`)
// 		if !memoryRegex.MatchString(c.Memory) {
// 			return ValidationError{Field: "container.memory", Message: "must be in format like '8g' or '1024m'"}
// 		}
// 	}

// 	// Validate CPU count
// 	if c.CPUCount < 0 {
// 		return ValidationError{Field: "container.cpu_count", Message: "must be non-negative"}
// 	}

// 	// Validate entrypoint parts if provided
// 	for i, part := range c.Entrypoint {
// 		if strings.TrimSpace(part) == "" {
// 			return ValidationError{Field: fmt.Sprintf("container.entrypoint[%d]", i), Message: "entrypoint parts must be non-empty"}
// 		}
// 		if strings.ContainsAny(part, "\n\r") {
// 			return ValidationError{Field: fmt.Sprintf("container.entrypoint[%d]", i), Message: "entrypoint parts must not contain newlines"}
// 		}
// 	}

// 	return nil
// }

// // CacheConfig and validation removed

// // Helper functions for validation

// func isValidPath(path string) bool {
// 	// Basic path validation - should not contain invalid characters
// 	pathRegex := regexp.MustCompile(`^[a-zA-Z0-9._/~$-]+$`)
// 	return pathRegex.MatchString(path)
// }

// func isValidPathOrEnvVar(path string) bool {
// 	// More permissive path validation that allows environment variable substitution
// 	// This is used after env var substitution has occurred
// 	// Allow common path characters including forward slashes
// 	pathRegex := regexp.MustCompile(`^[a-zA-Z0-9._/~$-:]+$`)
// 	return pathRegex.MatchString(path)
// }

// func isValidHostPort(hostPort string) bool {
// 	// Validate host:port format
// 	hostPortRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+:\d+$`)
// 	return hostPortRegex.MatchString(hostPort)
// }

// func contains(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }
