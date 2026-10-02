<script lang="ts">
	import { v4 as uuid } from 'uuid';
	import Page from '../components/Page.svelte';
	import Input from '../components/Input.svelte';
	import Button from '../components/Button.svelte';
	import Avatar from '../components/Avatar.svelte';
	import Icon from '../components/Icon.svelte';
	import IconButton from '../components/IconButton.svelte';
	import MessageBubble from '../components/MessageBubble.svelte';
	import MessageSearchModal from '../components/MessageSearchModal.svelte';
	import ChatInfoModal from '../components/ChatInfoModal.svelte';
	import { auth } from '../lib/auth.svelte.js';
	import { getChat, type Chat } from '../lib/chats.js';
	import { getMessages, sendMessage, type Message } from '../lib/messages.js';
	import { connection, subscribe } from '../lib/ws.svelte.js';
	import { setOpenedChat } from '../lib/unread.svelte.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';
	import { pollInterval } from '../config/connection.js';
	import { messageMaxLength } from '../config/chat.js';

	let { params }: { params: { id: string } } = $props();

	type Bubble = {
		key: string;
		seqnum: number | null;
		senderId: string;
		sentAt: string;
		text: string;
		failed?: boolean;
	};

	let chat = $state<Chat | null>(null);
	let messages = $state<Bubble[]>([]);
	let pending = $state<Bubble[]>([]);
	let hasMore = $state(false);
	let freshFrom = $state(Infinity);
	let text = $state('');
	let searchOpen = $state(false);
	let infoOpen = $state(false);

	const newestFirst = $derived([...messages, ...pending].reverse());

	function toBubble(message: Message): Bubble {
		return {
			key: String(message.seqnum),
			seqnum: message.seqnum,
			senderId: message.senderId,
			sentAt: message.sentAt,
			text: message.content.text,
		};
	}

	function addMessages(added: Bubble[]) {
		const known = new Set(messages.map((bubble) => bubble.seqnum));
		const unique = added.filter((bubble) => !known.has(bubble.seqnum));
		messages = [...messages, ...unique].sort((a, b) => a.seqnum! - b.seqnum!);
	}

	function lastSeqnum(): number {
		return messages.at(-1)?.seqnum ?? 0;
	}

	async function load(chatId: string) {
		try {
			chat = await getChat(chatId);
			const page = await getMessages(chatId, 'before', chat.lastSeqnum + 1);
			addMessages(page.messages.map(toBubble));
			hasMore = page.hasMore;
			freshFrom = lastSeqnum() + 1;
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function loadOlder() {
		try {
			const page = await getMessages(params.id, 'before', messages[0].seqnum!);
			addMessages(page.messages.map(toBubble));
			hasMore = page.hasMore;
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function loadNewer(chatId: string, after: number) {
		try {
			const page = await getMessages(chatId, 'after', after);
			addMessages(page.messages.map(toBubble));
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	$effect(() => {
		const chatId = params.id;

		chat = null;
		messages = [];
		pending = [];
		hasMore = false;
		freshFrom = Infinity;
		load(chatId);
		setOpenedChat(chatId);

		const unsubscribe = subscribe((event) => {
			if (!chat) {
				return;
			}

			if (event.method === 'receiveMessage' && event.chatId === chatId) {
				const last = lastSeqnum();
				addMessages([toBubble(event)]);
				if (event.seqnum > last + 1) {
					loadNewer(chatId, last);
				}
			}

			if (event.method === 'authorizeResponse' && event.success) {
				loadNewer(chatId, lastSeqnum());
			}
		});

		return () => {
			unsubscribe();
			setOpenedChat(null);
		};
	});

	$effect(() => {
		if (connection.status === 'online' || !chat) {
			return;
		}

		const chatId = chat.id;
		const timer = setInterval(async () => {
			const page = await getMessages(chatId, 'after', lastSeqnum()).catch(() => null);
			if (page) {
				addMessages(page.messages.map(toBubble));
			}
		}, pollInterval);
		return () => clearInterval(timer);
	});

	async function send(e: SubmitEvent) {
		e.preventDefault();

		const bubble: Bubble = {
			key: uuid(),
			seqnum: null,
			senderId: auth.session!.userId,
			sentAt: new Date().toISOString(),
			text,
		};
		text = '';
		deliver(bubble);
	}

	function retry(bubble: Bubble) {
		deliver({ ...bubble, failed: false });
	}

	function statusOf(bubble: Bubble): 'failed' | 'pending' | 'sent' {
		if (bubble.failed) {
			return 'failed';
		}
		return bubble.seqnum === null ? 'pending' : 'sent';
	}

	function discard(bubble: Bubble) {
		pending = pending.filter((item) => item.key !== bubble.key);
	}

	async function deliver(bubble: Bubble) {
		pending = [...pending.filter((item) => item.key !== bubble.key), bubble];

		try {
			const sent = await sendMessage(params.id, bubble.key, bubble.text);
			discard(bubble);
			addMessages([{ ...bubble, seqnum: sent.seqnum, sentAt: sent.sentAt }]);
			clearAlert();
		} catch (err) {
			pending = pending.map((item) => (item.key === bubble.key ? { ...item, failed: true } : item));
			showAlert('error', (err as Error).message);
		}
	}
</script>

<Page back="/" padded={false}>
	{#snippet heading()}
		<button type="button" class="peer" onclick={() => (infoOpen = true)} disabled={!chat}>
			<Avatar name={chat?.title} id={params.id} size={40} />
			<span class="peer-text">
				<span class="peer-title">{chat?.title ?? ''}</span>
				<span class="peer-type">{chat?.type === 'group' ? 'группа' : 'личный чат'}</span>
			</span>
		</button>
	{/snippet}

	{#snippet actions()}
		<IconButton icon="search" label="Поиск" onclick={() => (searchOpen = true)} />
	{/snippet}

	<div class="messages">
		{#each newestFirst as bubble, i (bubble.key)}
			{@const mine = bubble.senderId === auth.session?.userId}
			{@const first = newestFirst[i + 1]?.senderId !== bubble.senderId}
			<MessageBubble
				text={bubble.text}
				sentAt={bubble.sentAt}
				status={mine ? statusOf(bubble) : undefined}
				{mine}
				{first}
				fresh={bubble.seqnum === null || bubble.seqnum >= freshFrom}
				senderId={chat?.type === 'group' && !mine && first ? bubble.senderId : undefined}
				onretry={() => retry(bubble)}
				ondiscard={() => discard(bubble)}
			/>
		{/each}

		{#if hasMore}
			<div class="more">
				<Button variant="secondary" onclick={loadOlder}>Загрузить ещё</Button>
			</div>
		{/if}
	</div>

	{#snippet footer()}
		<form class="composer" onsubmit={send}>
			<Input placeholder="Сообщение" bind:value={text} required maxlength={messageMaxLength} />
			<button type="submit" class="send" aria-label="Отправить"><Icon name="send" size={22} /></button>
		</form>
	{/snippet}
</Page>

{#if searchOpen}
	<MessageSearchModal chatId={params.id} onclose={() => (searchOpen = false)} />
{/if}

{#if infoOpen && chat}
	<ChatInfoModal {chat} onclose={() => (infoOpen = false)} />
{/if}

<style>
	.peer {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 0;
		padding: 4px 12px 4px 4px;
		border: none;
		border-radius: 999px;
		text-align: left;
		color: var(--text);
		background: none;
		cursor: pointer;
		transition: background-color 0.2s var(--ease);
	}

	.peer:hover {
		background-color: var(--hover);
	}

	.peer-text {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.peer-title {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		font-size: 16px;
		font-weight: 600;
	}

	.peer-type {
		font-size: 13px;
		color: var(--text-muted);
	}

	.messages {
		display: flex;
		flex-direction: column-reverse;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		gap: 3px;
		padding: 16px 14px;
	}

	.more {
		align-self: center;
		margin-bottom: 12px;
	}

	.composer {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.send {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		width: 52px;
		height: 52px;
		border: none;
		border-radius: 50%;
		color: var(--on-primary);
		background-color: var(--primary);
		cursor: pointer;
		transition:
			transform 0.15s var(--ease),
			filter 0.2s var(--ease);
	}

	.send:hover {
		filter: brightness(1.06);
	}

	.send:active {
		transform: scale(0.94);
	}
</style>
