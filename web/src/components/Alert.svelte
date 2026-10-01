<script lang="ts">
	import { fly } from 'svelte/transition';
	import { alertBox } from '../lib/alert.svelte.js';
</script>

{#if alertBox.current}
	<p class="alert {alertBox.current.kind}" transition:fly={{ y: -16, duration: 200 }}>{alertBox.current.text}</p>
{/if}

<style>
	.alert {
		position: fixed;
		top: 16px;
		left: 0;
		right: 0;
		z-index: 10;
		width: fit-content;
		max-width: calc(100% - 32px);
		margin: 0 auto;
		padding: 12px 22px;
		border-radius: 999px;
		font-size: 15px;
		font-weight: 500;
		color: var(--text);
		box-shadow: 0 8px 28px rgba(0, 0, 0, 0.35);
	}

	.info {
		background-color: var(--info);
	}

	.success {
		background-color: var(--success);
	}

	.error {
		background-color: var(--error);
	}
</style>
