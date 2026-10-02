<script lang="ts">
	import { getUser } from '../lib/users.js';
	import { formatTime } from '../lib/format.js';

	let {
		text,
		sentAt,
		status,
		mine,
		first,
		fresh,
		senderId,
		onretry,
		ondiscard,
	}: {
		text: string;
		sentAt: string;
		status?: 'failed' | 'pending' | 'sent' | 'delivered' | 'read';
		mine: boolean;
		first: boolean;
		fresh: boolean;
		senderId?: string;
		onretry?: () => void;
		ondiscard?: () => void;
	} = $props();
</script>

<div class="message" class:mine class:first class:fresh>
	{#if senderId}
		<span class="sender">
			{#await getUser(senderId) then user}{user.visibleName}{/await}
		</span>
	{/if}
	<span class="text">{text}</span>
	<span class="meta">
		{formatTime(sentAt)}
		{#if status}
			<span class="dot {status}"></span>
		{/if}
	</span>
	{#if status === 'failed'}
		<span class="actions">
			<button type="button" onclick={onretry}>Повторить</button>
			<button type="button" onclick={ondiscard}>Удалить</button>
		</span>
	{/if}
</div>

<style>
	.message {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: 2px 10px;
		align-self: flex-start;
		max-width: 76%;
		padding: 8px 12px 8px 14px;
		border-radius: 20px;
		background-color: var(--surface);
	}

	.message.mine {
		align-self: flex-end;
		background-color: var(--accent);
	}

	.message.first {
		margin-top: 10px;
	}

	.message.first:not(.mine) {
		border-top-left-radius: 6px;
	}

	.message.first.mine {
		border-top-right-radius: 6px;
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

	.sender {
		flex-basis: 100%;
		font-size: 13px;
		font-weight: 600;
		color: var(--text-muted);
	}

	.text {
		flex: 1 1 auto;
		min-width: 0;
		font-size: 15px;
		line-height: 1.4;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}

	.meta {
		display: flex;
		align-items: center;
		gap: 5px;
		margin-left: auto;
		white-space: nowrap;
		font-size: 11px;
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.message.mine .meta {
		color: rgba(224, 225, 221, 0.65);
	}

	.dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		transition:
			background-color 0.3s var(--ease),
			box-shadow 0.3s var(--ease);
	}

	.failed {
		background-color: #FF5A52;
		box-shadow: 0 0 6px rgba(255, 90, 82, 0.9);
	}

	.pending {
		background-color: rgba(224, 225, 221, 0.25);
	}

	.sent {
		background-color: #F5A04C;
		box-shadow: 0 0 6px rgba(245, 160, 76, 0.8);
	}

	.delivered {
		background-color: #E0E1DD;
		box-shadow: 0 0 6px rgba(224, 225, 221, 0.8);
	}

	.read {
		background-color: #5AA2FF;
		box-shadow: 0 0 6px rgba(90, 162, 255, 0.8);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 6px;
		flex-basis: 100%;
		margin-top: 4px;
	}

	.actions button {
		padding: 4px 10px;
		border: none;
		border-radius: 999px;
		font-size: 12px;
		font-weight: 600;
		color: var(--text);
		background-color: rgba(13, 27, 42, 0.45);
		cursor: pointer;
	}

	.actions button:hover {
		background-color: rgba(13, 27, 42, 0.7);
	}
</style>
