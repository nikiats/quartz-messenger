import { auth, refresh } from './auth.svelte.js';
import type { Message } from './messages.js';

export type ServerEvent =
	| ({ method: 'receiveMessage' } & Message)
	| { method: 'authorizeResponse'; success: boolean }
	| { method: 'sendMessageResponse'; clientMessageId: string; success: boolean; seqnum?: number; sentAt?: string }
	| { method: 'requireAuthorize' };

type Sent = { seqnum: number; sentAt: string };

const pingInterval = 50_000;
const reconnectDelay = 3_000;

let socket: WebSocket | null = null;
let authorized = false;
let active = false;
let reconnectTimer: ReturnType<typeof setTimeout>;

const listeners = new Set<(event: ServerEvent) => void>();
const pending = new Map<string, { resolve: (sent: Sent) => void; reject: (err: Error) => void }>();

export function connect() {
	active = true;
	open();
}

export function disconnect() {
	active = false;
	clearTimeout(reconnectTimer);
	socket?.close();
	socket = null;
}

export function subscribe(listener: (event: ServerEvent) => void): () => void {
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
}

export function sendMessage(chatId: string, text: string): Promise<Sent> {
	if (!authorized) {
		return Promise.reject(new Error('Нет соединения с сервером'));
	}

	const clientMessageId = crypto.randomUUID();
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
		authorized = false;
		for (const request of pending.values()) {
			request.reject(new Error('Соединение потеряно'));
		}
		pending.clear();

		if (active) {
			reconnectTimer = setTimeout(open, reconnectDelay);
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
			authorized = event.success;
			if (!event.success) {
				refresh();
			}
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
