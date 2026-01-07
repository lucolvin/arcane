<script lang="ts">
	import { ArcaneButton } from '$lib/components/arcane-button/index.js';
	import * as Card from '$lib/components/ui/card';
	import * as TreeView from '$lib/components/ui/tree-view/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import MonacoEditor from '$lib/components/monaco-editor.svelte';
	import {
		FolderIcon,
		FileTextIcon,
		PlusIcon,
		RefreshIcon,
		TrashIcon,
		CopyIcon,
		EditIcon,
		DownloadIcon,
		FolderPlusIcon,
		EllipsisIcon
	} from '$lib/icons';
	import { toast } from 'svelte-sonner';
	import { projectFilesService } from '$lib/services/project-files-service';
	import { handleApiResultWithCallbacks } from '$lib/utils/api.util';
	import { tryCatch } from '$lib/utils/try-catch';
	import type { FileEntry } from '$lib/types/project-files.type';

	let { projectId }: { projectId: string } = $props();

	let currentPath = $state('');
	let files = $state<FileEntry[]>([]);
	let selectedFile = $state<FileEntry | null>(null);
	let isLoading = $state(false);
	let fileContent = $state('');
	let isEditing = $state(false);
	let editorRef: any = $state(null);

	let showCreateDialog = $state(false);
	let showRenameDialog = $state(false);
	let showCopyDialog = $state(false);
	let createType = $state<'file' | 'folder'>('file');
	let createName = $state('');
	let renameName = $state('');
	let copyDestination = $state('');

	$effect(() => {
		if (projectId) {
			loadFiles();
		}
	});

	async function loadFiles(path: string = currentPath) {
		isLoading = true;
		try {
			const response = await projectFilesService.browseFiles(projectId, path);
			const entries = Array.isArray(response?.entries) ? response.entries : [];
			files = entries.sort((a, b) => {
				if (a.isDirectory && !b.isDirectory) return -1;
				if (!a.isDirectory && b.isDirectory) return 1;
				return a.name.localeCompare(b.name);
			});
			currentPath = response?.path ?? path ?? '';
		} catch (err) {
			toast.error('Failed to load files');
			console.error(err);
		} finally {
			isLoading = false;
		}
	}

	async function handleFileClick(file: FileEntry) {
		if (file.isDirectory) {
			currentPath = file.path;
			await loadFiles(file.path);
			selectedFile = null;
		} else {
			selectedFile = file;
			await loadFileContent(file.path);
		}
	}

	async function loadFileContent(path: string) {
		isLoading = true;
		try {
			const response = await projectFilesService.readFile(projectId, path);
			fileContent = response.content;
			isEditing = false;
		} catch (err) {
			toast.error('Failed to read file');
			console.error(err);
		} finally {
			isLoading = false;
		}
	}

	async function saveFile() {
		if (!selectedFile || !editorRef) return;

		const content = editorRef.getValue();

		handleApiResultWithCallbacks({
			result: await tryCatch(
				projectFilesService.writeFile(projectId, {
					path: selectedFile.path,
					content
				})
			),
			message: 'Failed to save file',
			onSuccess: () => {
				toast.success('File saved successfully');
				isEditing = false;
				loadFiles();
			}
		});
	}

	function getFileLanguage(fileName: string): string {
		const ext = fileName.split('.').pop()?.toLowerCase() ?? '';
		const languageMap: Record<string, string> = {
			js: 'javascript',
			ts: 'typescript',
			jsx: 'jsx',
			tsx: 'typescript',
			json: 'json',
			yaml: 'yaml',
			yml: 'yaml',
			html: 'html',
			xml: 'xml',
			css: 'css',
			scss: 'scss',
			less: 'less',
			sql: 'sql',
			sh: 'shell',
			bash: 'shell',
			python: 'python',
			py: 'python',
			java: 'java',
			go: 'go',
			rs: 'rust',
			rb: 'ruby',
			php: 'php',
			c: 'c',
			cpp: 'cpp',
			cs: 'csharp',
			dockerfile: 'dockerfile',
			env: 'plaintext',
			md: 'markdown',
			txt: 'plaintext'
		};
		return languageMap[ext] ?? 'plaintext';
	}

	async function createNewItem() {
		const path = currentPath ? `${currentPath}/${createName}` : createName;

		if (createType === 'folder') {
			handleApiResultWithCallbacks({
				result: await tryCatch(projectFilesService.createDirectory(projectId, { path })),
				message: 'Failed to create folder',
				onSuccess: () => {
					toast.success('Folder created successfully');
					showCreateDialog = false;
					createName = '';
					loadFiles();
				}
			});
		} else {
			handleApiResultWithCallbacks({
				result: await tryCatch(
					projectFilesService.writeFile(projectId, {
						path,
						content: '',
						createDirectories: true
					})
				),
				message: 'Failed to create file',
				onSuccess: () => {
					toast.success('File created successfully');
					showCreateDialog = false;
					createName = '';
					loadFiles();
				}
			});
		}
	}

	async function deleteFile(file: FileEntry) {
		const confirmed = await import('$lib/components/confirm-dialog').then((m) =>
			m.openConfirmDialog({
				title: 'Delete File',
				message: `Are you sure you want to delete "${file.name}"?${file.isDirectory ? ' This will delete all contents.' : ''}`,
				confirmText: 'Delete',
				cancelText: 'Cancel',
				variant: 'destructive'
			})
		);

		if (!confirmed) return;

		handleApiResultWithCallbacks({
			result: await tryCatch(
				projectFilesService.deleteFile(projectId, {
					path: file.path,
					recursive: file.isDirectory
				})
			),
			message: 'Failed to delete',
			onSuccess: () => {
				toast.success('Deleted successfully');
				if (selectedFile?.path === file.path) {
					selectedFile = null;
				}
				loadFiles();
			}
		});
	}

	async function copyFile(file: FileEntry) {
		const destPath = currentPath ? `${currentPath}/${copyDestination}` : copyDestination;

		handleApiResultWithCallbacks({
			result: await tryCatch(
				projectFilesService.copyFile(projectId, {
					sourcePath: file.path,
					destinationPath: destPath
				})
			),
			message: 'Failed to copy',
			onSuccess: () => {
				toast.success('Copied successfully');
				showCopyDialog = false;
				copyDestination = '';
				loadFiles();
			}
		});
	}

	async function renameFile(file: FileEntry) {
		const dir = file.path.includes('/') ? file.path.substring(0, file.path.lastIndexOf('/')) : '';
		const destPath = dir ? `${dir}/${renameName}` : renameName;

		handleApiResultWithCallbacks({
			result: await tryCatch(
				projectFilesService.moveFile(projectId, {
					sourcePath: file.path,
					destinationPath: destPath
				})
			),
			message: 'Failed to rename',
			onSuccess: () => {
				toast.success('Renamed successfully');
				showRenameDialog = false;
				renameName = '';
				if (selectedFile?.path === file.path) {
					selectedFile = null;
				}
				loadFiles();
			}
		});
	}

	function downloadFile() {
		if (!selectedFile) return;

		const blob = new Blob([fileContent], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = selectedFile.name;
		a.click();
		URL.revokeObjectURL(url);
	}

	function copyFilePath() {
		if (!selectedFile) return;

		navigator.clipboard.writeText(selectedFile.path).then(() => {
			toast.success('File path copied to clipboard');
		}).catch(() => {
			toast.error('Failed to copy file path');
		});
	}

	function navigateUp() {
		if (currentPath.includes('/')) {
			const newPath = currentPath.substring(0, currentPath.lastIndexOf('/'));
			currentPath = newPath;
			loadFiles(newPath);
		} else if (currentPath !== '') {
			currentPath = '';
			loadFiles('');
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString();
	}
</script>

<div class="flex h-full gap-4">
	<Card.Root class="flex w-96 flex-col">
		<Card.Header class="flex-shrink-0">
			<div class="flex items-center justify-between">
				<Card.Title>
					<div class="flex items-center gap-2">
						<FolderIcon class="size-5" />
						<span>Files</span>
					</div>
				</Card.Title>
				<div class="flex gap-1">
					<ArcaneButton
						action="base"
						tone="ghost"
						size="icon-sm"
						icon={RefreshIcon}
						onclick={() => loadFiles()}
						disabled={isLoading}
						showLabel={false}
						customLabel="Refresh"
					/>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<ArcaneButton
									{...props}
									action="base"
									tone="ghost"
									size="icon-sm"
									icon={PlusIcon}
									showLabel={false}
									customLabel="Create"
								/>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content>
							<DropdownMenu.Item
								onclick={() => {
									createType = 'file';
									showCreateDialog = true;
								}}
							>
								<FileTextIcon class="size-4" />
								New File
							</DropdownMenu.Item>
							<DropdownMenu.Item
								onclick={() => {
									createType = 'folder';
									showCreateDialog = true;
								}}
							>
								<FolderPlusIcon class="size-4" />
								New Folder
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			</div>
			{#if currentPath}
				<div class="mt-2 flex items-center gap-2">
					<ArcaneButton
						action="base"
						tone="ghost"
						size="sm"
						onclick={navigateUp}
						customLabel="← Back"
					/>
					<span class="text-muted-foreground truncate text-sm">/{currentPath}</span>
				</div>
			{/if}
		</Card.Header>
		<Card.Content class="min-h-0 flex-1 overflow-y-auto">
			{#if isLoading}
				<div class="flex items-center justify-center py-8">
					<div
						class="size-6 animate-spin rounded-full border-2 border-primary border-t-transparent"
					></div>
				</div>
			{:else if files.length === 0}
				<div class="text-muted-foreground py-8 text-center text-sm">No files found</div>
			{:else}
				<TreeView.Root>
					{#each files as file}
						<TreeView.File
							name={file.name}
							onclick={() => handleFileClick(file)}
							class={selectedFile?.path === file.path ? 'bg-accent' : ''}
						>
							{#snippet icon()}
								{#if file.isDirectory}
									<FolderIcon class="size-4 text-amber-500" />
								{:else}
									<FileTextIcon class="size-4 text-blue-500" />
								{/if}
							{/snippet}
							{#snippet actions()}
								<DropdownMenu.Root>
									<DropdownMenu.Trigger>
										{#snippet child({ props })}
											<button
												{...props}
												class="rounded p-1 hover:bg-accent"
												onclick={(e) => {
													e.stopPropagation();
												}}
											>
												<EllipsisIcon class="size-4" />
											</button>
										{/snippet}
									</DropdownMenu.Trigger>
									<DropdownMenu.Content>
										<DropdownMenu.Item
											onclick={() => {
												selectedFile = file;
												renameName = file.name;
												showRenameDialog = true;
											}}
										>
											<EditIcon class="size-4" />
											Rename
										</DropdownMenu.Item>
										<DropdownMenu.Item
											onclick={() => {
												selectedFile = file;
												copyDestination = `${file.name}-copy`;
												showCopyDialog = true;
											}}
										>
											<CopyIcon class="size-4" />
											Copy
										</DropdownMenu.Item>
										<DropdownMenu.Item onclick={() => deleteFile(file)}>
											<TrashIcon class="size-4" />
											Delete
										</DropdownMenu.Item>
									</DropdownMenu.Content>
								</DropdownMenu.Root>
							{/snippet}
						</TreeView.File>
					{/each}
				</TreeView.Root>
			{/if}
		</Card.Content>
	</Card.Root>

	<div class="flex flex-1 flex-col">
		{#if selectedFile}
			<div class="mb-4 flex items-center justify-between">
				<div>
					<h3 class="font-semibold">{selectedFile.name}</h3>
					<p class="text-muted-foreground text-sm">
						{formatFileSize(selectedFile.size)} • Modified {formatDate(selectedFile.modifiedAt)}
					</p>
				</div>
				<div class="flex gap-2">
					<ArcaneButton
						action="base"
						tone="ghost"
						size="sm"
						icon={DownloadIcon}
						onclick={downloadFile}
						customLabel="Download"
					/>
					<ArcaneButton
						action="base"
						tone="ghost"
						size="sm"
						icon={CopyIcon}
						onclick={copyFilePath}
						customLabel="Copy Path"
					/>
					{#if isEditing}
						<ArcaneButton
							action="base"
							tone="ghost"
							size="sm"
							onclick={() => {
								isEditing = false;
								loadFileContent(selectedFile.path);
							}}
							customLabel="Cancel"
						/>
						<ArcaneButton action="update" size="sm" onclick={saveFile} customLabel="Save" />
					{:else}
						<ArcaneButton
							action="base"
							tone="outline"
							size="sm"
							icon={EditIcon}
							onclick={() => (isEditing = true)}
							customLabel="Edit"
						/>
					{/if}
				</div>
			</div>
			<div class="flex-1 overflow-hidden">
				<MonacoEditor
					bind:this={editorRef}
					value={fileContent}
					language={getFileLanguage(selectedFile.name)}
					readonly={!isEditing}
					height="100%"
					theme="vs-dark"
				/>
			</div>
		{:else}
			<div class="text-muted-foreground flex h-full items-center justify-center">
				<div class="text-center">
					<FileTextIcon class="mx-auto size-12 opacity-50" />
					<p class="mt-2">Select a file to view or edit</p>
				</div>
			</div>
		{/if}
	</div>
</div>

<Dialog.Root bind:open={showCreateDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Create {createType === 'file' ? 'File' : 'Folder'}</Dialog.Title>
		</Dialog.Header>
		<div class="space-y-4">
			<div>
				<Label for="create-name">Name</Label>
				<Input
					id="create-name"
					bind:value={createName}
					placeholder={createType === 'file' ? 'filename.txt' : 'folder-name'}
				/>
			</div>
			{#if currentPath}
				<p class="text-muted-foreground text-sm">Will be created in: /{currentPath}</p>
			{/if}
		</div>
		<Dialog.Footer>
			<ArcaneButton
				action="base"
				tone="ghost"
				onclick={() => {
					showCreateDialog = false;
					createName = '';
				}}
				customLabel="Cancel"
			/>
			<ArcaneButton
				action="create"
				onclick={createNewItem}
				disabled={!createName}
				customLabel="Create"
			/>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showRenameDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Rename</Dialog.Title>
		</Dialog.Header>
		<div class="space-y-4">
			<div>
				<Label for="rename-name">New name</Label>
				<Input id="rename-name" bind:value={renameName} />
			</div>
		</div>
		<Dialog.Footer>
			<ArcaneButton
				action="base"
				tone="ghost"
				onclick={() => {
					showRenameDialog = false;
					renameName = '';
				}}
				customLabel="Cancel"
			/>
			<ArcaneButton
				action="update"
				onclick={() => selectedFile && renameFile(selectedFile)}
				disabled={!renameName}
				customLabel="Rename"
			/>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showCopyDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Copy</Dialog.Title>
		</Dialog.Header>
		<div class="space-y-4">
			<div>
				<Label for="copy-dest">Destination name</Label>
				<Input id="copy-dest" bind:value={copyDestination} />
			</div>
			{#if currentPath}
				<p class="text-muted-foreground text-sm">Will be copied to: /{currentPath}</p>
			{/if}
		</div>
		<Dialog.Footer>
			<ArcaneButton
				action="base"
				tone="ghost"
				onclick={() => {
					showCopyDialog = false;
					copyDestination = '';
				}}
				customLabel="Cancel"
			/>
			<ArcaneButton
				action="update"
				onclick={() => selectedFile && copyFile(selectedFile)}
				disabled={!copyDestination}
				customLabel="Copy"
			/>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
