<script lang="ts">
	import Router from 'svelte-spa-router';

	import Alert from './components/Alert.svelte';

	import Index from './pages/Index.svelte';
	import Chat from './pages/Chat.svelte';
	import Signup from './pages/Signup.svelte';
	import Login from './pages/Login.svelte';
	import Profile from './pages/Profile.svelte';
	import Search from './pages/Search.svelte';

	import { auth, loadUser } from './lib/auth.svelte.js';
	import { connect, disconnect } from './lib/ws.js';

	const routes = {
		'/': Index,
		'/chats/:id': Chat,
		'/signup': Signup,
		'/login': Login,
		'/profile': Profile,
		'/search': Search,
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
	<Router {routes} />
</main>

<style>
	main {
		display: flex;
		flex-direction: column;
		height: 100%;
		max-width: 480px;
		margin: 0 auto;
		border-left: 1px solid var(--divider);
		border-right: 1px solid var(--divider);
	}
</style>
