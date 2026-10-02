<script lang="ts">
	import type { Snippet } from 'svelte';
	import IconButton from './IconButton.svelte';

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
			<IconButton href={back} icon="back" label="Назад" />
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
