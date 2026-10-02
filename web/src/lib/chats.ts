import { authFetch } from './auth.svelte.js';

export type Chat = {
	id: string;
	type: 'private' | 'group';
	title?: string;
	lastSeqnum: number;
	lastMessageAt: string;
};

export type Member = {
	userId: string;
	role: 'admin' | 'regular';
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

export async function deleteChat(chatId: string) {
	const res = await authFetch(`/api/chats/${chatId}`, 'DELETE');

	switch (res.status) {
		case 204:
			return;
		case 403:
			throw new Error('Нужны права администратора');
		default:
			throw new Error('Не удалось удалить чат');
	}
}

export async function getMembers(chatId: string): Promise<Member[]> {
	const res = await authFetch(`/api/chats/${chatId}/members`);
	if (!res.ok) {
		throw new Error('Не удалось загрузить участников');
	}

	return res.json();
}

export async function addMember(chatId: string, userId: string) {
	const res = await authFetch(`/api/chats/${chatId}/members`, 'POST', { userId });

	switch (res.status) {
		case 201:
			return;
		case 403:
			throw new Error('Нужны права администратора');
		case 404:
			throw new Error('Пользователь не найден');
		case 409:
			throw new Error('Уже участник');
		default:
			throw new Error('Не удалось добавить участника');
	}
}

export async function removeMember(chatId: string, userId: string) {
	const res = await authFetch(`/api/chats/${chatId}/members/${userId}`, 'DELETE');

	switch (res.status) {
		case 204:
			return;
		case 403:
			throw new Error('Нужны права администратора');
		default:
			throw new Error('Не удалось изменить состав группы');
	}
}

export async function createChat(type: Chat['type'], memberIds: string[], title?: string): Promise<Chat> {
	const res = await authFetch('/api/chats', 'POST', { type, title, memberIds });
	if (!res.ok) {
		throw new Error(res.status === 404 ? 'Пользователь не найден' : 'Не удалось создать чат');
	}

	return res.json();
}
