<script lang="ts">
	import { link, push } from 'svelte-spa-router';
	import Page from '../components/Page.svelte';
	import Form from '../components/Form.svelte';
	import Input from '../components/Input.svelte';
	import Button from '../components/Button.svelte';
	import { login } from '../lib/auth.svelte.js';
	import { clearAlert, showAlert } from '../lib/alert.svelte.js';

	let credentials = $state({ login: '', password: '' });

	async function submit() {
		try {
			await login(credentials.login, credentials.password);
			clearAlert();
			push('/');
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}
</script>

<Page title="Вход" back="/">
	<Form onsubmit={submit}>
		<Input placeholder="Логин" bind:value={credentials.login} required />
		<Input placeholder="Пароль" type="password" bind:value={credentials.password} required />
		<Button type="submit">Войти</Button>
	</Form>
	<p class="hint">Нет аккаунта? <a href="/signup" use:link>Создать</a></p>
</Page>

<style>
	.hint {
		text-align: center;
		color: var(--text-muted);
	}

	.hint a {
		color: var(--text);
		font-weight: 600;
		text-decoration: none;
	}
</style>
