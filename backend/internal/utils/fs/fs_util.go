package fs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/getarcaneapp/arcane/backend/internal/common"
	"github.com/getarcaneapp/arcane/backend/internal/utils/projects"
)

func GetProjectsDirectory(ctx context.Context, projectsDir string) (string, error) {
	// Prefer host-absolute projects directory in development to avoid Docker Desktop mount issues on macOS
	// If HOST_PROJECT_ROOT is provided (from compose.dev) we build the absolute host path
	// e.g. /Users/you/Dev/arcane/backend/data/projects
	if os.Getenv("ENVIRONMENT") == "development" {
		if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
			hostProjects := filepath.Join(hostRoot, "backend", "data", "projects")
			// If the path exists or can be created, use it
			if _, herr := os.Stat(hostProjects); os.IsNotExist(herr) {
				if mkErr := os.MkdirAll(hostProjects, common.DirPerm); mkErr == nil {
					slog.InfoContext(ctx, "Created host projects directory", "path", hostProjects)
					return hostProjects, nil
				}
			} else if herr == nil {
				return hostProjects, nil
			}
			// Fall through to default if host path unusable
			slog.DebugContext(ctx, "Host projects directory not usable, falling back", "path", hostProjects)
		}
	}

	projectsDirectory := projectsDir
	if projectsDirectory == "" {
		projectsDirectory = "data/projects"
	}

	if _, err := os.Stat(projectsDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(projectsDirectory, common.DirPerm); err != nil {
			return "", err
		}
		slog.InfoContext(ctx, "Created projects directory", "path", projectsDirectory)
	}

	return projectsDirectory, nil
}

func ReadProjectFiles(projectPath string) (composeContent, envContent string, err error) {
	if composeFile, derr := projects.DetectComposeFile(projectPath); derr == nil && composeFile != "" {
		if content, rerr := os.ReadFile(composeFile); rerr == nil {
			composeContent = string(content)
		}
	}

	envPath := filepath.Join(projectPath, ".env")
	if content, rerr := os.ReadFile(envPath); rerr == nil {
		envContent = string(content)
	}

	return composeContent, envContent, nil
}

func GetTemplatesDirectory(ctx context.Context) (string, error) {
	templatesDir := filepath.Join("data", "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(templatesDir, common.DirPerm); err != nil {
			return "", err
		}
		slog.InfoContext(ctx, "Created templates directory", "path", templatesDir)
	}
	return templatesDir, nil
}

// CreateUniqueDir creates a unique directory within the allowed projectsRoot.
// It validates that the created directory is always within projectsRoot.
func CreateUniqueDir(projectsRoot, basePath, name string, perm os.FileMode) (path, folderName string, err error) {
	sanitized := SanitizeProjectName(name)

	// Reject empty or invalid sanitized names
	if sanitized == "" || strings.Trim(sanitized, "_") == "" {
		return "", "", fmt.Errorf("invalid project name: results in empty directory name")
	}

	// Get absolute path of the true projects root for validation
	projectsRootAbs, err := filepath.Abs(projectsRoot)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve projects root directory: %w", err)
	}
	projectsRootAbs = filepath.Clean(projectsRootAbs)

	candidate := basePath
	folderName = sanitized

	for counter := 1; ; counter++ {
		// Validate candidate is within the allowed projects root
		candidateAbs, absErr := filepath.Abs(candidate)
		if absErr != nil {
			return "", "", fmt.Errorf("failed to resolve candidate path: %w", absErr)
		}
		candidateAbs = filepath.Clean(candidateAbs)

		// Security check: ensure candidate is a subdirectory of projectsRoot
		if !IsSafeSubdirectory(projectsRootAbs, candidateAbs) {
			return "", "", fmt.Errorf("project directory would be outside allowed projects root")
		}

		if mkErr := os.Mkdir(candidate, perm); mkErr == nil {
			// Double-check after creation - paranoid validation
			if !IsSafeSubdirectory(projectsRootAbs, candidateAbs) {
				// Security violation detected - remove the unsafe directory
				// We only reach here if somehow a directory was created outside the root
				// despite pre-checks. Clean up by removing ONLY if it's actually within root.
				if strings.HasPrefix(candidateAbs, projectsRootAbs+string(filepath.Separator)) {
					os.Remove(candidateAbs)
				}
				return "", "", fmt.Errorf("created directory is outside allowed projects root")
			}

			return candidate, folderName, nil
		} else if !os.IsExist(mkErr) {
			return "", "", mkErr
		}
		candidate = fmt.Sprintf("%s-%d", basePath, counter)
		folderName = fmt.Sprintf("%s-%d", sanitized, counter)
	}
}

func SanitizeProjectName(name string) string {
	name = strings.TrimSpace(name)
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
}

// IsSafeSubdirectory returns true if subdir is a subdirectory of baseDir (absolute, normalized)
func IsSafeSubdirectory(baseDir, subdir string) bool {
	absBase, err1 := filepath.Abs(baseDir)
	absSubdir, err2 := filepath.Abs(subdir)
	if err1 != nil || err2 != nil {
		return false
	}

	// Ensure both paths end consistently for comparison
	absBase = filepath.Clean(absBase)
	absSubdir = filepath.Clean(absSubdir)

	rel, err := filepath.Rel(absBase, absSubdir)
	if err != nil {
		return false
	}

	// The path must not escape the base directory
	return !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}

func SaveOrUpdateProjectFiles(projectsRoot, projectPath, composeContent string, envContent *string) error {
	return WriteProjectFiles(projectsRoot, projectPath, composeContent, envContent)
}
