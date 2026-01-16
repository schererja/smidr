package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrik8-labs/smidr/internal/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	log     *logging.Logger
	added   bool
)
var rootCmd = &cobra.Command{
	Use:   "smidr-core",
	Short: "The digital forge for your embedded Linux builds (core)",
	Long: `Smidr is a command-line tool designed to streamline and enhance the process of building
embedded Linux systems. It provides a comprehensive suite of features to manage configurations,
dependencies, and build processes, making it easier for developers to create and maintain
custom Linux distributions for embedded devices.`,
	Version: "0.1.0-dev",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// cfg, err := config.LoadFromViper()
		// if err != nil {
		// 	return fmt.Errorf("failed to load configuration: %w", err)
		// }
		// ctx = logging.WithExecutor(ctx, cfg.AgentConfig.ID)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// Start server in a goroutine
		errCh := make(chan error, 1)
		go func() {

		}()

		// Wait for shutdown signal or error
		select {
		case <-sigCh:
			log.Info("\nReceived shutdown signal")
			// server.Stop()
			return nil
		case err := <-errCh:
			return fmt.Errorf("daemon error: %w", err)
		case <-ctx.Done():
			// server.Stop()
			return nil
		}

	},
}

func Execute(logger *logging.Logger) error {
	log = logger

	if !added {
		// rootCmd.AddCommand(initcmd.New(log))
		added = true
	}
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("RootCommand failure: %v", err)
	}
	return nil
}

// GetLogger returns the global logger instance for use in subcommands
func GetLogger() *logging.Logger {
	return log
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/smidr.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Commands are added during Execute once logger is available
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("smidr-core")
	}
	viper.SetEnvPrefix("SMIDR")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			// Error only if config was explicitly specified
			fmt.Fprintf(os.Stderr, "Error: failed to read config file %s: %v\n", cfgFile, err)
		}
		// Otherwise silently continue (no config file is ok)
	} else if verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}
