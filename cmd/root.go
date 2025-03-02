package cmd

import (
	"github.com/spf13/cobra"
)

var (
	structureFile string
	jsonInput     string

	RootCmd = &cobra.Command{
		Use:   "automation",
		Short: "A tool for automated YAML updates",
	}
)

func init() {
	RootCmd.PersistentFlags().StringVar(&structureFile, "structure", "structure.yaml", "Path to structure YAML file")
	RootCmd.PersistentFlags().StringVar(&jsonInput, "input", "", "JSON input string")
	RootCmd.AddCommand(updateCmd)
}
