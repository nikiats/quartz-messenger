<script lang="ts">
	let {
		name,
		size = 24,
	}: { name: 'back' | 'close' | 'send' | 'plus' | 'pencil' | 'group'; size?: number } = $props();

	const paths = {
		back: 'M15 5l-7 7 7 7',
		close: 'M6 6l12 12M18 6L6 18',
		send: 'M12 19V5M5 12l7-7 7 7',
		plus: 'M12 5v14M5 12h14',
		pencil: 'M4 20h4L19 9l-4-4L4 16v4zM13.5 6.5l4 4',
		group: 'M9 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7zM2.5 20a6.5 6.5 0 0 1 13 0M16 4.5a3.5 3.5 0 0 1 0 6.5M18 14a6 6 0 0 1 3.5 6',
	};
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width="2"
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	<path d={paths[name]} />
</svg>
