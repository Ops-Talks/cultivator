// Package cmd provides command-line interface definitions for Cultivator.
package cmd

import (
	"fmt"
	"os"

	"github.com/cultivator-dev/cultivator/pkg/config"
	"github.com/cultivator-dev/cultivator/pkg/orchestrator"
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command for cultivator
func NewRootCommand(version, commit, date string) *cobra.Command {
	var configFile string
	var workingDir string

	rootCmd := &cobra.Command{
		Use:   "cultivator",
		Short: "Terragrunt automation for Pull Requests",
		Long: `Cultivator automates Terragrunt plan and apply operations from Pull Requests.
		
Similar to Digger/Atlantis for Terraform, but built specifically for Terragrunt
with support for dependencies, run-all operations, and hierarchical configs.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "cultivator.yml", "Config file path")
	rootCmd.PersistentFlags().StringVarP(&workingDir, "dir", "d", ".", "Working directory")

	// Add subcommands
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newPlanCommand())
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.AddCommand(newDetectCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newVersionCommand(version, commit, date))

	return rootCmd
}

func newRunCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run cultivator based on GitHub event",
		Long:  "Automatically detects the GitHub event type and runs the appropriate action",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			fmt.Println("Running Cultivator...")

			// Load configuration
			cfg, err := loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get GitHub token
			ghToken := os.Getenv("GITHUB_TOKEN")
			if ghToken == "" {
				return fmt.Errorf("GITHUB_TOKEN environment variable not set")
			}

			// Create and run orchestrator
			orch, err := createOrchestrator(cfg, ghToken)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator: %w", err)
			}

			return orch.Run(ctx)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "cultivator.yml", "Path to configuration file")

	return cmd
}

func newPlanCommand() *cobra.Command {
	var all bool
	var module string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Run terragrunt plan",
		RunE: func(_ *cobra.Command, _ []string) error {
			if all {
				fmt.Println("Running plan-all...")
			} else if module != "" {
				fmt.Printf("Running plan for module: %s\n", module)
			} else {
				fmt.Println("Running plan for changed modules...")
			}
			// TODO: Implement plan logic
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Plan all modules")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specific module to plan")

	return cmd
}

func newApplyCommand() *cobra.Command {
	var all bool
	var module string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Run terragrunt apply",
		RunE: func(_ *cobra.Command, _ []string) error {
			if all {
				fmt.Println("Running apply-all...")
			} else if module != "" {
				fmt.Printf("Running apply for module: %s\n", module)
			} else {
				fmt.Println("Running apply for changed modules...")
			}
			// TODO: Implement apply logic
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Apply all modules")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specific module to apply")

	return cmd
}

func newDetectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Detect changed Terragrunt modules",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Detecting changed modules...")
			// TODO: Implement detection logic
			return nil
		},
	}
}

func newValidateCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate cultivator configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("Validating configuration: %s\n", configPath)

			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("configuration is invalid: %w", err)
			}

			// Validate configuration
			if err := validateConfig(cfg); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Println("Configuration is valid")
			fmt.Printf("\nProjects: %d\n", len(cfg.Projects))
			for _, proj := range cfg.Projects {
				fmt.Printf("  - %s (%s)\n", proj.Name, proj.Dir)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "cultivator.yml", "Path to configuration file")

	return cmd
}

func validateConfig(cfg *config.Config) error {
	if len(cfg.Projects) == 0 {
		return fmt.Errorf("no projects defined")
	}

	for i, project := range cfg.Projects {
		if project.Name == "" {
			return fmt.Errorf("project %d: name is required", i)
		}
		if project.Dir == "" {
			return fmt.Errorf("project %s: dir is required", project.Name)
		}
	}

	return nil
}

func newVersionCommand(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Cultivator %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", date)
		},
	}
}

// Helper functions

func loadConfig(path string) (*config.Config, error) {
	return config.LoadConfig(path)
}

func createOrchestrator(cfg *config.Config, ghToken string) (*orchestrator.Orchestrator, error) {
	return orchestrator.NewOrchestrator(cfg, ghToken, os.Stdout, os.Stderr)
}
