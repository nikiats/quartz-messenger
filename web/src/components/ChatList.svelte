<script lang="ts">
	import Row from './Row.svelte';
	import { getChats, type Chat } from '../lib/chats.js';
	import { connection, subscribe } from '../lib/ws.svelte.js';
	import { unread } from '../lib/unread.svelte.js';
	import { formatDay } from '../lib/format.js';
	import { pollInterval } from '../config/connection.js';

	let chats = $state<Chat[] | null>(null);
	let error = $state('');

	async function load() {
		try {
			const fresh = await getChats();
			markUnread(fresh);
			chats = fresh;
		} catch (err) {
			error = (err as Error).message;
		}
	}

	function markUnread(fresh: Chat[]) {
		if (!chats) {
			return;
		}

		for (const chat of fresh) {
			const before = chats.find((known) => known.id === chat.id);
			if (chat.lastMessage && chat.lastMessage.id !== before?.lastMessage?.id) {
				unread.add(chat.id);
			}
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

			chat.lastMessage = {
				id: event.messageId,
				senderId: event.senderId,
				preview: event.content.text,
				sentAt: event.sentAt,
			};
			chats = [chat, ...chats.filter((other) => other !== chat)];
		});
	});

	$effect(() => {
		if (connection.status === 'online') {
			return;
		}

		const timer = setInterval(load, pollInterval);
		return () => clearInterval(timer);
	});
</script>

{#if chats}
	{#each chats as chat (chat.id)}
		<Row
			href="/chats/{chat.id}"
			id={chat.id}
			title={chat.title ?? ''}
			subtitle={chat.type === 'group' ? 'Группа' : 'Личный чат'}
			meta={chat.lastMessage ? formatDay(chat.lastMessage.sentAt) : ''}
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
