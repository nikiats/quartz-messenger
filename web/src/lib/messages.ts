import { authFetch } from './auth.svelte.js';

export type Message = {
	chatId: string;
	senderId: string;
	seqnum: number;
	sentAt: string;
	content: { text: string };
};

const pageSize = 50;

export async function getMessages(
	chatId: string,
	direction: 'before' | 'after',
	seqnum: number,
): Promise<{ messages: Message[]; hasMore: boolean }> {
	const res = await authFetch(`/api/chats/${chatId}/messages?${direction}=${seqnum}&limit=${pageSize}`);
	if (!res.ok) {
		throw new Error('Не удалось загрузить сообщения');
	}

	return res.json();
}
