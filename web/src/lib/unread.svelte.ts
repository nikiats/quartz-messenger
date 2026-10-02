import { SvelteSet } from 'svelte/reactivity';
import { subscribe } from './ws.svelte.js';

export const unread = new SvelteSet<string>();

let openedChatId: string | null = null;

export function setOpenedChat(chatId: string | null) {
	openedChatId = chatId;
	if (chatId) {
		unread.delete(chatId);
	}
}

subscribe((event) => {
	if (event.method === 'receiveMessage' && event.chatId !== openedChatId) {
		unread.add(event.chatId);
	}
});
