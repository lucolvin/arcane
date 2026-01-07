export interface FileEntry {
	name: string;
	path: string;
	isDirectory: boolean;
	size: number;
	modifiedAt: string;
	permissions: string;
}

export interface BrowseFilesResponse {
	entries: FileEntry[];
	path: string;
}

export interface ReadFileResponse {
	content: string;
	path: string;
	size: number;
}

export interface WriteFileRequest {
	path: string;
	content: string;
	createDirectories?: boolean;
}

export interface CreateDirectoryRequest {
	path: string;
}

export interface DeleteFileRequest {
	path: string;
	recursive?: boolean;
}

export interface CopyFileRequest {
	sourcePath: string;
	destinationPath: string;
}

export interface MoveFileRequest {
	sourcePath: string;
	destinationPath: string;
}
