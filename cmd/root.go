package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "auto-pr",
	Short: "Automatically generate pull requests and merge requests using AI",
	Long: `Auto PR is a CLI tool that analyzes your git changes and uses AI to 
generate comprehensive pull requests and merge requests for GitHub and GitLab.

It analyzes your commits, code changes, and repository context to create
meaningful PR/MR titles, descriptions, and metadata automatically.`,
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.auto-pr/config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "preview changes without executing")
	
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic(fmt.Errorf("failed to bind verbose flag: %w", err))
	}
	if err := viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run")); err != nil {
		panic(fmt.Errorf("failed to bind dry-run flag: %w", err))
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		
		viper.AddConfigPath(home + "/.auto-pr")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}