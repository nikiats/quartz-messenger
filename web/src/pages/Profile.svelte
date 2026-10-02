<script lang="ts">
	import { push } from 'svelte-spa-router';
	import Page from '../components/Page.svelte';
	import Form from '../components/Form.svelte';
	import Input from '../components/Input.svelte';
	import Button from '../components/Button.svelte';
	import ButtonRow from '../components/ButtonRow.svelte';
	import ProfileCard from '../components/ProfileCard.svelte';
	import { auth, loadUser, logout } from '../lib/auth.svelte.js';
	import { updateProfile } from '../lib/users.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';
	import * as limits from '../config/account.js';

	let editing = $state(false);
	let visibleName = $state('');
	let username = $state('');

	function edit() {
		visibleName = auth.user?.visibleName ?? '';
		username = auth.user?.username ?? '';
		editing = true;
	}

	async function save() {
		try {
			await updateProfile(visibleName || undefined, username || undefined);
			await loadUser();
			editing = false;
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
	<ProfileCard
		id={auth.session?.userId}
		name={auth.user?.visibleName}
		caption={auth.user?.username ? `@${auth.user.username}` : ''}
	/>

	{#if editing}
		<Form onsubmit={save}>
			<Input
				placeholder="Имя"
				bind:value={visibleName}
				minlength={limits.visibleName.min}
				maxlength={limits.visibleName.max}
			/>
			<Input
				placeholder="Username"
				bind:value={username}
				minlength={limits.username.min}
				maxlength={limits.username.max}
			/>
			<ButtonRow>
				<Button variant="secondary" onclick={() => (editing = false)}>Отмена</Button>
				<Button type="submit">Сохранить</Button>
			</ButtonRow>
		</Form>
	{:else}
		<Button variant="secondary" onclick={edit}>Редактировать</Button>
	{/if}

	{#snippet footer()}
		<Button variant="danger" onclick={signOut}>Выйти</Button>
	{/snippet}
</Page>
