<script lang="ts">
	import { push } from 'svelte-spa-router';
	import Page from '../components/Page.svelte';
	import Row from '../components/Row.svelte';
	import UserSearch from '../components/UserSearch.svelte';
	import { createChat } from '../lib/chats.js';
	import type { User } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';

	async function write(user: User) {
		try {
			const chat = await createChat('private', [user.userId]);
			clearAlert();
			push(`/chats/${chat.id}`);
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}
</script>

<Page title="Новый чат" back="/" padded={false}>
	<UserSearch>
		{#snippet row(user)}
			<Row
				onclick={() => write(user)}
				id={user.userId}
				title={user.visibleName ?? ''}
				subtitle={user.username ? `@${user.username}` : ''}
			/>
		{/snippet}
	</UserSearch>
</Page>
