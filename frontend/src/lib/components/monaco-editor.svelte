<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { monaco, initShiki } from './monaco-code-editor/monaco';
	import { mode } from 'mode-watcher';

	type CodeLanguage = 'yaml' | 'env' | 'javascript' | 'typescript' | 'json' | 'html' | 'css' | 'xml' | 'python' | 'go' | 'bash' | 'sql' | 'plaintext';

	let {
		value = $bindable(''),
		language = 'plaintext' as CodeLanguage,
		readonly = false,
		fontSize = '12px',
		height = '100%'
	}: {
		value: string;
		language: CodeLanguage;
		readonly?: boolean;
		fontSize?: string;
		height?: string;
	} = $props();

	let editorElement = $state<HTMLDivElement>();
	let editor = $state.raw<monaco.editor.IStandaloneCodeEditor | null>(null);
	let model = $state.raw<monaco.editor.ITextModel | null>(null);
	let changeDisposable: monaco.IDisposable | null = null;

	// Map our language names to Monaco language IDs
	const languageMap: Record<CodeLanguage, string> = {
		javascript: 'javascript',
		typescript: 'typescript',
		json: 'json',
		html: 'html',
		css: 'css',
		xml: 'xml',
		yaml: 'yaml',
		env: 'ini',
		python: 'python',
		go: 'go',
		bash: 'bash',
		sql: 'sql',
		plaintext: 'plaintext'
	};

	const langId = $derived(languageMap[language] || 'plaintext');
	const theme = $derived(mode.current === 'dark' ? 'catppuccin-mocha' : 'catppuccin-latte');

	onMount(async () => {
		if (!editorElement) return;

		await initShiki(monaco);

		// Wait for container to be properly sized
		await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve)));

		// Create model
		const uri = monaco.Uri.parse(`inmemory://model-${Date.now()}.${langId}`);
		model = monaco.editor.createModel(value, langId, uri);

		editor = monaco.editor.create(editorElement, {
			model: model,
			automaticLayout: true,
			theme,
			readOnly: readonly,
			fontSize: parseInt(fontSize.replace('px', '')),
			minimap: { enabled: true },
			scrollBeyondLastLine: false,
			wordWrap: 'on',
			fixedOverflowWidgets: true,
			dragAndDrop: false,
			contextmenu: true,
			quickSuggestions: {
				other: true,
				comments: false,
				strings: true
			},
			suggestOnTriggerCharacters: true,
			fontFamily:
				'"Geist Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
			padding: { top: 10, bottom: 10 }
		});

		changeDisposable = model.onDidChangeContent(() => {
			value = model?.getValue() || '';
		});
	});

	onDestroy(() => {
		changeDisposable?.dispose();
		editor?.dispose();
		model?.dispose();
	});

	// Sync value to model
	$effect(() => {
		if (model && value !== model.getValue()) {
			model.setValue(value);
		}
	});

	// Sync language
	$effect(() => {
		if (model) {
			monaco.editor.setModelLanguage(model, langId);
		}
	});

	// Sync options and layout
	$effect(() => {
		if (editor) {
			editor.updateOptions({
				readOnly: readonly,
				theme,
				fontSize: parseInt(fontSize.replace('px', ''))
			});
			editor.layout();
		}
	});

	// Global theme sync
	$effect(() => {
		monaco.editor.setTheme(theme);
	});

	export function getValue() {
		return model?.getValue() || '';
	}
</script>

<div style="height: {height};" class="w-full min-h-0 overflow-hidden rounded border" bind:this={editorElement}></div>
