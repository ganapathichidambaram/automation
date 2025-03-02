package yaml

import (
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/automation/pkg/models"
	"gopkg.in/yaml.v3"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) LoadStructure(filename string) (*models.Structure, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var structure models.Structure
	if err := yaml.Unmarshal(data, &structure); err != nil {
		return nil, err
	}

	return &structure, nil
}

func (h *Handler) UpdateFile(filePath string, targets []models.UpdatePath, input models.JSONInput) error {
	var node yaml.Node
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(fileData, &node); err != nil {
		return err
	}

	// Process each update target
	for _, target := range targets {
		if err := h.updateNode(&node, target, target.ObjectKey, input.Version(), input.Sha); err != nil {
			return fmt.Errorf("failed to update path %s: %w", target.StructurePath, err)
		}
	}

	return h.writeFile(filePath, &node)
}

func (h *Handler) writeFile(filename string, node *yaml.Node) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(node)
}

func (h *Handler) updateNode(node *yaml.Node, target models.UpdatePath, objectKey, version, sha string) error {
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("expected document node")
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node")
	}

	// First navigate to the structure path
	current := root
	if err := h.navigateToPath(target.StructurePath, &current); err != nil {
		return fmt.Errorf("structure path error: %w", err)
	}

	// Find the object by key (objectKey is already handled by the caller)
	found := false
	for i := 0; i < len(current.Content); i += 2 {
		if current.Content[i].Value == objectKey {
			found = true
			if target.ImageParentPath != "" {
				// If we have a parent path, navigate to it from the object node
				parentNode := current.Content[i+1]
				if err := h.navigateToPath(target.ImageParentPath, &parentNode); err != nil {
					return fmt.Errorf("parent path error: %w", err)
				}
				current = parentNode
			} else {
				current = current.Content[i+1]
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("object key %s not found", objectKey)
	}

	// Update the imageTag and imageDigest at the final location
	return h.updateImageValues(current, version, sha)
}

func (h *Handler) navigateToPath(path string, current **yaml.Node) error {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		if (*current).Kind != yaml.MappingNode && (*current).Kind != yaml.SequenceNode {
			return fmt.Errorf("expected mapping or sequence node while navigating to %s", part)
		}

		// Check if we're dealing with a sequence node access (e.g., "containers" in path)
		isSequence := false
		key := part
		if (*current).Kind == yaml.MappingNode {
			// Look for the key first
			found := false
			for i := 0; i < len((*current).Content); i += 2 {
				if (*current).Content[i].Value == key {
					*current = (*current).Content[i+1]
					found = true
					if (*current).Kind == yaml.SequenceNode {
						isSequence = true
					}
					break
				}
			}
			if !found {
				return fmt.Errorf("key %s not found", key)
			}
		}

		// If we found a sequence node, check if the next part specifies an item
		if isSequence {
			// Look for a node with matching 'name' field
			found := false
			for _, item := range (*current).Content {
				// Each item should be a mapping node
				if item.Kind != yaml.MappingNode {
					continue
				}

				// Look for the 'name' field in the mapping
				for i := 0; i < len(item.Content); i += 2 {
					if item.Content[i].Value == "name" {
						if item.Content[i+1].Value == key {
							*current = item
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
			if !found {
				return fmt.Errorf("container with name %s not found", key)
			}
		}
	}
	return nil
}

func (h *Handler) updateImageValues(node *yaml.Node, version, sha string) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node for image values")
	}

	foundTag := false
	foundDigest := false

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		originalStyle := valueNode.Style

		switch keyNode.Value {
		case models.ImageTagKey:
			valueNode.Value = version
			valueNode.Style = originalStyle
			foundTag = true
		case models.ImageDigestKey:
			valueNode.Value = sha
			valueNode.Style = originalStyle
			foundDigest = true
		}
	}

	if !foundTag || !foundDigest {
		return fmt.Errorf("missing required keys imageTag and/or imageDigest")
	}

	return nil
}
