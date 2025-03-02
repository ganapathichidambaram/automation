package models

import "strings"

type JSONInput struct {
	Image string `json:"image"`
	Sha   string `json:"sha"`
}

func (j *JSONInput) MatchesImage(name string) bool {
	imageName := strings.Split(j.Image, ":")[0]
	return imageName == name
}

func (j *JSONInput) Version() string {
	parts := strings.Split(j.Image, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

type Structure struct {
	App []AppEntry `yaml:"app"`
}

type AppEntry struct {
	Name     string       `yaml:"name"`
	Files    []FileConfig `yaml:"files"`          // Changed from single file to multiple
	Obsolete string       `yaml:"file,omitempty"` // For backward compatibility
}

type FileConfig struct {
	Path          string       `yaml:"path"`
	UpdateTargets []UpdatePath `yaml:"update-targets"`
}

type UpdatePath struct {
	StructurePath   string `yaml:"structure-path"`
	ImageParentPath string `yaml:"image-parent-path,omitempty"` // Optional, path to parent of imageTag/imageDigest
	ObjectKey       string `yaml:"object-key,omitempty"`        // Optional, defaults to image name
}

// Constants for key names
const (
	ImageTagKey    = "imageTag"
	ImageDigestKey = "imageDigest"
)
