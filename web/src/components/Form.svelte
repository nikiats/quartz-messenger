<script lang="ts">
	import type { Snippet } from 'svelte';

	let { onsubmit, children }: { onsubmit: () => void; children: Snippet } = $props();

	function submit(e: SubmitEvent) {
		e.preventDefault();
		onsubmit();
	}
</script>

<form onsubmit={submit}>
	{@render children()}
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
</style>
