import { auth, refresh } from './auth.svelte.js';
import { showAlert } from './alert.svelte.js';
import { pingInterval, reconnectDelay } from '../config/connection.js';
import type { Message } from './messages.js';

export type ServerEvent =
	| ({ method: 'receiveMessage' } & Message)
	| { method: 'authorizeResponse'; success: boolean }
	| { method: 'sendMessageResponse'; clientMessageId: string; success: boolean; seqnum?: number; sentAt?: string }
	| { method: 'requireAuthorize' };

type Sent = { seqnum: number; sentAt: string };

export const connection = $state<{ status: 'connecting' | 'online' | 'offline' }>({ status: 'connecting' });

let socket: WebSocket | null = null;
let active = false;
let reconnectTimer: ReturnType<typeof setTimeout>;

const listeners = new Set<(event: ServerEvent) => void>();
const pending = new Map<string, { resolve: (sent: Sent) => void; reject: (err: Error) => void }>();

export function connect() {
	active = true;
	connection.status = 'connecting';
	open();
}

export function disconnect() {
	active = false;
	clearTimeout(reconnectTimer);
	socket?.close();
	socket = null;
	connection.status = 'connecting';
}

export function subscribe(listener: (event: ServerEvent) => void): () => void {
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
}

export function sendOverSocket(chatId: string, clientMessageId: string, text: string): Promise<Sent> {
	if (connection.status !== 'online') {
		return Promise.reject(new Error('Нет соединения с сервером'));
	}

	return new Promise((resolve, reject) => {
		pending.set(clientMessageId, { resolve, reject });
		send({ method: 'sendMessage', chatId, clientMessageId, content: { text } });
	});
}

function open() {
	const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
	const ws = new WebSocket(`${protocol}//${location.host}/ws`);
	socket = ws;

	let pingTimer: ReturnType<typeof setInterval>;

	ws.onopen = () => {
		authorize();
		pingTimer = setInterval(() => send({ method: 'ping' }), pingInterval);
	};

	ws.onmessage = (e) => handle(JSON.parse(e.data));

	ws.onclose = () => {
		clearInterval(pingTimer);
		if (socket !== ws) {
			return;
		}

		socket = null;
		connection.status = 'offline';
		for (const request of pending.values()) {
			request.reject(new Error('Соединение потеряно'));
		}
		pending.clear();

		if (active) {
			const delay = reconnectDelay.min + Math.random() * (reconnectDelay.max - reconnectDelay.min);
			reconnectTimer = setTimeout(open, delay);
		}
	};
}

function send(data: object) {
	socket?.send(JSON.stringify(data));
}

function authorize() {
	send({ method: 'authorize', token: auth.session?.token });
}

function handle(event: ServerEvent) {
	switch (event.method) {
		case 'authorizeResponse':
			if (!event.success) {
				refresh();
				break;
			}
			if (connection.status === 'offline') {
				showAlert('success', 'Соединение восстановлено');
			}
			connection.status = 'online';
			break;
		case 'requireAuthorize':
			refresh().then((ok) => ok && authorize());
			break;
		case 'sendMessageResponse': {
			const request = pending.get(event.clientMessageId);
			pending.delete(event.clientMessageId);
			if (event.success && event.seqnum && event.sentAt) {
				request?.resolve({ seqnum: event.seqnum, sentAt: event.sentAt });
			} else {
				request?.reject(new Error('Не удалось отправить сообщение'));
			}
			break;
		}
	}

	for (const listener of listeners) {
		listener(event);
	}
}
