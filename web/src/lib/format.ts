export function formatTime(iso: string): string {
	return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function formatDay(iso: string): string {
	const date = new Date(iso);
	if (date.toDateString() === new Date().toDateString()) {
		return formatTime(iso);
	}

	return date.toLocaleDateString([], { day: '2-digit', month: '2-digit' });
}
