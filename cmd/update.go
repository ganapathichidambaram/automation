package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/automation/pkg/processor"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update YAML files based on JSON input",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := processor.New()
		return p.Process(structureFile, jsonInput)
	},
}
