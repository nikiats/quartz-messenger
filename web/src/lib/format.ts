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

export function formatDate(iso: string): string {
	const date = new Date(iso);
	const today = new Date();
	const yesterday = new Date(today);
	yesterday.setDate(today.getDate() - 1);

	if (date.toDateString() === today.toDateString()) {
		return 'Сегодня';
	}
	if (date.toDateString() === yesterday.toDateString()) {
		return 'Вчера';
	}

	const sameYear = date.getFullYear() === today.getFullYear();
	return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: sameYear ? undefined : 'numeric' });
}

export function isSameDay(a: string, b: string): boolean {
	return new Date(a).toDateString() === new Date(b).toDateString();
}
