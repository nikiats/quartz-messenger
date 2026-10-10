import { authFetch } from './auth.svelte.js';
import { sendOverSocket } from './ws.svelte.js';
import { pageSize } from '../config/chat.js';

export type Message = {
	id: string;
	chatId: string;
	senderId: string;
	sentAt: string;
	content: { text: string };
};

export type Sent = { id: string; sentAt: string };

type Page = { messages: Message[]; hasMore: boolean };

export const newestId = '9223372036854775807';

export function compareIds(a: string, b: string): number {
	const diff = BigInt(a) - BigInt(b);
	return diff > 0n ? 1 : diff < 0n ? -1 : 0;
}

export async function getMessages(chatId: string, direction: 'before' | 'after', id: string): Promise<Page> {
	const res = await authFetch(`/api/chats/${chatId}/messages?${direction}=${id}&limit=${pageSize}`);
	if (!res.ok) {
		throw new Error('Не удалось загрузить сообщения');
	}

	return res.json();
}

export async function searchMessages(chatId: string, query: string, before: string): Promise<Page> {
	const q = encodeURIComponent(query);
	const res = await authFetch(`/api/chats/${chatId}/messages?q=${q}&before=${before}&limit=${pageSize}`);
	if (!res.ok) {
		throw new Error('Не удалось выполнить поиск');
	}

	return res.json();
}

export async function sendMessage(chatId: string, clientMessageId: string, text: string): Promise<Sent> {
	try {
		return await sendOverSocket(chatId, clientMessageId, text);
	} catch {
		return sendOverHttp(chatId, clientMessageId, text);
	}
}

async function sendOverHttp(chatId: string, clientMessageId: string, text: string): Promise<Message> {
	const res = await authFetch(`/api/chats/${chatId}/messages`, 'POST', { clientMessageId, content: { text } });
	if (!res.ok) {
		throw new Error(res.status === 404 ? 'Чат не найден' : 'Не удалось отправить сообщение');
	}

	return res.json();
}
