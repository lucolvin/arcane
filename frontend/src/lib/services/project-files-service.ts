import BaseAPIService from './api-service';
import { environmentStore } from '$lib/stores/environment.store.svelte';
import type {
	BrowseFilesResponse,
	ReadFileResponse,
	WriteFileRequest,
	CreateDirectoryRequest,
	DeleteFileRequest,
	CopyFileRequest,
	MoveFileRequest
} from '$lib/types/project-files.type';

export class ProjectFilesService extends BaseAPIService {
	async browseFiles(projectId: string, path: string = ''): Promise<BrowseFilesResponse> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.get(`/environments/${envId}/projects/${projectId}/files`, {
				params: { path }
			})
		);
	}

	async readFile(projectId: string, path: string): Promise<ReadFileResponse> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.get(`/environments/${envId}/projects/${projectId}/files/read`, {
				params: { path }
			})
		);
	}

	async writeFile(projectId: string, request: WriteFileRequest): Promise<void> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.post(`/environments/${envId}/projects/${projectId}/files/write`, request)
		);
	}

	async createDirectory(projectId: string, request: CreateDirectoryRequest): Promise<void> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.post(`/environments/${envId}/projects/${projectId}/files/mkdir`, request)
		);
	}

	async deleteFile(projectId: string, request: DeleteFileRequest): Promise<void> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.delete(`/environments/${envId}/projects/${projectId}/files`, { data: request })
		);
	}

	async copyFile(projectId: string, request: CopyFileRequest): Promise<void> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.post(`/environments/${envId}/projects/${projectId}/files/copy`, request)
		);
	}

	async moveFile(projectId: string, request: MoveFileRequest): Promise<void> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(
			this.api.post(`/environments/${envId}/projects/${projectId}/files/move`, request)
		);
	}
}

export const projectFilesService = new ProjectFilesService();
