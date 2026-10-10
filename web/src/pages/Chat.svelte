<script lang="ts">
	import { untrack } from 'svelte';
	import { v4 as uuid } from 'uuid';
	import Page from '../components/Page.svelte';
	import Button from '../components/Button.svelte';
	import Avatar from '../components/Avatar.svelte';
	import Icon from '../components/Icon.svelte';
	import IconButton from '../components/IconButton.svelte';
	import MessageBubble from '../components/MessageBubble.svelte';
	import Composer from '../components/Composer.svelte';
	import MessageSearchModal from '../components/MessageSearchModal.svelte';
	import ChatInfoModal from '../components/ChatInfoModal.svelte';
	import { auth } from '../lib/auth.svelte.js';
	import { formatDate, isSameDay } from '../lib/format.js';
	import { getChat, type Chat } from '../lib/chats.js';
	import { compareIds, getMessages, newestId, sendMessage, type Message } from '../lib/messages.js';
	import { connection, subscribe } from '../lib/ws.svelte.js';
	import { setOpenedChat } from '../lib/unread.svelte.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';
	import { pollInterval } from '../config/connection.js';
	import { scrollButtonThreshold } from '../config/chat.js';

	let { params }: { params: { id: string } } = $props();

	type Bubble = {
		key: string;
		id: string | null;
		senderId: string;
		sentAt: string;
		text: string;
		failed?: boolean;
	};

	let chat = $state<Chat | null>(null);
	let messages = $state<Bubble[]>([]);
	let pending = $state<Bubble[]>([]);
	let hasMore = $state(false);
	let freshAfter = $state<string | null>(null);
	let text = $state('');
	let searchOpen = $state(false);
	let infoOpen = $state(false);
	let feed: HTMLDivElement;
	let scrolledUp = $state(false);
	let unseen = $state(0);
	let currentChatId: string | null = null;

	const newestFirst = $derived([...messages, ...pending].reverse());

	function toBubble(message: Message): Bubble {
		return {
			key: message.id,
			id: message.id,
			senderId: message.senderId,
			sentAt: message.sentAt,
			text: message.content.text,
		};
	}

	function addMessages(added: Bubble[]) {
		const known = new Set(messages.map((bubble) => bubble.id));
		const unique = added.filter((bubble) => !known.has(bubble.id));
		messages = [...messages, ...unique].sort((a, b) => compareIds(a.id!, b.id!));
	}

	function onScroll() {
		scrolledUp = feed.scrollTop < -scrollButtonThreshold;
		if (!scrolledUp) {
			unseen = 0;
		}
	}

	function scrollToBottom() {
		feed.scrollTo({ top: 0, behavior: 'smooth' });
	}

	function lastId(): string {
		return messages.at(-1)?.id ?? '0';
	}

	async function load(chatId: string): Promise<boolean> {
		try {
			const loaded = await getChat(chatId);
			const page = await getMessages(chatId, 'before', newestId);
			if (chatId !== currentChatId) {
				return false;
			}

			addMessages(page.messages.map(toBubble));
			hasMore = page.hasMore;
			freshAfter = lastId();
			chat = loaded;
			return true;
		} catch (err) {
			showAlert('error', (err as Error).message);
			return false;
		}
	}

	async function loadOlder() {
		try {
			const page = await getMessages(params.id, 'before', messages[0].id!);
			addMessages(page.messages.map(toBubble));
			hasMore = page.hasMore;
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function loadNewer(chatId: string, after: string) {
		try {
			let page;
			do {
				page = await getMessages(chatId, 'after', after);
				if (chatId !== currentChatId) {
					return;
				}

				addMessages(page.messages.map(toBubble));
				after = page.messages.at(-1)?.id ?? after;
			} while (page.hasMore && page.messages.length > 0);
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	$effect(() => {
		const chatId = params.id;
		let ready = false;

		currentChatId = chatId;
		chat = null;
		messages = [];
		pending = [];
		hasMore = false;
		freshAfter = null;
		untrack(() => {
			setOpenedChat(chatId);

			load(chatId).then((loaded) => {
				if (!loaded) {
					return;
				}

				ready = true;
				loadNewer(chatId, lastId());
			});
		});

		const unsubscribe = subscribe((event) => {
			if (!ready) {
				return;
			}

			if (event.method === 'receiveMessage' && event.chatId === chatId) {
				addMessages([toBubble({ ...event, id: event.messageId })]);
				if (scrolledUp) {
					unseen += 1;
				}
			}

			if (event.method === 'authorizeResponse' && event.success) {
				loadNewer(chatId, lastId());
			}
		});

		return () => {
			unsubscribe();
			setOpenedChat(null);
			currentChatId = null;
		};
	});

	$effect(() => {
		if (connection.status === 'online' || !chat) {
			return;
		}

		const chatId = chat.id;
		const timer = setInterval(async () => {
			const page = await getMessages(chatId, 'after', lastId()).catch(() => null);
			if (page) {
				addMessages(page.messages.map(toBubble));
			}
		}, pollInterval);
		return () => clearInterval(timer);
	});

	function send() {
		const bubble: Bubble = {
			key: uuid(),
			id: null,
			senderId: auth.session!.userId,
			sentAt: new Date().toISOString(),
			text,
		};
		text = '';
		deliver(bubble);
		scrollToBottom();
	}

	function retry(bubble: Bubble) {
		deliver({ ...bubble, failed: false });
	}

	function statusOf(bubble: Bubble): 'failed' | 'pending' | 'sent' {
		if (bubble.failed) {
			return 'failed';
		}
		return bubble.id === null ? 'pending' : 'sent';
	}

	function discard(bubble: Bubble) {
		pending = pending.filter((item) => item.key !== bubble.key);
	}

	async function deliver(bubble: Bubble) {
		pending = [...pending.filter((item) => item.key !== bubble.key), bubble];

		try {
			const sent = await sendMessage(params.id, bubble.key, bubble.text);
			discard(bubble);
			addMessages([{ ...bubble, id: sent.id, sentAt: sent.sentAt }]);
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

	<div class="feed">
		<div class="messages" bind:this={feed} onscroll={onScroll}>
			{#each newestFirst as bubble, i (bubble.key)}
				{@const previous = newestFirst[i + 1]}
				{@const newDay = !previous || !isSameDay(previous.sentAt, bubble.sentAt)}
				{@const mine = bubble.senderId === auth.session?.userId}
				{@const first = newDay || previous.senderId !== bubble.senderId}
				<MessageBubble
					text={bubble.text}
					sentAt={bubble.sentAt}
					status={mine ? statusOf(bubble) : undefined}
					{mine}
					{first}
					fresh={bubble.id === null || (freshAfter !== null && compareIds(bubble.id, freshAfter) > 0)}
					senderId={chat?.type === 'group' && !mine && first ? bubble.senderId : undefined}
					onretry={() => retry(bubble)}
					ondiscard={() => discard(bubble)}
				/>
				{#if newDay}
					<p class="date">{formatDate(bubble.sentAt)}</p>
				{/if}
			{/each}

			{#if hasMore}
				<div class="more">
					<Button variant="secondary" onclick={loadOlder}>Загрузить ещё</Button>
				</div>
			{/if}
		</div>

		{#if scrolledUp}
			<button type="button" class="down" aria-label="К новым сообщениям" onclick={scrollToBottom}>
				<Icon name="down" size={22} />
				{#if unseen > 0}
					<span class="badge">{unseen}</span>
				{/if}
			</button>
		{/if}
	</div>

	{#snippet footer()}
		<Composer bind:value={text} onsend={send} />
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

	.feed {
		position: relative;
		display: flex;
		flex: 1;
		min-height: 0;
	}

	.date {
		align-self: center;
		margin: 14px 0 6px;
		padding: 4px 12px;
		border-radius: 999px;
		font-size: 12px;
		font-weight: 500;
		color: var(--text-muted);
		background-color: var(--surface);
	}

	.down {
		position: absolute;
		right: 14px;
		bottom: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		border: 1px solid var(--border);
		border-radius: 50%;
		color: var(--text);
		background-color: var(--surface);
		box-shadow: 0 6px 18px rgba(0, 0, 0, 0.35);
		cursor: pointer;
	}

	.badge {
		position: absolute;
		top: -6px;
		right: -4px;
		min-width: 20px;
		padding: 2px 6px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 700;
		color: var(--on-primary);
		background-color: var(--primary);
	}
</style>
