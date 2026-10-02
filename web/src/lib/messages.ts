import { authFetch } from './auth.svelte.js';
import { sendOverSocket } from './ws.svelte.js';
import { pageSize } from '../config/chat.js';

export type Message = {
	chatId: string;
	senderId: string;
	seqnum: number;
	sentAt: string;
	content: { text: string };
};

type Page = { messages: Message[]; hasMore: boolean };

export async function getMessages(chatId: string, direction: 'before' | 'after', seqnum: number): Promise<Page> {
	const res = await authFetch(`/api/chats/${chatId}/messages?${direction}=${seqnum}&limit=${pageSize}`);
	if (!res.ok) {
		throw new Error('Не удалось загрузить сообщения');
	}

	return res.json();
}

export async function searchMessages(chatId: string, query: string, before: number): Promise<Page> {
	const q = encodeURIComponent(query);
	const res = await authFetch(`/api/chats/${chatId}/messages?q=${q}&before=${before}&limit=${pageSize}`);
	if (!res.ok) {
		throw new Error('Не удалось выполнить поиск');
	}

	return res.json();
}

export async function sendMessage(
	chatId: string,
	clientMessageId: string,
	text: string,
): Promise<{ seqnum: number; sentAt: string }> {
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
