import type { User } from './users.js';

type Session = {
	token: string;
	refreshToken: string;
	sessionId: string;
	userId: string;
};

const storageKey = 'session';

function restore(): Session | null {
	try {
		return JSON.parse(localStorage.getItem(storageKey) ?? 'null');
	} catch {
		return null;
	}
}

export const auth = $state<{ session: Session | null; user: User | null }>({
	session: restore(),
	user: null,
});

function save(session: Session | null) {
	auth.session = session;

	if (session) {
		localStorage.setItem(storageKey, JSON.stringify(session));
	} else {
		auth.user = null;
		localStorage.removeItem(storageKey);
	}
}

function userIdFrom(token: string): string {
	const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
	return JSON.parse(atob(payload)).sub;
}

function request(url: string, method: string, body: unknown, token?: string): Promise<Response> {
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}

	return fetch(url, { method, headers, body: body ? JSON.stringify(body) : undefined });
}

async function errorCode(res: Response): Promise<string | undefined> {
	const err = await res.json().catch(() => null);
	return err?.code;
}

export async function authFetch(url: string, method = 'GET', body?: unknown): Promise<Response> {
	let res = await request(url, method, body, auth.session?.token);

	if (res.status === 401 && (await errorCode(res.clone())) === 'session_expired' && (await refresh())) {
		res = await request(url, method, body, auth.session?.token);
	}
	if (res.status === 401) {
		save(null);
	}

	return res;
}

let refreshing: Promise<boolean> | null = null;

export function refresh(): Promise<boolean> {
	if (!refreshing) {
		const run = navigator.locks ? navigator.locks.request('token-refresh', refreshTokens) : refreshTokens();
		refreshing = run.finally(() => (refreshing = null));
	}

	return refreshing;
}

async function refreshTokens(): Promise<boolean> {
	const current = auth.session;
	if (!current) {
		return false;
	}

	const stored = restore();
	if (stored && stored.refreshToken !== current.refreshToken) {
		save(stored);
		return true;
	}

	const res = await request('/api/tokens', 'POST', { refreshToken: current.refreshToken });
	if (!res.ok) {
		save(null);
		return false;
	}

	const tokens = await res.json();
	save({ ...current, token: tokens.token, refreshToken: tokens.refreshToken });
	return true;
}

export async function register(login: string, password: string) {
	const res = await request('/api/users', 'POST', { login, password });
	if (res.ok) {
		return;
	}

	switch (await errorCode(res)) {
		case 'login_taken':
			throw new Error('Логин уже занят');
		case 'invalid_request':
			throw new Error('Некорректный логин или пароль');
		default:
			throw new Error('Ошибка сервера');
	}
}

export async function login(login: string, password: string) {
	const res = await request('/api/sessions', 'POST', { login, password });

	if (!res.ok) {
		switch (await errorCode(res)) {
			case 'authorization_failed':
				throw new Error('Неверный логин или пароль');
			case 'suspended_user':
				throw new Error('Аккаунт заблокирован');
			default:
				throw new Error('Ошибка сервера');
		}
	}

	const data = await res.json();
	save({ ...data, userId: userIdFrom(data.token) });
	await loadUser();
}

export async function loadUser() {
	if (!auth.session) {
		return;
	}

	const res = await authFetch(`/api/users/${auth.session.userId}`);
	if (res.ok) {
		auth.user = await res.json();
	}
}

export async function logout() {
	if (!auth.session) {
		return;
	}

	await authFetch(`/api/sessions/${auth.session.sessionId}`, 'DELETE').catch(() => {});
	save(null);
}
