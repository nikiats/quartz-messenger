<script lang="ts">
	import type { Snippet } from 'svelte';
	import { fade, scale } from 'svelte/transition';
	import Icon from './Icon.svelte';

	let {
		title,
		onclose,
		onback,
		footer,
		children,
	}: {
		title: string;
		onclose: () => void;
		onback?: () => void;
		footer?: Snippet;
		children: Snippet;
	} = $props();

	function keydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onclose();
		}
	}
</script>

<svelte:window onkeydown={keydown} />

<div class="overlay" transition:fade={{ duration: 180 }}>
	<button type="button" class="backdrop" aria-label="Закрыть" onclick={onclose}></button>

	<div class="sheet" role="dialog" aria-modal="true" aria-label={title} transition:scale={{ start: 0.96, duration: 200 }}>
		<header>
			{#if onback}
				<button type="button" class="icon-btn" aria-label="Назад" onclick={onback}>
					<Icon name="back" />
				</button>
			{/if}
			<h2 class:indent={!onback}>{title}</h2>
			<button type="button" class="icon-btn" aria-label="Закрыть" onclick={onclose}>
				<Icon name="close" size={22} />
			</button>
		</header>

		<div class="body">
			{@render children()}
		</div>

		{#if footer}
			<footer>{@render footer()}</footer>
		{/if}
	</div>
</div>

<style>
	.overlay {
		position: fixed;
		inset: 0;
		z-index: 5;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 16px;
	}

	.backdrop {
		position: absolute;
		inset: 0;
		border: none;
		background-color: rgba(4, 10, 17, 0.6);
		backdrop-filter: blur(6px);
		cursor: default;
	}

	.sheet {
		position: relative;
		display: flex;
		flex-direction: column;
		width: 100%;
		max-width: 440px;
		max-height: min(640px, 100%);
		border: 1px solid var(--border);
		border-radius: 28px;
		background-color: var(--bg);
		box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
		overflow: hidden;
	}

	header {
		display: flex;
		align-items: center;
		gap: 4px;
		height: 64px;
		padding: 0 10px 0 6px;
	}

	h2 {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		font-size: 20px;
		font-weight: 700;
		letter-spacing: -0.02em;
	}

	h2.indent {
		padding-left: 18px;
	}

	.icon-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		width: 44px;
		height: 44px;
		border: none;
		border-radius: 50%;
		color: var(--text);
		background: none;
		cursor: pointer;
		transition: background-color 0.2s var(--ease);
	}

	.icon-btn:hover {
		background-color: var(--hover);
	}

	.body {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
	}

	footer {
		padding: 12px 20px 20px;
	}
</style>
