<script lang="ts">
	import Button from './Button.svelte';
	import UserRow from './UserRow.svelte';
	import UserSearch from './UserSearch.svelte';
	import type { User } from '../lib/users.js';

	let {
		added,
		onadd,
		onremove,
	}: {
		added: (user: User) => boolean;
		onadd: (user: User) => void;
		onremove?: (user: User) => void;
	} = $props();
</script>

<UserSearch>
	{#snippet row(user)}
		<UserRow {user}>
			{#snippet trailing()}
				{#if added(user)}
					<Button small variant="secondary" onclick={() => onremove?.(user)}>Добавлен</Button>
				{:else}
					<Button small onclick={() => onadd(user)}>Добавить</Button>
				{/if}
			{/snippet}
		</UserRow>
	{/snippet}
</UserSearch>
