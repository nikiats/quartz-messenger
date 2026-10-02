<script lang="ts">
	import Icon from './Icon.svelte';
	import { messageMaxLength } from '../config/chat.js';

	let { value = $bindable(''), onsend }: { value?: string; onsend: () => void } = $props();

	const touchScreen = matchMedia('(pointer: coarse)').matches;

	let textarea: HTMLTextAreaElement;

	$effect(() => {
		if (!touchScreen) {
			textarea.focus();
		}
	});

	function send() {
		if (!value.trim()) {
			return;
		}

		onsend();
		textarea.focus();
	}

	function submit(e: SubmitEvent) {
		e.preventDefault();
		send();
	}

	function keydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey && !e.isComposing && !touchScreen) {
			e.preventDefault();
			send();
		}
	}
</script>

<form class="composer" onsubmit={submit}>
	<textarea
		bind:this={textarea}
		bind:value
		rows="1"
		placeholder="Сообщение"
		maxlength={messageMaxLength}
		onkeydown={keydown}
	></textarea>
	<button type="submit" class="send" aria-label="Отправить" onmousedown={(e) => e.preventDefault()}>
		<Icon name="send" size={22} />
	</button>
</form>

<style>
	.composer {
		display: flex;
		align-items: flex-end;
		gap: 10px;
	}

	textarea {
		flex: 1;
		min-height: 52px;
		max-height: 160px;
		field-sizing: content;
		padding: 14px 20px;
		resize: none;
		font: inherit;
		font-size: 16px;
		line-height: 1.4;
		color: var(--text);
		background-color: var(--surface);
		border: 1px solid transparent;
		border-radius: 26px;
		outline: none;
		transition:
			border-color 0.2s var(--ease),
			box-shadow 0.2s var(--ease);
	}

	textarea::placeholder {
		color: var(--text-muted);
	}

	textarea:focus {
		border-color: var(--text-muted);
		box-shadow: 0 0 0 4px var(--hover);
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
