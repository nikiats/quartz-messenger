<script lang="ts">
	import Modal from './Modal.svelte';
	import Input from './Input.svelte';
	import Button from './Button.svelte';
	import { auth } from '../lib/auth.svelte.js';
	import { searchMessages, type Message } from '../lib/messages.js';
	import { getUser } from '../lib/users.js';
	import { formatDay } from '../lib/format.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';
	import { messageQueryMaxLength, messageSearchDelay } from '../config/search.js';

	let { chatId, onclose }: { chatId: string; onclose: () => void } = $props();

	const newest = Number.MAX_SAFE_INTEGER;

	let query = $state('');
	let found = $state<Message[] | null>(null);
	let hasMore = $state(false);

	$effect(() => {
		const q = query.trim();
		if (!q) {
			found = null;
			return;
		}

		const timer = setTimeout(() => search(q), messageSearchDelay);
		return () => clearTimeout(timer);
	});

	async function search(q: string) {
		try {
			const page = await searchMessages(chatId, q, newest);
			if (q !== query.trim()) {
				return;
			}

			found = page.messages;
			hasMore = page.hasMore;
			clearAlert();
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function loadMore() {
		if (!found) {
			return;
		}

		try {
			const page = await searchMessages(chatId, query.trim(), found[found.length - 1].seqnum);
			found = [...found, ...page.messages];
			hasMore = page.hasMore;
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}
</script>

<Modal title="Поиск в чате" {onclose}>
	<div class="field">
		<Input placeholder="Текст сообщения" bind:value={query} maxlength={messageQueryMaxLength} autofocus />
	</div>

	{#each found ?? [] as message (message.seqnum)}
		<div class="result">
			<div class="top">
				<span class="sender">
					{#if message.senderId === auth.session?.userId}
						Вы
					{:else}
						{#await getUser(message.senderId) then user}{user.visibleName}{/await}
					{/if}
				</span>
				<span class="date">{formatDay(message.sentAt)}</span>
			</div>
			<p class="text">{message.content.text}</p>
		</div>
	{/each}

	{#if found?.length === 0}
		<p class="note">Ничего не нашлось</p>
	{/if}

	{#if hasMore}
		<div class="more">
			<Button variant="secondary" onclick={loadMore}>Загрузить ещё</Button>
		</div>
	{/if}
</Modal>

<style>
	.field {
		padding: 4px 20px 12px;
	}

	.result {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 12px 20px;
		border-bottom: 1px solid var(--divider);
	}

	.top {
		display: flex;
		gap: 8px;
	}

	.sender {
		flex: 1;
		font-size: 14px;
		font-weight: 600;
	}

	.date {
		font-size: 13px;
		color: var(--text-muted);
	}

	.text {
		font-size: 15px;
		line-height: 1.4;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}

	.note {
		padding: 24px 20px;
		text-align: center;
		color: var(--text-muted);
	}

	.more {
		padding: 8px 20px 20px;
	}
</style>
