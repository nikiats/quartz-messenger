<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="row">
	{@render children()}
</div>

<style>
	.row {
		display: grid;
		grid-auto-flow: column;
		grid-auto-columns: 1fr;
		gap: 10px;
	}
</style>
