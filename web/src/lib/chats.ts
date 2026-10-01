import { authFetch } from './auth.svelte.js';

export type Chat = {
	id: string;
	type: 'private' | 'group';
	title?: string;
	lastSeqnum: number;
	lastMessageAt: string;
};

export async function getChats(): Promise<Chat[]> {
	const res = await authFetch('/api/chats');
	if (!res.ok) {
		throw new Error('Не удалось загрузить чаты');
	}

	return res.json();
}

export async function getChat(chatId: string): Promise<Chat> {
	const res = await authFetch(`/api/chats/${chatId}`);
	if (!res.ok) {
		throw new Error(res.status === 404 ? 'Чат не найден' : 'Не удалось загрузить чат');
	}

	return res.json();
}

export async function createChat(type: Chat['type'], memberIds: string[], title?: string): Promise<Chat> {
	const res = await authFetch('/api/chats', 'POST', { type, title, memberIds });
	if (!res.ok) {
		throw new Error(res.status === 404 ? 'Пользователь не найден' : 'Не удалось создать чат');
	}

	return res.json();
}
