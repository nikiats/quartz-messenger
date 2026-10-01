type AlertKind = 'info' | 'success' | 'error';

export const alertBox = $state<{ current: { kind: AlertKind; text: string } | null }>({
	current: null,
});

const hideAfter = 5000;
let timer: ReturnType<typeof setTimeout>;

export function showAlert(kind: AlertKind, text: string) {
	alertBox.current = { kind, text };

	clearTimeout(timer);
	timer = setTimeout(clearAlert, hideAfter);
}

export function clearAlert() {
	clearTimeout(timer);
	alertBox.current = null;
}
