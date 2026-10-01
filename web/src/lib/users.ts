import { authFetch } from './auth.svelte.js';

export type User = {
	userId: string;
	username?: string;
	visibleName?: string;
};

const users = new Map<string, Promise<User>>();

export function getUser(userId: string): Promise<User> {
	if (!users.has(userId)) {
		users.set(
			userId,
			authFetch(`/api/users/${userId}`).then((res) => res.json()),
		);
	}

	return users.get(userId)!;
}

export async function searchUsers(query: string): Promise<User[]> {
	const res = await authFetch(`/api/users?q=${encodeURIComponent(query)}`);
	if (!res.ok) {
		throw new Error('Не удалось найти пользователей');
	}

	return res.json();
}

export async function updateProfile(visibleName?: string, username?: string) {
	const res = await authFetch('/api/profile', 'PATCH', { visibleName, username });

	switch (res.status) {
		case 200:
			return;
		case 409:
			throw new Error('Username уже занят');
		case 400:
			throw new Error('Некорректное имя или username');
		default:
			throw new Error('Не удалось сохранить профиль');
	}
}
