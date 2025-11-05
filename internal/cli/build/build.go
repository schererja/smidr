package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/schererja/smidr/internal/artifacts"
	buildpkg "github.com/schererja/smidr/internal/build"
	config "github.com/schererja/smidr/internal/config"
	docker "github.com/schererja/smidr/internal/container/docker"
	"github.com/schererja/smidr/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New creates and returns the build command
func New() *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build a Yocto-based embedded Linux image in a managed container",
		Long: `Run a Yocto build in a managed container environment using the configuration in smidr.yaml.

	This command will:
		1. Prepare the container environment
		2. Fetch and cache source code and layers
		3. Execute the build process (bitbake)
		4. Extract build artifacts and store them in the artifact directory

	Flags allow you to override the build target, force a rebuild, or group builds by customer/project.

	Examples:
		smidr build --target core-image-minimal
		smidr build --target core-image-minimal --customer acme
		smidr build --target core-image-minimal --force
		smidr build --target core-image-minimal --fetch-only
		smidr build --target core-image-minimal --clean
	`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runBuild(cmd); err != nil {
				fmt.Println("Error during build:", err)
				os.Exit(1)
			}
		},
	}

	// Build-specific flags
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuild (ignore cache)")
	buildCmd.Flags().StringP("target", "t", "", "Override build target")
	buildCmd.Flags().Bool("fetch-only", false, "Only fetch layers but don't build it")
	buildCmd.Flags().String("customer", "", "Optional: customer/user name for build directory grouping")
	buildCmd.Flags().Bool("clean", false, "If set, deletes the build directory before building (for a full rebuild)")
	buildCmd.Flags().Bool("clean-image", false, "If set, runs 'bitbake -c clean <image>' to regenerate only image artifacts without rebuilding dependencies")

	return buildCmd
}

func runBuild(cmd *cobra.Command) error {
	// Initialize logger
	log := logger.NewLogger()

	// Set up signal handling for graceful cancellation
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received signal, initiating graceful shutdown")
		cancel()
	}()

	configFile := viper.GetString("config")
	if configFile == "" {
		configFile = "smidr.yaml"
	}

	// Get flags
	customer, _ := cmd.Flags().GetString("customer")
	clean, _ := cmd.Flags().GetBool("clean")
	cleanImage, _ := cmd.Flags().GetBool("clean-image")
	fetchOnly, _ := cmd.Flags().GetBool("fetch-only")

	log.Info("🔨 Starting Smidr build")
	log.Info("📄 Loading configuration", slog.String("file", configFile))

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	log.Info("Loaded project", slog.String("name", cfg.Name))
	if cfg.Description != "" {
		log.Debug("Project description", slog.String("description", cfg.Description))
	}

	// Set up build directories
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	setDefaultDirs(cfg, workDir)

	// Expand paths helper
	expandPath := func(path string) string {
		if path == "" {
			return ""
		}
		if strings.HasPrefix(path, "~") {
			if len(path) == 1 || path[1] == '/' {
				homedir, err := os.UserHomeDir()
				if err != nil {
					log.Fatal("Could not resolve ~ to home directory: ", err)
				}
				path = homedir + path[1:]
			}
		}
		if !strings.HasPrefix(path, "/") {
			abs, err := os.Getwd()
			if err != nil {
				log.Fatal("Could not get working directory: ", err)
			}
			path = abs + "/" + path
		}
		return path
	}

	// Create unique or customer-specific build directory
	homedir, _ := os.UserHomeDir()
	imageName := cfg.Build.Image
	buildUUID := uuid.New().String()

	var buildDir string
	if customer != "" {
		buildDir = fmt.Sprintf("%s/.smidr/builds/build-%s/%s", homedir, customer, imageName)
		if clean {
			os.RemoveAll(buildDir)
			log.Info("Cleaned build directory", slog.String("path", buildDir))
		}
	} else {
		buildDir = fmt.Sprintf("%s/.smidr/builds/build-%s", homedir, buildUUID)
	}

	// TMPDIR: use config if set, else per-build
	var tmpDir string
	if cfg.Directories.Tmp != "" {
		tmpDir = expandPath(cfg.Directories.Tmp)
		if clean && customer != "" {
			tmpDir = fmt.Sprintf("%s/%s-%s", tmpDir, customer, imageName)
			os.RemoveAll(tmpDir)
			log.Info("🧹 Cleaned TMPDIR", slog.String("path", tmpDir))
		}
	} else {
		tmpDir = fmt.Sprintf("%s/tmp", buildDir)
	}

	deployDir := fmt.Sprintf("%s/deploy", buildDir)

	// Ensure directories exist with permissive permissions
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		return fmt.Errorf("failed to create tmp dir: %w", err)
	}
	if err := os.Chown(tmpDir, 1000, 1000); err != nil {
		_ = os.Chmod(tmpDir, 0777)
	}
	if err := os.MkdirAll(deployDir, 0777); err != nil {
		return fmt.Errorf("failed to create deploy dir: %w", err)
	}

	cfg.Directories.Build = buildDir
	cfg.Directories.Tmp = tmpDir
	cfg.Directories.Deploy = deployDir

	log.Info("🔒 Build directories configured",
		slog.String("tmpdir", tmpDir),
		slog.String("deploy_dir", deployDir),
		slog.String("build_dir", buildDir))

	// Expand and resolve all relevant directories
	downloadsDir := cfg.Directories.Downloads
	if downloadsDir == "" && cfg.Directories.Source != "" {
		downloadsDir = cfg.Directories.Source
	}
	downloadsDir = expandPath(downloadsDir)
	if downloadsDir != "" {
		cfg.Directories.Downloads = downloadsDir
	}
	cfg.Directories.Layers = expandPath(cfg.Directories.Layers)
	cfg.Directories.Source = expandPath(cfg.Directories.Source)
	cfg.Directories.SState = expandPath(cfg.Directories.SState)
	cfg.Directories.Build = expandPath(cfg.Directories.Build)
	cfg.Directories.Tmp = expandPath(cfg.Directories.Tmp)
	cfg.Directories.Deploy = expandPath(cfg.Directories.Deploy)

	// If --fetch-only was requested, stop here
	if fetchOnly {
		log.Info("🛑 Fetch-only mode enabled, skipping container start and build")
		return nil
	}

	// Prepare build options
	opts := buildpkg.BuildOptions{
		BuildID:    buildUUID,
		Target:     cfg.Build.Image,
		Customer:   customer,
		ForceClean: clean,
		ForceImage: cleanImage,
	}

	// Create log files
	logDir := cfg.Directories.Build
	os.MkdirAll(logDir, 0755)
	plainLogPath := filepath.Join(logDir, "build-log.txt")
	plainLogFile, err := os.Create(plainLogPath)
	if err != nil {
		return fmt.Errorf("failed to create build-log.txt: %w", err)
	}
	defer plainLogFile.Close()

	logSink := &cliLogSink{
		stdout: io.MultiWriter(os.Stdout, plainLogFile),
	}

	log.Info("🚀 Starting Yocto build")
	log.Info("💡 Use Ctrl+C to gracefully cancel the build")

	// Use the Runner (no DB in CLI mode)
	runner := buildpkg.NewRunner(log, nil)
	buildResult, err := runner.Run(ctx, cfg, opts, logSink)

	if err != nil {
		if ctx.Err() == context.Canceled {
			log.Warn("🛑 Build was cancelled by user signal")
			if buildResult != nil {
				log.Info("Build duration before cancellation", slog.Duration("duration", buildResult.Duration))
			}
			return fmt.Errorf("build cancelled by user")
		}

		log.Error("❌ Build failed", err)
		if buildResult != nil {
			log.Info("Build duration", slog.Duration("duration", buildResult.Duration))
		}
		return fmt.Errorf("build execution failed: %w", err)
	}

	log.Info("✅ Build completed successfully",
		slog.Duration("duration", buildResult.Duration),
		slog.Int("exit_code", buildResult.ExitCode))

	// Extract build artifacts
	log.Info("📦 Extracting build artifacts")
	dm, err := docker.NewDockerManager(log)
	if err != nil {
		log.Warn("Failed to create docker manager for artifact extraction", slog.String("error", err.Error()))
	} else {
		if err := extractBuildArtifacts(ctx, dm, cfg, customer, imageName, buildResult, log); err != nil {
			log.Warn("Failed to extract artifacts", slog.String("error", err.Error()))
		}
	}

	log.Info("💡 Use 'smidr artifacts list' to view build artifacts")
	return nil
}

// Helper to expand ~ and make absolute

// setDefaultDirs ensures default directory paths are populated on the config
// when the user did not provide them. This centralizes defaulting and makes
// it testable.
func setDefaultDirs(cfg *config.Config, workDir string) {
	if cfg.Directories.Source == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Directories.Source = fmt.Sprintf("%s/.smidr/sources", homedir)
		} else {
			cfg.Directories.Source = fmt.Sprintf("%s/sources", workDir)
		}
	}
	if cfg.Directories.Build == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Directories.Build = fmt.Sprintf("%s/.smidr/build", homedir)
		} else {
			cfg.Directories.Build = fmt.Sprintf("%s/build", workDir)
		}
	}
	if cfg.Directories.SState == "" {
		if homedir, err := os.UserHomeDir(); err == nil {
			cfg.Directories.SState = fmt.Sprintf("%s/.smidr/sstate-cache", homedir)
		} else {
			cfg.Directories.SState = fmt.Sprintf("%s/sstate-cache", workDir)
		}
	}

	// Provide sane defaults for global cache locations if not supplied
	// Removed cache defaults; defaults exist for Directories in setDefaultDirs
}

// extractBuildArtifacts extracts build artifacts from the build result to persistent storage
func extractBuildArtifacts(ctx context.Context, dm *docker.DockerManager, cfg *config.Config, customer, imageName string, result *buildpkg.BuildResult, log *logger.Logger) error {
	// Get current user for metadata
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Determine artifact directory based on customer
	artifactDir := ""
	timestamp := time.Now().Format("20060102-150405")

	if customer != "" {
		// Use customer artifact dir
		artifactDir = fmt.Sprintf("%s/.smidr/artifacts/artifact-%s/%s-%s", currentUser.HomeDir, customer, imageName, timestamp)
	} else {
		// Fallback: use buildID as before
		buildID := artifacts.GenerateBuildID(cfg.Name, currentUser.Username)
		artifactDir = fmt.Sprintf("%s/.smidr/artifacts/%s", currentUser.HomeDir, buildID)
	}

	// Use result's DeployDir directly (Runner already determined the correct path)
	deploySrc := result.DeployDir
	deployDst := filepath.Join(artifactDir, "deploy")

	// Check if source exists and is a directory
	info, statErr := os.Stat(deploySrc)
	if statErr != nil || !info.IsDir() {
		return fmt.Errorf("deploy source directory does not exist or is not a directory: %v", statErr)
	}

	// Copy deploy artifacts
	if err := copyDir(deploySrc, deployDst); err != nil {
		return fmt.Errorf("failed to copy deploy artifacts: %w", err)
	}

	// Copy build logs to artifact directory
	buildLogTxt := filepath.Join(cfg.Directories.Build, "build-log.txt")
	buildLogJsonl := filepath.Join(cfg.Directories.Build, "build-log.jsonl")
	destLogTxt := filepath.Join(artifactDir, "build-log.txt")
	destLogJsonl := filepath.Join(artifactDir, "build-log.jsonl")
	if err := copyFile(buildLogTxt, destLogTxt); err != nil {
		log.Warn("Failed to copy build-log.txt to artifact dir", slog.String("error", err.Error()))
	}
	if err := copyFile(buildLogJsonl, destLogJsonl); err != nil {
		log.Warn("Failed to copy build-log.jsonl to artifact dir", slog.String("error", err.Error()))
	}

	// Create build metadata
	metadata := artifacts.BuildMetadata{
		BuildID:     filepath.Base(artifactDir),
		ProjectName: cfg.Name,
		User:        currentUser.Username,
		Timestamp:   time.Now(),
		ConfigUsed: map[string]string{
			"yocto_series":  cfg.YoctoSeries,
			"machine":       cfg.Build.Machine,
			"image":         cfg.Build.Image,
			"base_provider": cfg.Base.Provider,
			"base_version":  cfg.Base.Version,
		},
		BuildDuration: result.Duration,
		TargetImage:   cfg.Build.Image,
		Status:        "success",
	}

	// Save metadata as JSON
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifact dir: %w", err)
	}
	metaPath := filepath.Join(artifactDir, "build-metadata.json")
	f, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	log.Info("✅ Artifacts copied", slog.String("path", artifactDir))
	return nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// best-effort copy; log to stdout via stderr semantics not available here
			// leave as silent skip to reduce noise in logs
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			// fmt.Printf("[DEBUG] Failed to get relative path for %s: %v\n", path, err)
			return err
		}
		target := filepath.Join(dst, rel)
		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, lerr := os.Readlink(path)
			if lerr != nil {
				// skip broken symlink silently
				return nil
			}
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(linkTarget, target); err != nil {
				// ignore symlink creation issues (best-effort copy)
			}
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		} else {
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			srcFile, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer srcFile.Close()
			dstFile, err := os.Create(target)
			if err != nil {
				return err
			}
			defer dstFile.Close()
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}
			if err := os.Chmod(target, info.Mode()); err != nil {
				// ignore chmod issues
			}
			return nil
		}
	})
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// cliLogSink adapts the runner's LogSink interface to write to stdout and log files
type cliLogSink struct {
	stdout io.Writer
}

func (s *cliLogSink) Write(stream, line string) {
	if s.stdout != nil {
		fmt.Fprintln(s.stdout, line)
	}
}
