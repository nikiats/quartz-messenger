<script lang="ts">
	import { push } from 'svelte-spa-router';
	import { v4 as uuid } from 'uuid';
	import Modal from './Modal.svelte';
	import Input from './Input.svelte';
	import Button from './Button.svelte';
	import Icon from './Icon.svelte';
	import UserRow from './UserRow.svelte';
	import MemberSearch from './MemberSearch.svelte';
	import { createChat } from '../lib/chats.js';
	import type { User } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';
	import { groupTitleMaxLength } from '../config/chat.js';

	let { onclose }: { onclose: () => void } = $props();

	let searching = $state(false);
	let title = $state('');
	let members = $state<User[]>([]);
	let creating = $state(false);

	const requestId = uuid();

	function isMember(user: User): boolean {
		return members.some((member) => member.userId === user.userId);
	}

	function add(user: User) {
		members = [...members, user];
	}

	function remove(user: User) {
		members = members.filter((member) => member.userId !== user.userId);
	}

	async function create(e: SubmitEvent) {
		e.preventDefault();

		creating = true;
		try {
			const chat = await createChat(
				'group',
				members.map((member) => member.userId),
				title,
				requestId,
			);
			clearAlert();
			onclose();
			push(`/chats/${chat.id}`);
		} catch (err) {
			showAlert('error', (err as Error).message);
		} finally {
			creating = false;
		}
	}
</script>

<Modal
	title={searching ? 'Добавить участников' : 'Новая группа'}
	{onclose}
	onback={searching ? () => (searching = false) : undefined}
>
	{#if searching}
		<MemberSearch added={isMember} onadd={add} onremove={remove} />
	{:else}
		<form id="new-group" class="form" onsubmit={create}>
			<Input placeholder="Название группы" bind:value={title} required maxlength={groupTitleMaxLength} autofocus />
		</form>

		<p class="label">Участники · {members.length}</p>

		{#each members as member (member.userId)}
			<UserRow user={member}>
				{#snippet trailing()}
					<Button small variant="secondary" onclick={() => remove(member)}>Убрать</Button>
				{/snippet}
			</UserRow>
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
			<Button type="submit" form="new-group" disabled={creating}>Создать</Button>
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
