package cli

import (
	"fmt"
	"os"

	buildcmd "github.com/schererja/smidr/internal/cli/build"
	clientcmd "github.com/schererja/smidr/internal/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
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

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("RootCommand failure: %v", err)
	}
	return nil
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/smidr.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add commands from subpackages
	rootCmd.AddCommand(buildcmd.New())
	rootCmd.AddCommand(clientcmd.New())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("smidr")
	}
	viper.SetEnvPrefix("SMIDR")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verbose") {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())

	}
}
