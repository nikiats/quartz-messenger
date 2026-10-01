<script lang="ts">
	import type { Snippet } from 'svelte';
	import { link } from 'svelte-spa-router';
	import Icon from './Icon.svelte';

	let {
		title = '',
		back,
		padded = true,
		heading,
		actions,
		footer,
		children,
	}: {
		title?: string;
		back?: string;
		padded?: boolean;
		heading?: Snippet;
		actions?: Snippet;
		footer?: Snippet;
		children: Snippet;
	} = $props();
</script>

<section class="page">
	<header>
		{#if back}
			<a href={back} class="back" use:link aria-label="Назад">
				<Icon name="back" />
			</a>
		{/if}
		<div class="heading" class:indent={!back}>
			{#if heading}
				{@render heading()}
			{:else}
				<h1>{title}</h1>
			{/if}
		</div>
		{@render actions?.()}
	</header>

	<div class="content" class:padded>
		{@render children()}
	</div>

	{#if footer}
		<footer>{@render footer()}</footer>
	{/if}
</section>

<style>
	.page {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
	}

	header {
		display: flex;
		align-items: center;
		gap: 4px;
		height: 64px;
		padding: 0 12px 0 6px;
		border-bottom: 1px solid var(--divider);
	}

	.back {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		border-radius: 50%;
		color: var(--text);
		transition: background-color 0.2s var(--ease);
	}

	.back:hover {
		background-color: var(--hover);
	}

	.back:active {
		background-color: var(--active);
	}

	.heading {
		display: flex;
		align-items: center;
		flex: 1;
		min-width: 0;
	}

	.heading.indent {
		padding-left: 14px;
	}

	h1 {
		font-size: 22px;
		font-weight: 700;
		letter-spacing: -0.02em;
	}

	.content {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
	}

	.content.padded {
		gap: 16px;
		padding: 24px 20px;
	}

	footer {
		padding: 12px 16px 20px;
		border-top: 1px solid var(--divider);
	}
</style>
