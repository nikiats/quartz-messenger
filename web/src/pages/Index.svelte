<script lang="ts">
	import { link } from 'svelte-spa-router';
	import Page from '../components/Page.svelte';
	import Button from '../components/Button.svelte';
	import ButtonRow from '../components/ButtonRow.svelte';
	import Avatar from '../components/Avatar.svelte';
	import Icon from '../components/Icon.svelte';
	import ChatList from '../components/ChatList.svelte';
	import NewGroupModal from '../components/NewGroupModal.svelte';
	import { auth } from '../lib/auth.svelte.js';

	let groupOpen = $state(false);
</script>

{#if auth.session}
	<Page title="Чаты" padded={false}>
		{#snippet actions()}
			<a href="/profile" class="me" use:link aria-label="Профиль">
				<Avatar name={auth.user?.visibleName} id={auth.session?.userId} size={38} />
			</a>
		{/snippet}

		<ChatList />

		{#snippet footer()}
			<ButtonRow>
				<Button href="/new-chat"><Icon name="pencil" size={20} />Написать</Button>
				<Button variant="secondary" onclick={() => (groupOpen = true)}>
					<Icon name="group" size={20} />Новая группа
				</Button>
			</ButtonRow>
		{/snippet}
	</Page>

	{#if groupOpen}
		<NewGroupModal onclose={() => (groupOpen = false)} />
	{/if}
{:else}
	<div class="welcome">
		<div class="brand">
			<span class="mark">q</span>
			<h1>quaztz</h1>
			<p>Безопасный быстрый мессенджер.</p>
		</div>
		<div class="actions">
			<Button href="/login">Войти</Button>
			<Button href="/signup" variant="secondary">Создать аккаунт</Button>
		</div>
	</div>
{/if}

<style>
	.me {
		display: flex;
		border-radius: 50%;
		text-decoration: none;
		transition: box-shadow 0.2s var(--ease);
	}

	.me:hover {
		box-shadow: 0 0 0 3px var(--active);
	}

	.welcome {
		display: flex;
		flex-direction: column;
		flex: 1;
		padding: 32px 20px 24px;
	}

	.brand {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 10px;
		flex: 1;
	}

	.mark {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 88px;
		height: 88px;
		margin-bottom: 12px;
		border-radius: 28px;
		font-size: 48px;
		font-weight: 700;
		color: var(--on-primary);
		background: linear-gradient(145deg, var(--primary), var(--text-muted));
	}

	h1 {
		font-size: 34px;
		font-weight: 700;
		letter-spacing: -0.03em;
	}

	p {
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
</style>
