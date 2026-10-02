<script lang="ts">
	import { slide } from 'svelte/transition';
	import { auth } from '../lib/auth.svelte.js';
	import { connection } from '../lib/ws.svelte.js';
</script>

{#if auth.session && connection.status === 'offline'}
	<p class="banner" transition:slide={{ duration: 200 }}>Нет соединения в реальном времени · возможны задержки</p>
{/if}

<style>
	.banner {
		padding: 7px 16px;
		font-size: 13px;
		font-weight: 500;
		text-align: center;
		color: var(--text);
		background-color: var(--warning);
	}
</style>
