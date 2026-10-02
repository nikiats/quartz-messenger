<script lang="ts">
	import { link, push } from 'svelte-spa-router';
	import Page from '../components/Page.svelte';
	import Form from '../components/Form.svelte';
	import Input from '../components/Input.svelte';
	import Button from '../components/Button.svelte';
	import { login, register } from '../lib/auth.svelte.js';
	import { showAlert } from '../lib/alert.svelte.js';
	import * as limits from '../config/account.js';

	let credentials = $state({ login: '', password: '' });

	async function submit() {
		try {
			await register(credentials.login, credentials.password);
			await login(credentials.login, credentials.password);
			showAlert('success', 'Аккаунт создан');
			push('/');
		} catch (err) {
			showAlert('error', (err as Error).message);
		}
	}
</script>

<Page title="Регистрация" back="/">
	<Form onsubmit={submit}>
		<Input
			placeholder="Логин"
			bind:value={credentials.login}
			required
			minlength={limits.login.min}
			maxlength={limits.login.max}
		/>
		<Input
			placeholder="Пароль"
			type="password"
			bind:value={credentials.password}
			required
			minlength={limits.password.min}
			maxlength={limits.password.max}
		/>
		<Button type="submit">Создать аккаунт</Button>
	</Form>
	<p class="hint">Уже есть аккаунт? <a href="/login" use:link>Войти</a></p>
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
