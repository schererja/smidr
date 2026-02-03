package root

import (
	"fmt"
	"log/slog"

	daemonCmd "github.com/schererja/smidr/internal/cli/agent/daemon"
	initcmd "github.com/schererja/smidr/internal/cli/agent/init"
	config "github.com/schererja/smidr/internal/config/agent"
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
	Use:   "smidr-agent",
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

	rootCmd.AddCommand(daemonCmd.New(log, &cfg.AgentConfig))

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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/agent.yaml)")
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
		viper.SetConfigName("agent")
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
		log.Info("Loaded config", slog.String("controlPlane", cfg.AgentConfig.ControlPlaneAddress))

	}
}
