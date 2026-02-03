package root

import (
	"fmt"
	"log/slog"

	initcmd "github.com/schererja/smidr/internal/cli/controlplane/init"
	serveCmd "github.com/schererja/smidr/internal/cli/controlplane/serve"
	config "github.com/schererja/smidr/internal/config/controlplane"
	"github.com/schererja/smidr/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	log     *logger.Logger
	cfg     *config.Config
)
var rootCmd = &cobra.Command{
	Use:   "smidr",
	Short: "The digital forge for your embedded Linux builds",
	Long: `Smidr is a command-line tool designed to streamline and enhance the process of building
embedded Linux systems. It provides a comprehensive suite of features to manage configurations,
dependencies, and build processes, making it easier for developers to create and maintain
custom Linux distributions for embedded devices.`,
	Version: "0.1.0-dev",
}

func Execute(logger *logger.Logger) error {
	log = logger
	initConfig()
	if log == nil {
		return fmt.Errorf("Logger not initialized")
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	// // Add commands from subpackages
	rootCmd.AddCommand(initcmd.New(log))

	rootCmd.AddCommand(serveCmd.New(log, &cfg.ServerConfig))
	// rootCmd.AddCommand(buildcmd.New())
	// rootCmd.AddCommand(clientcmd.New())
	// rootCmd.AddCommand(artifacts.New())
	// rootCmd.AddCommand(daemon.New(log))
	// rootCmd.AddCommand(logs.New())
	// rootCmd.AddCommand(status.New())
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("RootCommand failure: %v", err)
	}
	return nil
}

// GetLogger returns the global logger instance for use in subcommands
func GetLogger() *logger.Logger {
	return log
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/smidr.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	viper.SetEnvPrefix("SMIDR")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			log.Info("Using config file:", slog.String("file", viper.ConfigFileUsed()))
		}
		cfg = &config.Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			log.Error("Failed to parse configuration", err)
		}

	}
}
