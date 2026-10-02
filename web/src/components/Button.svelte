<script lang="ts">
	import type { Snippet } from 'svelte';
	import { link } from 'svelte-spa-router';

	let {
		href,
		type = 'button',
		variant = 'primary',
		small = false,
		disabled = false,
		form,
		onclick,
		children,
	}: {
		href?: string;
		type?: 'button' | 'submit';
		form?: string;
		variant?: 'primary' | 'secondary' | 'danger';
		small?: boolean;
		disabled?: boolean;
		onclick?: () => unknown;
		children: Snippet;
	} = $props();

	let busy = $state(false);

	async function click() {
		if (!onclick) {
			return;
		}

		busy = true;
		try {
			await onclick();
		} finally {
			busy = false;
		}
	}
</script>

{#if href}
	<a {href} class="btn {variant}" class:small use:link>{@render children()}</a>
{:else}
	<button {type} {form} class="btn {variant}" class:small disabled={disabled || busy} onclick={click}>
		{@render children()}
	</button>
{/if}

<style>
	.btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		width: 100%;
		height: 52px;
		padding: 0 24px;
		border: none;
		border-radius: 999px;
		font-size: 16px;
		font-weight: 600;
		letter-spacing: -0.01em;
		white-space: nowrap;
		text-decoration: none;
		cursor: pointer;
		transition:
			transform 0.15s var(--ease),
			filter 0.2s var(--ease),
			background-color 0.2s var(--ease);
	}

	.btn:active {
		transform: scale(0.98);
	}

	.btn:disabled {
		opacity: 0.55;
		cursor: default;
		transform: none;
		filter: none;
	}

	.small {
		flex-shrink: 0;
		width: auto;
		height: 34px;
		padding: 0 16px;
		font-size: 14px;
	}

	.primary {
		color: var(--on-primary);
		background-color: var(--primary);
	}

	.primary:hover {
		filter: brightness(1.06);
	}

	.secondary {
		color: var(--text);
		background-color: var(--surface);
	}

	.secondary:hover,
	.danger:hover {
		background-color: var(--accent);
	}

	.danger {
		color: #E8837E;
		background-color: var(--surface);
	}
</style>
