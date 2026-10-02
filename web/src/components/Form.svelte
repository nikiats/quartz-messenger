<script lang="ts">
	import type { Snippet } from 'svelte';

	let { onsubmit, children }: { onsubmit: () => unknown; children: Snippet } = $props();

	let busy = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();

		busy = true;
		try {
			await onsubmit();
		} finally {
			busy = false;
		}
	}
</script>

<form onsubmit={submit}>
	<fieldset disabled={busy}>
		{@render children()}
	</fieldset>
</form>

<style>
	fieldset {
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
		border: none;
	}
</style>
