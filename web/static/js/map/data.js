// web/static/js/map/data.js
import { state, elements } from './config.js';
import { draw } from './map_render.js';
import { filterState } from '../filters.js';

// Размер ячейки кластеризации на экране, в пикселях.
// Должен совпадать с CLUSTER_CELL_PX в map_render.js (для визуального соответствия).
export const CLUSTER_CELL_PX = 40;

// Задержка перед перезапросом после zoom/pan.
// За это время пользователь может ещё подвигать карту — лишний запрос не уйдёт.
const RELOAD_DEBOUNCE_MS = 180;

let loadingData = false;
let pendingReload = false;
let reloadTimer = null;
let currentWorldIdLoaded = false;

// ==================== DEBOUNCE ====================

export function scheduleReload() {
    if (reloadTimer) clearTimeout(reloadTimer);
    reloadTimer = setTimeout(() => {
        reloadTimer = null;
        loadClusters().catch(err => console.error('scheduleReload:', err));
    }, RELOAD_DEBOUNCE_MS);
}

// ==================== ЗАГРУЗКА КЛАСТЕРОВ ====================

export async function loadClusters() {
    if (loadingData) {
        pendingReload = true;
        return;
    }
    loadingData = true;

    try {
        const token = localStorage.getItem('token');
        if (!token) throw new Error('No token');

        const bounds = getViewportBounds();
        const cell = getCellSize();
        const url = buildUrl(bounds, cell);

        const res = await fetch(url, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (res.status === 401 || res.status === 403) {
            localStorage.removeItem('token');
            if (window.location.pathname !== '/login-page') {
                window.location.href = '/login-page';
            }
            return;
        }
        if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`);

        const clusters = await res.json();
        state.clusters = Array.isArray(clusters) ? clusters : [];

        // Обновляем кэш отдельных миров — из кластеров cnt=1.
        for (const c of state.clusters) {
            if (c.cnt === 1 && c.sid) {
                if (!state.worlds.some(w => w.id === c.sid)) {
                    state.worlds.push({
                        id: c.sid,
                        name: c.sname || '—',
                        spectral_class: c.sspec || 'G',
                        coord_x: c.x,
                        coord_y: c.y,
                    });
                }
            }
        }

        if (elements.loading) elements.loading.style.display = 'none';
    } catch (err) {
        console.error('loadClusters error:', err);
        if (elements.statusBar) elements.statusBar.textContent = '❌ Ошибка: ' + err.message;
    } finally {
        loadingData = false;
        if (pendingReload) {
            pendingReload = false;
            scheduleReload();
        }
    }

    draw();
}

// ==================== ГРАНИЦЫ VIEWPORT ====================

function getViewportBounds() {
    const invScale = 1 / state.scale;
    return {
        xMin: -state.offsetX * invScale,
        xMax: (state.canvasWidth - state.offsetX) * invScale,
        yMin: -state.offsetY * invScale,
        yMax: (state.canvasHeight - state.offsetY) * invScale,
    };
}

function getCellSize() {
    return CLUSTER_CELL_PX / state.scale;
}

// ==================== URL ====================

function buildUrl(bounds, cell) {
    const params = new URLSearchParams();
    params.set('x_min', bounds.xMin.toFixed(3));
    params.set('x_max', bounds.xMax.toFixed(3));
    params.set('y_min', bounds.yMin.toFixed(3));
    params.set('y_max', bounds.yMax.toFixed(3));
    params.set('cell', cell.toFixed(3));

    if (filterState.hasPlanets) params.set('has_planets', 'true');
    if (filterState.hasLife) params.set('has_life', 'true');
    if (filterState.hasHabitable) params.set('has_habitable', 'true');
    if (filterState.planetType) params.set('planet_type', filterState.planetType);
    if (filterState.resourceCategory) params.set('resource_category', filterState.resourceCategory);

    return '/api/worlds/filter?' + params.toString();
}

// ==================== ПОЛЬЗОВАТЕЛЬ ====================

export async function loadUserData() {
    if (currentWorldIdLoaded) return;
    try {
        const token = localStorage.getItem('token');
        if (!token) return;
        const res = await fetch('/me', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!res.ok) return;
        const user = await res.json();
        if (user.current_world_id) {
            state.currentWorldId = user.current_world_id;
            const world = state.worlds.find(w => w.id === user.current_world_id);
            if (world) {
                document.getElementById('currentWorldName').textContent = world.name;
            }
        }
        currentWorldIdLoaded = true;
    } catch (e) {
        console.warn('loadUserData error:', e);
    }
}

// ==================== ОБРАТНАЯ СОВМЕСТИМОСТЬ ====================

// Раньше эту функцию звали из main.js / animation.js.
// Теперь это алиас на loadClusters — чтобы не переписывать импорты.
export function loadData() {
    return loadClusters();
}

// Заглушка, оставленная для совместимости со старым кодом main.js.
export function filterWorlds(worlds) {
    return worlds || [];
}