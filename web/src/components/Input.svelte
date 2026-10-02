<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	let { value = $bindable(''), autofocus = false, ...rest }: HTMLInputAttributes = $props();

	let input: HTMLInputElement;

	$effect(() => {
		if (autofocus) {
			input.focus();
		}
	});
</script>

<input bind:this={input} bind:value {...rest} />

<style>
	input {
		width: 100%;
		height: 52px;
		padding: 0 20px;
		font-size: 16px;
		color: var(--text);
		background-color: var(--surface);
		border: 1px solid transparent;
		border-radius: 999px;
		outline: none;
		transition:
			border-color 0.2s var(--ease),
			box-shadow 0.2s var(--ease);
	}

	input::placeholder {
		color: var(--text-muted);
	}

	input:focus {
		border-color: var(--text-muted);
		box-shadow: 0 0 0 4px var(--hover);
	}

	input:disabled {
		opacity: 0.55;
	}
</style>
