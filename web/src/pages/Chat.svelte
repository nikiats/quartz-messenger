<script lang="ts">
	import Page from '../components/Page.svelte';
	import Input from '../components/Input.svelte';
	import Button from '../components/Button.svelte';
	import Avatar from '../components/Avatar.svelte';
	import Icon from '../components/Icon.svelte';
	import { auth } from '../lib/auth.svelte.js';
	import { formatTime } from '../lib/format.js';
	import { getChat, type Chat } from '../lib/chats.js';
	import { getMessages, type Message } from '../lib/messages.js';
	import { getUser } from '../lib/users.js';
	import { sendMessage, subscribe } from '../lib/ws.js';
	import { setOpenedChat } from '../lib/unread.svelte.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';

	let { params }: { params: { id: string } } = $props();

	let chat = $state<Chat | null>(null);
	let messages = $state<Message[]>([]);
	let hasMore = $state(false);
	let freshFrom = $state(Infinity);
	let text = $state('');

	const newestFirst = $derived([...messages].reverse());

	function merge(incoming: Message[]) {
		const bySeqnum = new Map(messages.map((message) => [message.seqnum, message]));
		for (const message of incoming) {
			bySeqnum.set(message.seqnum, message);
		}
		messages = [...bySeqnum.values()].sort((a, b) => a.seqnum - b.seqnum);
	}

	function lastSeqnum(): number {
		return messages.at(-1)?.seqnum ?? 0;
	}

	async function load(chatId: string) {
		try {
			chat = await getChat(chatId);
			const page = await getMessages(chatId, 'before', chat.lastSeqnum + 1);
			merge(page.messages);
			hasMore = page.hasMore;
			freshFrom = lastSeqnum() + 1;
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function loadOlder() {
		try {
			const page = await getMessages(params.id, 'before', messages[0].seqnum);
			merge(page.messages);
			hasMore = page.hasMore;
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function loadNewer(chatId: string, after: number) {
		try {
			const page = await getMessages(chatId, 'after', after);
			merge(page.messages);
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	$effect(() => {
		const chatId = params.id;

		chat = null;
		messages = [];
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
				merge([event]);
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

	async function send(e: SubmitEvent) {
		e.preventDefault();

		const content = text;
		text = '';

		try {
			const sent = await sendMessage(params.id, content);
			merge([
				{
					chatId: params.id,
					senderId: auth.session!.userId,
					seqnum: sent.seqnum,
					sentAt: sent.sentAt,
					content: { text: content },
				},
			]);
			clearAlert();
		} catch (err) {
			text = content;
			showAlert('error', (err as Error).message);
		}
	}
</script>

<Page back="/" padded={false}>
	{#snippet heading()}
		<div class="peer">
			<Avatar name={chat?.title} id={params.id} size={40} />
			<div class="peer-text">
				<span class="peer-title">{chat?.title ?? ''}</span>
				<span class="peer-type">{chat?.type === 'group' ? 'группа' : 'личный чат'}</span>
			</div>
		</div>
	{/snippet}

	<div class="messages">
		{#each newestFirst as message, i (message.seqnum)}
			{@const mine = message.senderId === auth.session?.userId}
			{@const first = newestFirst[i + 1]?.senderId !== message.senderId}
			<div class="message" class:mine class:first class:fresh={message.seqnum >= freshFrom}>
				{#if chat?.type === 'group' && !mine && first}
					<span class="sender">
						{#await getUser(message.senderId) then user}{user.visibleName}{/await}
					</span>
				{/if}
				<span class="text">{message.content.text}</span>
				<span class="time">{formatTime(message.sentAt)}</span>
			</div>
		{/each}

		{#if hasMore}
			<div class="more">
				<Button variant="secondary" onclick={loadOlder}>Загрузить ещё</Button>
			</div>
		{/if}
	</div>

	{#snippet footer()}
		<form class="composer" onsubmit={send}>
			<Input placeholder="Сообщение" bind:value={text} required maxlength={1000} />
			<button type="submit" class="send" aria-label="Отправить"><Icon name="send" size={22} /></button>
		</form>
	{/snippet}
</Page>

<style>
	.peer {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 0;
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

	.message {
		display: flex;
		flex-direction: column;
		align-self: flex-start;
		max-width: 76%;
		padding: 8px 14px;
		border-radius: 20px;
		background-color: var(--surface);
		overflow-wrap: anywhere;
	}

	.message.fresh {
		animation: appear 0.4s cubic-bezier(0.25, 0.8, 0.35, 1);
		transform-origin: bottom left;
	}

	.message.fresh.mine {
		transform-origin: bottom right;
	}

	@keyframes appear {
		from {
			opacity: 0;
			transform: translateY(18px) scale(0.85);
		}
	}

	.message.first {
		margin-top: 10px;
	}

	.message.mine {
		align-self: flex-end;
		color: var(--text);
		background-color: var(--accent);
	}

	.message.first:not(.mine) {
		border-top-left-radius: 6px;
	}

	.message.first.mine {
		border-top-right-radius: 6px;
	}

	.sender {
		margin-bottom: 2px;
		font-size: 13px;
		font-weight: 600;
		color: var(--text-muted);
	}

	.text {
		font-size: 15px;
		line-height: 1.4;
		white-space: pre-wrap;
	}

	.time {
		align-self: flex-end;
		margin-top: 2px;
		font-size: 11px;
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.message.mine .time {
		color: var(--text);
		opacity: 0.65;
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
