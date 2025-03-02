package processor

import (
	"encoding/json"
	"fmt"

	"github.com/yourusername/automation/pkg/models"
	"github.com/yourusername/automation/pkg/yaml"
)

type Processor struct {
	yamlHandler *yaml.Handler
}

func New() *Processor {
	return &Processor{
		yamlHandler: yaml.NewHandler(),
	}
}

func (p *Processor) Process(structureFile, jsonInput string) error {
	// Load structure
	structure, err := p.yamlHandler.LoadStructure(structureFile)
	if err != nil {
		return fmt.Errorf("failed to load structure: %w", err)
	}

	// Parse JSON input
	var jsonData []models.JSONInput
	if err := json.Unmarshal([]byte(jsonInput), &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Process each entry
	for _, entry := range structure.App {
		if err := p.processEntry(entry, jsonData); err != nil {
			fmt.Printf("Error processing entry %s: %v\n", entry.Name, err)
		}
	}

	return nil
}

func (p *Processor) processEntry(entry models.AppEntry, jsonData []models.JSONInput) error {
	for _, input := range jsonData {
		if !input.MatchesImage(entry.Name) {
			continue
		}

		// Handle legacy single file configuration
		if entry.Obsolete != "" {
			// Skip legacy handling if no files are defined
			if len(entry.Files) > 0 {
				// If Files is populated, assume Obsolete is truly obsolete
				fmt.Printf("Warning: Ignoring legacy file path %s as Files is populated\n", entry.Obsolete)
			} else {
				fmt.Printf("Warning: Using legacy file path %s. Please update to new format.\n", entry.Obsolete)
				// Create a dummy file config from the first update target
				if len(entry.Files) > 0 && len(entry.Files[0].UpdateTargets) > 0 {
					legacyFile := models.FileConfig{
						Path:          entry.Obsolete,
						UpdateTargets: entry.Files[0].UpdateTargets,
					}
					if err := p.updateFile(legacyFile, input); err != nil {
						fmt.Printf("Error updating legacy file %s: %v\n", entry.Obsolete, err)
					}
				}
			}
		}

		// Process all configured files
		for _, fileConfig := range entry.Files {
			if err := p.updateFile(fileConfig, input); err != nil {
				fmt.Printf("Error updating file %s: %v\n", fileConfig.Path, err)
				continue // Continue with other files even if one fails
			}
		}
	}
	return nil
}

func (p *Processor) updateFile(fileConfig models.FileConfig, input models.JSONInput) error {
	return p.yamlHandler.UpdateFile(fileConfig.Path, fileConfig.UpdateTargets, input)
}
