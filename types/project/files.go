package project

import "time"

// FileEntry represents a file or directory in the project filesystem.
type FileEntry struct {
	// Name is the name of the file or directory.
	//
	// Required: true
	Name string `json:"name"`

	// Path is the relative path from the project root.
	//
	// Required: true
	Path string `json:"path"`

	// IsDirectory indicates if this entry is a directory.
	//
	// Required: true
	IsDirectory bool `json:"isDirectory"`

	// Size is the file size in bytes (0 for directories).
	//
	// Required: true
	Size int64 `json:"size"`

	// ModifiedAt is the last modification time.
	//
	// Required: true
	ModifiedAt time.Time `json:"modifiedAt"`

	// Permissions is the file permissions in octal format.
	//
	// Required: true
	Permissions string `json:"permissions"`
}

// BrowseFilesResponse contains the list of files and directories.
type BrowseFilesResponse struct {
	// Entries is the list of files and directories.
	//
	// Required: true
	Entries []FileEntry `json:"entries"`

	// Path is the current directory path.
	//
	// Required: true
	Path string `json:"path"`
}

// ReadFileResponse contains the file content.
type ReadFileResponse struct {
	// Content is the file content.
	//
	// Required: true
	Content string `json:"content"`

	// Path is the file path.
	//
	// Required: true
	Path string `json:"path"`

	// Size is the file size in bytes.
	//
	// Required: true
	Size int64 `json:"size"`
}

// WriteFileRequest is used to write content to a file.
type WriteFileRequest struct {
	// Path is the relative path to the file from the project root.
	//
	// Required: true
	Path string `json:"path" binding:"required"`

	// Content is the file content to write.
	//
	// Required: true
	Content string `json:"content" binding:"required"`

	// CreateDirectories indicates if parent directories should be created.
	//
	// Required: false
	CreateDirectories bool `json:"createDirectories,omitempty"`
}

// CreateDirectoryRequest is used to create a new directory.
type CreateDirectoryRequest struct {
	// Path is the relative path to the directory from the project root.
	//
	// Required: true
	Path string `json:"path" binding:"required"`
}

// DeleteFileRequest is used to delete a file or directory.
type DeleteFileRequest struct {
	// Path is the relative path to the file/directory from the project root.
	//
	// Required: true
	Path string `json:"path" binding:"required"`

	// Recursive indicates if directories should be deleted recursively.
	//
	// Required: false
	Recursive bool `json:"recursive,omitempty"`
}

// CopyFileRequest is used to copy a file or directory.
type CopyFileRequest struct {
	// SourcePath is the source file/directory path relative to project root.
	//
	// Required: true
	SourcePath string `json:"sourcePath" binding:"required"`

	// DestinationPath is the destination path relative to project root.
	//
	// Required: true
	DestinationPath string `json:"destinationPath" binding:"required"`
}

// MoveFileRequest is used to move/rename a file or directory.
type MoveFileRequest struct {
	// SourcePath is the source file/directory path relative to project root.
	//
	// Required: true
	SourcePath string `json:"sourcePath" binding:"required"`

	// DestinationPath is the destination path relative to project root.
	//
	// Required: true
	DestinationPath string `json:"destinationPath" binding:"required"`
}
