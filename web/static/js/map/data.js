import { state, elements } from './config.js';
import { isFiniteNumber } from './utils.js';
import { resizeCanvas, draw } from './render.js';
import { centerOnAgent } from './navigation.js';

let isDataLoaded = false;

export async function loadUserData() {
    const token = localStorage.getItem('token');
    if (!token) return;
    try {
        const res = await fetch('/me', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (res.ok) {
            const data = await res.json();
            state.currentWorldId = data.current_world_id;
            elements.currentWorldNameEl.textContent = data.current_world_name || '—';
            console.log('User data loaded, currentWorldId:', state.currentWorldId);
        }
    } catch (e) {
        console.error('Failed to load user data:', e);
    }
}

export async function loadData() {
    console.log('loadData started');
    if (isDataLoaded) {
        console.log('loadData already loaded, skipping');
        return;
    }
    isDataLoaded = true;

    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        elements.statusBar.textContent = '❌ Не авторизован. Перейдите на /login-page';
        elements.loadingEl.style.display = 'none';
        isDataLoaded = false;
        return;
    }
    try {
        console.log('Loading user data...');
        await loadUserData();

        console.log('Fetching worlds...');
        const worldsRes = await fetch('/worlds', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!worldsRes.ok) {
            const errText = await worldsRes.text();
            console.error('Worlds fetch error:', errText);
            elements.statusBar.textContent = '❌ Ошибка загрузки миров: ' + errText;
            elements.loadingEl.style.display = 'none';
            isDataLoaded = false;
            return;
        }
        state.worlds = await worldsRes.json();
        console.log('Worlds fetched:', state.worlds.length);

        state.worlds = state.worlds.filter(w => isFiniteNumber(w.coord_x) && isFiniteNumber(w.coord_y));
        console.log('Worlds after filtering:', state.worlds.length);

        state.worlds.forEach(w => {
            w.level = Math.floor(Math.random() * 5) + 1;
            const types = ['tech', 'agri', 'military', 'trade', 'mixed'];
            w.type = types[Math.floor(Math.random() * types.length)];
        });
        console.log('Worlds processed with level and type');

        elements.loadingEl.style.display = 'none';

        if (state.worlds.length > 0) {
            let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
            state.worlds.forEach(w => {
                if (w.coord_x < minX) minX = w.coord_x;
                if (w.coord_x > maxX) maxX = w.coord_x;
                if (w.coord_y < minY) minY = w.coord_y;
                if (w.coord_y > maxY) maxY = w.coord_y;
            });
            const rangeX = maxX - minX || 1;
            const rangeY = maxY - minY || 1;
            const range = Math.max(rangeX, rangeY);
            if (range === 0) {
                state.scale = 1;
                state.offsetX = state.canvasWidth / 2;
                state.offsetY = state.canvasHeight / 2;
            } else {
                const centerX = (minX + maxX) / 2;
                const centerY = (minY + maxY) / 2;
                const padding = 80;
                const maxSize = Math.min(state.canvasWidth - padding * 2, state.canvasHeight - padding * 2);
                state.scale = maxSize / (range * 1.2);
                state.offsetX = state.canvasWidth / 2 - centerX * state.scale;
                state.offsetY = state.canvasHeight / 2 - centerY * state.scale;
            }
            console.log('Map scale:', state.scale, 'offsetX:', state.offsetX, 'offsetY:', state.offsetY);
        }
        resizeCanvas();

        if (!state.currentWorldId && state.worlds.length > 0) {
            console.log('No currentWorldId, centering on first world without travel');
            state.currentWorldId = state.worlds[0].id;
            elements.currentWorldNameEl.textContent = state.worlds[0].name;
        }

        console.log('Calling centerOnAgent...');
        setTimeout(() => {
            centerOnAgent();
        }, 100);
        elements.statusBar.textContent = 'Загружено миров: ' + state.worlds.length;
        console.log('loadData completed successfully');
        isDataLoaded = false;
    } catch (e) {
        console.error('Load error:', e);
        elements.statusBar.textContent = '❌ Ошибка загрузки данных: ' + e.message;
        elements.loadingEl.style.display = 'none';
        isDataLoaded = false;
    }
}