<script lang="ts">
	import type { ComponentProps } from 'svelte';
	import { link } from 'svelte-spa-router';
	import Icon from './Icon.svelte';

	let {
		icon,
		label,
		href,
		onclick,
	}: {
		icon: ComponentProps<typeof Icon>['name'];
		label: string;
		href?: string;
		onclick?: () => void;
	} = $props();
</script>

{#if href}
	<a {href} class="icon-btn" aria-label={label} use:link><Icon name={icon} /></a>
{:else}
	<button type="button" class="icon-btn" aria-label={label} {onclick}><Icon name={icon} /></button>
{/if}

<style>
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

	.icon-btn:active {
		background-color: var(--active);
	}
</style>
