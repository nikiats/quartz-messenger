import { alertDuration } from '../config/alerts.js';

type AlertKind = 'info' | 'success' | 'error';

export const alertBox = $state<{ current: { kind: AlertKind; text: string } | null }>({
	current: null,
});

let timer: ReturnType<typeof setTimeout>;

export function showAlert(kind: AlertKind, text: string) {
	alertBox.current = { kind, text };

	clearTimeout(timer);
	timer = setTimeout(clearAlert, alertDuration);
}

export function clearAlert() {
	clearTimeout(timer);
	alertBox.current = null;
}
