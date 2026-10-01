<script lang="ts">
	import { push } from 'svelte-spa-router';
	import Modal from './Modal.svelte';
	import Input from './Input.svelte';
	import Button from './Button.svelte';
	import Icon from './Icon.svelte';
	import Row from './Row.svelte';
	import UserSearch from './UserSearch.svelte';
	import { createChat } from '../lib/chats.js';
	import type { User } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';

	let { onclose }: { onclose: () => void } = $props();

	let searching = $state(false);
	let title = $state('');
	let members = $state<User[]>([]);

	function isMember(user: User): boolean {
		return members.some((member) => member.userId === user.userId);
	}

	function toggle(user: User) {
		members = isMember(user) ? members.filter((member) => member.userId !== user.userId) : [...members, user];
	}

	async function create(e: SubmitEvent) {
		e.preventDefault();

		try {
			const chat = await createChat(
				'group',
				members.map((member) => member.userId),
				title,
			);
			clearAlert();
			onclose();
			push(`/chats/${chat.id}`);
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}
</script>

{#snippet memberRow(user: User, label: string, variant: 'primary' | 'secondary')}
	<Row id={user.userId} title={user.visibleName ?? ''} subtitle={user.username ? `@${user.username}` : ''}>
		{#snippet trailing()}
			<Button small {variant} onclick={() => toggle(user)}>{label}</Button>
		{/snippet}
	</Row>
{/snippet}

<Modal
	title={searching ? 'Добавить участников' : 'Новая группа'}
	{onclose}
	onback={searching ? () => (searching = false) : undefined}
>
	{#if searching}
		<UserSearch>
			{#snippet row(user)}
				{#if isMember(user)}
					{@render memberRow(user, 'Добавлен', 'secondary')}
				{:else}
					{@render memberRow(user, 'Добавить', 'primary')}
				{/if}
			{/snippet}
		</UserSearch>
	{:else}
		<form id="new-group" class="form" onsubmit={create}>
			<Input placeholder="Название группы" bind:value={title} required maxlength={100} />
		</form>

		<p class="label">Участники · {members.length}</p>

		{#each members as member (member.userId)}
			{@render memberRow(member, 'Убрать', 'secondary')}
		{/each}

		<div class="add">
			<Button variant="secondary" onclick={() => (searching = true)}>
				<Icon name="plus" size={20} />Добавить участника
			</Button>
		</div>
	{/if}

	{#snippet footer()}
		{#if searching}
			<Button onclick={() => (searching = false)}>Готово · {members.length}</Button>
		{:else}
			<Button type="submit" form="new-group">Создать</Button>
		{/if}
	{/snippet}
</Modal>

<style>
	.form {
		padding: 4px 20px 8px;
	}

	.label {
		padding: 16px 20px 6px;
		font-size: 13px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--text-muted);
	}

	.add {
		padding: 12px 20px 4px;
	}
</style>
