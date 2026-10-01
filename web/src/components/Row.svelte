<script lang="ts">
	import type { Snippet } from 'svelte';
	import { link } from 'svelte-spa-router';
	import Avatar from './Avatar.svelte';

	let {
		href,
		onclick,
		id,
		title,
		subtitle,
		meta,
		unread = false,
		trailing,
	}: {
		href?: string;
		onclick?: () => void;
		id: string;
		title: string;
		subtitle?: string;
		meta?: string;
		unread?: boolean;
		trailing?: Snippet;
	} = $props();
</script>

{#snippet body()}
	<Avatar name={title} {id} size={52} />
	<span class="line">
		<span class="text">
			<span class="top">
				<span class="title">{title}</span>
				{#if meta}<span class="meta" class:unread>{meta}</span>{/if}
			</span>
			<span class="bottom">
				<span class="subtitle">{subtitle ?? ''}</span>
				{#if unread}<span class="dot" aria-label="Есть непрочитанные"></span>{/if}
			</span>
		</span>
		{@render trailing?.()}
	</span>
{/snippet}

{#if href}
	<a {href} class="row interactive" use:link>{@render body()}</a>
{:else if onclick}
	<button type="button" class="row interactive" {onclick}>{@render body()}</button>
{:else}
	<div class="row">{@render body()}</div>
{/if}

<style>
	.row {
		display: flex;
		align-items: center;
		gap: 14px;
		width: 100%;
		padding: 0 16px;
		border: none;
		background: none;
		text-align: left;
		text-decoration: none;
		color: var(--text);
		transition: background-color 0.2s var(--ease);
	}

	.interactive {
		cursor: pointer;
	}

	.interactive:hover {
		background-color: var(--hover);
	}

	.interactive:active {
		background-color: var(--active);
	}

	.line {
		display: flex;
		align-items: center;
		gap: 12px;
		flex: 1;
		min-width: 0;
		height: 76px;
		border-bottom: 1px solid var(--divider);
	}

	.row:last-child .line {
		border-bottom: none;
	}

	.text {
		display: flex;
		flex-direction: column;
		gap: 4px;
		flex: 1;
		min-width: 0;
	}

	.top {
		display: flex;
		align-items: baseline;
		gap: 8px;
	}

	.title {
		flex: 1;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		font-size: 16px;
		font-weight: 600;
	}

	.meta,
	.subtitle {
		font-size: 14px;
		color: var(--text-muted);
	}

	.meta.unread {
		color: var(--primary);
	}

	.bottom {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.subtitle {
		flex: 1;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.dot {
		flex-shrink: 0;
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background-color: var(--primary);
		box-shadow: 0 0 10px rgba(224, 225, 221, 0.55);
	}
</style>
