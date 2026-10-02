<script lang="ts">
	import Avatar from './Avatar.svelte';

	let { id, name = '', caption = '' }: { id?: string; name?: string; caption?: string } = $props();
</script>

<div class="card">
	<Avatar {name} {id} size={88} />
	<p class="name">{name}</p>
	{#if caption}
		<p class="caption">{caption}</p>
	{/if}
</div>

<style>
	.card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 8px 20px 20px;
	}

	.name {
		margin-top: 10px;
		font-size: 20px;
		font-weight: 600;
		text-align: center;
	}

	.caption {
		color: var(--text-muted);
	}
</style>
