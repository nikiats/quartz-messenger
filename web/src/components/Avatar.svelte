<script lang="ts">
	let { name = '', id = '', size = 48 }: { name?: string; id?: string; size?: number } = $props();

	const palette = ['#415A77', '#4E6A8C', '#5B7799', '#778DA9', '#34495F'];

	const color = $derived(palette[[...id].reduce((sum, char) => sum + char.charCodeAt(0), 0) % palette.length]);
	const letter = $derived(name.trim().charAt(0).toUpperCase() || '?');
</script>

<span
	class="avatar"
	style:width="{size}px"
	style:height="{size}px"
	style:font-size="{size * 0.4}px"
	style:background-color={color}
>
	{letter}
</span>

<style>
	.avatar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		border-radius: 50%;
		font-weight: 600;
		color: var(--text);
		user-select: none;
	}
</style>
