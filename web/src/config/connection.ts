// интервал ping по WebSocket, мс
export const pingInterval = 50_000;

// случайная пауза перед переподключением WebSocket, мс
export const reconnectDelay = {
	min: 2_000,
	max: 10_000,
};

// интервал опроса сервера без WebSocket, мс
export const pollInterval = 5_000;
