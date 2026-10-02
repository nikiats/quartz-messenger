<script lang="ts">
	import type { Snippet } from 'svelte';
	import Input from './Input.svelte';
	import { auth } from '../lib/auth.svelte.js';
	import { searchUsers, type User } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';
	import { userSearchDelay } from '../config/search.js';

	let { row }: { row: Snippet<[User]> } = $props();

	let query = $state('');
	let users = $state<User[] | null>(null);

	$effect(() => {
		const q = query.trim();
		if (!q) {
			users = null;
			return;
		}

		const timer = setTimeout(() => search(q), userSearchDelay);
		return () => clearTimeout(timer);
	});

	async function search(q: string) {
		try {
			const found = await searchUsers(q);
			if (q !== query.trim()) {
				return;
			}

			users = found.filter((user) => user.userId !== auth.session?.userId);
			clearAlert();
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}
</script>

<div class="search">
	<Input placeholder="Поиск по username" bind:value={query} autofocus />
</div>

{#each users ?? [] as user (user.userId)}
	{@render row(user)}
{/each}

{#if users?.length === 0}
	<p class="note">Никого не нашлось</p>
{/if}

<style>
	.search {
		padding: 12px 16px;
	}

	.note {
		padding: 24px 16px;
		text-align: center;
		color: var(--text-muted);
	}
</style>
