package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/schererja/smidr/internal/artifacts"
	buildpkg "github.com/schererja/smidr/internal/build"
	config "github.com/schererja/smidr/internal/config"
	docker "github.com/schererja/smidr/internal/container/docker"
	"github.com/schererja/smidr/pkg/logger"
)

// runBuildRefactored is a simplified version using the shared runner
func runBuildRefactored(ctx context.Context, configFile string, flags buildFlags) error {
	log := logger.NewLogger()
	// Helper to expand ~ and make absolute
	expandPath := func(path string) string {
		if path == "" {
			return ""
		}
		if strings.HasPrefix(path, "~") {
			if len(path) == 1 || path[1] == '/' {
				homedir, err := os.UserHomeDir()
				if err != nil {
					panic("Could not resolve ~ to home directory: " + err.Error())
				}
				path = homedir + path[1:]
			}
		}
		if !strings.HasPrefix(path, "/") {
			abs, err := os.Getwd()
			if err != nil {
				panic("Could not get working directory: " + err.Error())
			}
			path = abs + "/" + path
		}
		return path
	}

	fmt.Println("🔨 Starting Smidr build...")
	fmt.Println()

	fmt.Printf("📄 Loading configuration from %s...\n", configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}
	fmt.Printf("✅ Loaded project: %s\n", cfg.Name)
	fmt.Println()

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	setDefaultDirs(cfg, workDir)

	// Set up customer-specific or unique build directory
	homedir, _ := os.UserHomeDir()
	imageName := cfg.Build.Image

	var buildDir string
	if flags.customer != "" {
		buildDir = fmt.Sprintf("%s/.smidr/builds/build-%s/%s", homedir, flags.customer, imageName)
		if flags.clean {
			os.RemoveAll(buildDir)
			fmt.Printf("🧹 Cleaned build directory: %s\n", buildDir)
		}
	} else {
		buildUUID := uuid.New().String()
		buildDir = fmt.Sprintf("%s/.smidr/builds/build-%s", homedir, buildUUID)
	}

	// TMPDIR setup
	var tmpDir string
	if cfg.Directories.Tmp != "" {
		tmpDir = expandPath(cfg.Directories.Tmp)
		if flags.clean && flags.customer != "" {
			tmpDir = fmt.Sprintf("%s/%s-%s", tmpDir, flags.customer, imageName)
			os.RemoveAll(tmpDir)
			fmt.Printf("🧹 Cleaned TMPDIR: %s\n", tmpDir)
		}
	} else {
		tmpDir = fmt.Sprintf("%s/tmp", buildDir)
	}
	deployDir := fmt.Sprintf("%s/deploy", buildDir)

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

	fmt.Println("🔒 Using:")
	fmt.Printf("    TMPDIR: %s\n", tmpDir)
	fmt.Printf("    DEPLOY_DIR: %s\n", deployDir)
	fmt.Printf("    BUILD_DIR: %s\n", buildDir)

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
	if flags.fetchOnly {
		fmt.Println("🛑 Fetch-only mode enabled — skipping container start and build")
		// Runner will fetch layers for us, but we'll stop before building
		// For now, just return
		return nil
	}

	// Prepare build options from flags
	opts := buildpkg.BuildOptions{
		Target:     cfg.Build.Image,
		Customer:   flags.customer,
		ForceClean: flags.clean,
		ForceImage: flags.cleanImage,
	}

	// Create a log sink that writes to stdout and log files
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

	log.Info("🚀 Starting Yocto build via runner")
	log.Info("💡 Use Ctrl+C to gracefully cancel the build")
	log.Info("Project loaded", slog.String("name", cfg.Name), slog.String("description", cfg.Description))

	// Use the shared runner
	runner := buildpkg.NewRunner(log)
	buildResult, err := runner.Run(ctx, cfg, opts, logSink)

	if err != nil {
		if ctx.Err() == context.Canceled {
			fmt.Printf("🛑 Build was cancelled by user signal\n")
			if buildResult != nil {
				fmt.Printf("Build was running for: %v before cancellation\n", buildResult.Duration)
			}
			return fmt.Errorf("build cancelled by user")
		}

		fmt.Printf("❌ Build failed: %v\n", err)
		if buildResult != nil {
			fmt.Printf("Build took: %v\n", buildResult.Duration)
		}
		return fmt.Errorf("build execution failed: %w", err)
	}

	fmt.Printf("✅ Build completed successfully in %v\n", buildResult.Duration)
	fmt.Printf("Exit code: %d\n", buildResult.ExitCode)

	// Extract build artifacts
	fmt.Println("📦 Extracting build artifacts...")
	dm, err := docker.NewDockerManager()
	if err != nil {
		fmt.Printf("[WARNING] Failed to create docker manager for artifact extraction: %v\n", err)
	} else {
		if err := extractBuildArtifactsFromRunner(ctx, dm, cfg, buildResult); err != nil {
			fmt.Printf("[WARNING] Failed to extract artifacts: %v\n", err)
		}
	}

	fmt.Println("💡 Use 'smidr artifacts list' to view build artifacts once available")
	return nil
}

// buildFlags holds all the CLI flags for the build command
type buildFlags struct {
	customer   string
	clean      bool
	cleanImage bool
	fetchOnly  bool
}

// extractBuildArtifactsFromRunner extracts artifacts when using the runner
func extractBuildArtifactsFromRunner(ctx context.Context, dm *docker.DockerManager, cfg *config.Config, result *buildpkg.BuildResult) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Determine artifact directory
	artifactDir := ""
	customer := ""
	imageName := cfg.Build.Image
	timestamp := time.Now().Format("20060102-150405")

	if strings.Contains(cfg.Directories.Build, "/build-") {
		parts := strings.Split(cfg.Directories.Build, "/build-")
		if len(parts) > 1 {
			customerAndRest := parts[1]
			customerParts := strings.SplitN(customerAndRest, "/", 2)
			customer = customerParts[0]
		}
	}

	if customer != "" {
		artifactDir = fmt.Sprintf("%s/.smidr/artifacts/artifact-%s/%s-%s", currentUser.HomeDir, customer, imageName, timestamp)
	} else {
		buildID := artifacts.GenerateBuildID(cfg.Name, currentUser.Username)
		artifactDir = fmt.Sprintf("%s/.smidr/artifacts/%s", currentUser.HomeDir, buildID)
	}

	// Use result's DeployDir directly
	deploySrc := result.DeployDir
	deployDst := filepath.Join(artifactDir, "deploy")

	info, statErr := os.Stat(deploySrc)
	if statErr != nil || !info.IsDir() {
		fmt.Printf("[INFO] No deploy directory to extract\n")
		return nil
	}

	if err := copyDir(deploySrc, deployDst); err != nil {
		return fmt.Errorf("failed to copy deploy artifacts: %w", err)
	}

	// Copy build logs
	buildLogTxt := filepath.Join(cfg.Directories.Build, "build-log.txt")
	destLogTxt := filepath.Join(artifactDir, "build-log.txt")
	if err := copyFile(buildLogTxt, destLogTxt); err != nil {
		fmt.Printf("[WARNING] Failed to copy build-log.txt: %v\n", err)
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

	fmt.Printf("✅ Artifacts copied to: %s\n", artifactDir)
	return nil
}
