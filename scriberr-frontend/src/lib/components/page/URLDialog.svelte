<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { LoaderCircle, Link as LinkIcon } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open: boolean;
		onDownload: (url: string, title: string) => Promise<void>;
	};

	let { open = $bindable(), onDownload }: Props = $props();

	let fileUrl = $state('');
	let title = $state('');
	let isDownloading = $state(false);

	function handleSubmit() {
		if (!fileUrl.trim()) {
			toast.error('Please enter a file URL');
			return;
		}

		if (!isValidUrl(fileUrl)) {
			toast.error('Please enter a valid URL (must start with http:// or https://)');
			return;
		}

		downloadFile();
	}

	async function downloadFile() {
		isDownloading = true;
		try {
			await onDownload(fileUrl, title || 'Audio from URL');
			// Reset form
			fileUrl = '';
			title = '';
			open = false;
		} catch (error) {
			console.error('File download error:', error);
		} finally {
			isDownloading = false;
		}
	}

	function isValidUrl(url: string): boolean {
		try {
			const urlObj = new URL(url);
			return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
		} catch {
			return false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !isDownloading) {
			handleSubmit();
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="border-none bg-gray-700 text-gray-200 sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<LinkIcon class="h-5 w-5 text-blue-400" />
				Download Audio from URL
			</Dialog.Title>
			<Dialog.Description class="text-gray-400">
				Enter a direct URL to an audio file to download and transcribe it.
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			<div class="space-y-2">
				<label for="file-url" class="text-sm font-medium text-gray-300">File URL *</label>
				<Input
					id="file-url"
					type="url"
					placeholder="https://example.com/audio.mp3"
					bind:value={fileUrl}
					onkeydown={handleKeydown}
					disabled={isDownloading}
				/>
			</div>

			<div class="space-y-2">
				<label for="title" class="text-sm font-medium text-gray-300">Title (optional)</label>
				<Input
					id="title"
					type="text"
					placeholder="Enter a title for this audio"
					bind:value={title}
					onkeydown={handleKeydown}
					disabled={isDownloading}
				/>
			</div>
		</div>

		<Dialog.Footer>
			<Button
				variant="outline"
				onclick={() => (open = false)}
				disabled={isDownloading}
				class="border-gray-600 text-gray-300 hover:bg-gray-600 hover:text-gray-100"
			>
				Cancel
			</Button>
			<Button
				onclick={handleSubmit}
				disabled={isDownloading || !fileUrl.trim()}
				class="bg-blue-500 hover:bg-blue-600"
			>
				{#if isDownloading}
					<LoaderCircle class="mr-2 h-4 w-4 animate-spin" />
					Downloading...
				{:else}
					<LinkIcon class="mr-2 h-4 w-4" />
					Download
				{/if}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
