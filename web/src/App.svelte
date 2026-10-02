<script lang="ts">
	import Router from 'svelte-spa-router';

	import Alert from './components/Alert.svelte';
	import ConnectionBanner from './components/ConnectionBanner.svelte';

	import Index from './pages/Index.svelte';
	import Chat from './pages/Chat.svelte';
	import Signup from './pages/Signup.svelte';
	import Login from './pages/Login.svelte';
	import Profile from './pages/Profile.svelte';
	import NewChat from './pages/NewChat.svelte';

	import { auth, loadUser } from './lib/auth.svelte.js';
	import { connect, disconnect } from './lib/ws.svelte.js';

	const routes = {
		'/': Index,
		'/chats/:id': Chat,
		'/signup': Signup,
		'/login': Login,
		'/profile': Profile,
		'/new-chat': NewChat,
	};

	loadUser();

	const sessionId = $derived(auth.session?.sessionId);

	$effect(() => {
		if (!sessionId) {
			return;
		}

		connect();
		return disconnect;
	});
</script>

<main>
	<Alert />
	<ConnectionBanner />
	<Router {routes} />
</main>

<style>
	main {
		display: flex;
		flex-direction: column;
		height: 100%;
		max-width: 480px;
		padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left);
		margin: 0 auto;
		border-left: 1px solid var(--divider);
		border-right: 1px solid var(--divider);
	}
</style>
