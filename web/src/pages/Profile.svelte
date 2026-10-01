<script lang="ts">
	import { push } from 'svelte-spa-router';
	import Page from '../components/Page.svelte';
	import Form from '../components/Form.svelte';
	import Input from '../components/Input.svelte';
	import Button from '../components/Button.svelte';
	import Avatar from '../components/Avatar.svelte';
	import { auth, loadUser, logout } from '../lib/auth.svelte.js';
	import { updateProfile } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';

	let visibleName = $state(auth.user?.visibleName ?? '');
	let username = $state(auth.user?.username ?? '');

	async function save() {
		try {
			await updateProfile(visibleName || undefined, username || undefined);
			await loadUser();
			showAlert('success', 'Профиль сохранён');
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}

	async function signOut() {
		await logout();
		clearAlert();
		push('/');
	}
</script>

<Page title="Профиль" back="/">
	<div class="card">
		<Avatar name={auth.user?.visibleName} id={auth.session?.userId} size={88} />
		<p class="name">{auth.user?.visibleName ?? ''}</p>
		{#if auth.user?.username}
			<p class="username">@{auth.user.username}</p>
		{/if}
	</div>

	<Form onsubmit={save}>
		<Input placeholder="Имя" bind:value={visibleName} minlength={1} maxlength={32} />
		<Input placeholder="Username" bind:value={username} minlength={3} maxlength={24} />
		<Button type="submit">Сохранить</Button>
	</Form>

	{#snippet footer()}
		<Button variant="danger" onclick={signOut}>Выйти</Button>
	{/snippet}
</Page>

<style>
	.card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 12px 0 20px;
	}

	.name {
		margin-top: 10px;
		font-size: 20px;
		font-weight: 600;
	}

	.username {
		color: var(--text-muted);
	}
</style>
