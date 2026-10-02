<script lang="ts">
	import type { Snippet } from 'svelte';
	import Row from './Row.svelte';
	import type { User } from '../lib/users.js';

	let {
		user,
		subtitle = user.username ? `@${user.username}` : '',
		onclick,
		trailing,
	}: {
		user: User;
		subtitle?: string;
		onclick?: () => void;
		trailing?: Snippet;
	} = $props();
</script>

<Row id={user.userId} title={user.visibleName ?? ''} {subtitle} {onclick} {trailing} />
