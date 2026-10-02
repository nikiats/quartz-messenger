<script lang="ts">
	import { push } from 'svelte-spa-router';
	import Modal from './Modal.svelte';
	import Button from './Button.svelte';
	import ButtonRow from './ButtonRow.svelte';
	import Icon from './Icon.svelte';
	import UserRow from './UserRow.svelte';
	import ProfileCard from './ProfileCard.svelte';
	import MemberSearch from './MemberSearch.svelte';
	import { auth } from '../lib/auth.svelte.js';
	import { addMember, deleteChat, getMembers, removeMember, type Chat, type Member } from '../lib/chats.js';
	import { getUser, type User } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';

	let { chat, onclose }: { chat: Chat; onclose: () => void } = $props();

	type MemberProfile = User & Pick<Member, 'role'>;

	type ExitAction = {
		label: string;
		confirm: string;
		run: () => Promise<void>;
	};

	const leaveGroup: ExitAction = {
		label: 'Покинуть группу',
		confirm: 'Покинуть',
		run: () => removeMember(chat.id, auth.session!.userId),
	};

	const removeChat: ExitAction = {
		label: 'Удалить чат',
		confirm: 'Удалить',
		run: () => deleteChat(chat.id),
	};

	let members = $state<MemberProfile[]>([]);
	let searching = $state(false);
	let confirming = $state('');

	const isGroup = $derived(chat.type === 'group');
	const peer = $derived(members.find((member) => member.userId !== auth.session?.userId));
	const iAmAdmin = $derived(
		members.some((member) => member.userId === auth.session?.userId && member.role === 'admin'),
	);
	const exitActions = $derived.by(() => {
		if (!isGroup) {
			return [removeChat];
		}
		if (members.length === 0) {
			return [];
		}
		return iAmAdmin ? [leaveGroup, removeChat] : [leaveGroup];
	});

	async function load() {
		try {
			const list = await getMembers(chat.id);
			members = await Promise.all(list.map(async ({ userId, role }) => ({ ...(await getUser(userId)), role })));
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	function isMember(user: User): boolean {
		return members.some((member) => member.userId === user.userId);
	}

	async function add(user: User) {
		try {
			await addMember(chat.id, user.userId);
			await load();
			clearAlert();
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function kick(member: MemberProfile) {
		try {
			await removeMember(chat.id, member.userId);
			await load();
			clearAlert();
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function exit(action: ExitAction) {
		try {
			await action.run();
			clearAlert();
			onclose();
			push('/');
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	function subtitle(member: MemberProfile): string {
		if (member.userId === auth.session?.userId) {
			return 'вы';
		}
		if (member.role === 'admin') {
			return 'администратор';
		}
		return member.username ? `@${member.username}` : '';
	}

	load();
</script>

{#snippet done()}
	<Button onclick={() => (searching = false)}>Готово</Button>
{/snippet}

{#snippet exitButtons()}
	<div class="exit">
		{#each exitActions as action (action.label)}
			{#if confirming === action.label}
				<ButtonRow>
					<Button variant="secondary" onclick={() => (confirming = '')}>Отмена</Button>
					<Button variant="danger" onclick={() => exit(action)}>{action.confirm}</Button>
				</ButtonRow>
			{:else}
				<Button variant="danger" onclick={() => (confirming = action.label)}>{action.label}</Button>
			{/if}
		{/each}
	</div>
{/snippet}

<Modal
	title={searching ? 'Добавить участников' : isGroup ? 'Группа' : 'Профиль'}
	{onclose}
	onback={searching ? () => (searching = false) : undefined}
	footer={searching ? done : exitActions.length > 0 ? exitButtons : undefined}
>
	{#if searching}
		<MemberSearch added={isMember} onadd={add} />
	{:else if isGroup}
		<ProfileCard id={chat.id} name={chat.title} caption="Участников: {members.length}" />

		{#each members as member (member.userId)}
			<UserRow user={member} subtitle={subtitle(member)}>
				{#snippet trailing()}
					{#if iAmAdmin && member.userId !== auth.session?.userId}
						<Button small variant="secondary" onclick={() => kick(member)}>Исключить</Button>
					{/if}
				{/snippet}
			</UserRow>
		{/each}

		{#if iAmAdmin}
			<div class="add">
				<Button variant="secondary" onclick={() => (searching = true)}>
					<Icon name="plus" size={20} />Добавить участника
				</Button>
			</div>
		{/if}
	{:else}
		<ProfileCard id={chat.id} name={peer?.visibleName} caption={peer?.username ? `@${peer.username}` : ''} />
	{/if}
</Modal>

<style>
	.exit {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.add {
		padding: 12px 20px 20px;
	}
</style>
