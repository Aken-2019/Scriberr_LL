<script lang="ts">
	import { Dialog as DialogPrimitive } from "bits-ui";
	import XIcon from "@lucide/svelte/icons/x";
	import type { Snippet } from "svelte";
	import * as Dialog from "./index.js";
	import { cn, type WithoutChildrenOrChild } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		portalProps,
		children,
		showCloseButton = true,
		...restProps
	}: WithoutChildrenOrChild<DialogPrimitive.ContentProps> & {
		portalProps?: DialogPrimitive.PortalProps;
		children: Snippet;
		showCloseButton?: boolean;
	} = $props();
</script>

<Dialog.Portal {...portalProps}>
	<Dialog.Overlay />
	<DialogPrimitive.Content
		bind:ref
		data-slot="dialog-content"
		class={cn(
			"bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed left-[50%] top-4 bottom-4 z-50 flex max-h-[calc(100vh-2rem)] w-full max-w-[calc(100%-2rem)] -translate-x-1/2 flex-col gap-4 overflow-y-auto rounded-lg border p-6 shadow-lg duration-200 sm:top-[50%] sm:bottom-auto sm:max-h-[calc(100vh-4rem)] sm:-translate-y-1/2 sm:max-w-lg",
			className
		)}
		{...restProps}
	>
		<div class="relative flex-1">
			{@render children?.()}
			{#if showCloseButton}
				<DialogPrimitive.Close
					class="ring-offset-background focus:ring-ring bg-background/80 backdrop-blur-sm rounded-full p-1 absolute -right-2 -top-2 opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:pointer-events-none [&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none [&_svg]:shrink-0"
				>
					<XIcon class="h-5 w-5" />
					<span class="sr-only">Close</span>
				</DialogPrimitive.Close>
			{/if}
		</div>
	</DialogPrimitive.Content>
</Dialog.Portal>
