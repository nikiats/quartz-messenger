<script lang="ts">
	import Row from './Row.svelte';
	import { getChats, type Chat } from '../lib/chats.js';
	import { subscribe } from '../lib/ws.js';
	import { unread } from '../lib/unread.svelte.js';
	import { formatDay } from '../lib/format.js';

	let chats = $state<Chat[] | null>(null);
	let error = $state('');

	async function load() {
		try {
			chats = await getChats();
		} catch (err) {
			error = (err as Error).message;
		}
	}

	$effect(() => {
		load();

		return subscribe((event) => {
			if (event.method !== 'receiveMessage' || !chats) {
				return;
			}

			const chat = chats.find((chat) => chat.id === event.chatId);
			if (!chat) {
				load();
				return;
			}

			chat.lastSeqnum = event.seqnum;
			chat.lastMessageAt = event.sentAt;
			chats = [chat, ...chats.filter((other) => other !== chat)];
		});
	});
</script>

{#if chats}
	{#each chats as chat (chat.id)}
		<Row
			href="/chats/{chat.id}"
			id={chat.id}
			title={chat.title ?? ''}
			subtitle={chat.type === 'group' ? 'Группа' : 'Личный чат'}
			meta={formatDay(chat.lastMessageAt)}
			unread={unread.has(chat.id)}
		/>
	{:else}
		<p class="note">Чатов пока нет</p>
	{/each}
{:else if error}
	<p class="note">{error}</p>
{:else}
	<p class="note">Загрузка...</p>
{/if}

<style>
	.note {
		margin: auto;
		color: var(--text-muted);
	}
</style>
