// web/static/js/map/data.js
import { state, elements } from './config.js';
import { isFiniteNumber, worldToCanvas, getStarColor } from './utils.js';
import { CONFIG } from '../config.js';
import { centerOnAgent } from './navigation.js';
import { filterState } from '../filters.js';
import { resizeCanvas } from './map_render.js'; // <-- обновлён импорт

const { map: mapCfg } = CONFIG;

let loadingData = false;
let dataLoaded = false;

export async function loadData() {
    if (loadingData) return;
    loadingData = true;
    if (elements.loading) elements.loading.style.display = 'block';
    if (elements.statusBar) elements.statusBar.textContent = '⏳ Загрузка данных...';

    try {
        const token = localStorage.getItem('token');
        if (!token) {
            throw new Error('No token');
        }
        const res = await fetch('/worlds', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!res.ok) {
            throw new Error('Failed to fetch worlds');
        }
        const worlds = await res.json();
        state.worlds = worlds;
        dataLoaded = true;

        await loadUserData();

        state.filteredWorlds = filterWorlds(state.worlds);

        if (state.currentWorldId) {
            centerOnAgent();
        } else {
            if (state.worlds.length > 0) {
                const first = state.worlds[0];
                const pos = worldToCanvas(first);
                if (isFiniteNumber(pos.x) && isFiniteNumber(pos.y)) {
                    state.offsetX = state.canvasWidth / 2 - pos.x;
                    state.offsetY = state.canvasHeight / 2 - pos.y;
                }
            }
        }

        resizeCanvas();
        if (elements.loading) elements.loading.style.display = 'none';
        if (elements.statusBar) {
            elements.statusBar.textContent = `${state.worlds.length} миров загружено`;
        }
        return state.worlds;
    } catch (err) {
        console.error('Load data error:', err);
        if (elements.loading) elements.loading.textContent = '❌ Ошибка загрузки данных';
        if (elements.statusBar) elements.statusBar.textContent = '❌ Ошибка';
        throw err;
    } finally {
        loadingData = false;
    }
}

async function loadUserData() {
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
    } catch (e) {
        console.warn('Failed to load user data:', e);
    }
}

export function filterWorlds(worlds) {
    if (!worlds || worlds.length === 0) return [];

    const hasActiveFilters = filterState.hasPlanets || filterState.hasLife || filterState.hasHabitable ||
                             filterState.planetType || filterState.resourceCategory;
    if (!hasActiveFilters) return worlds;

    return worlds.filter(world => {
        if (!world.planets || world.planets.length === 0) {
            if (filterState.hasPlanets) return false;
            if (filterState.hasLife) return false;
            if (filterState.hasHabitable) return false;
            if (filterState.planetType) return false;
            if (filterState.resourceCategory) return false;
            return true;
        }

        if (filterState.hasPlanets && world.planets.length === 0) return false;

        if (filterState.hasLife) {
            const hasLife = world.planets.some(p => p.life === true);
            if (!hasLife) return false;
        }

        if (filterState.hasHabitable) {
            const hasHabitable = world.planets.some(p => p.habitable === true);
            if (!hasHabitable) return false;
        }

        if (filterState.planetType) {
            const hasType = world.planets.some(p => (p.type || '').toLowerCase() === filterState.planetType.toLowerCase());
            if (!hasType) return false;
        }

        if (filterState.resourceCategory) {
            const hasResource = world.planets.some(p => {
                if (!p.resources) return false;
                const val = p.resources[filterState.resourceCategory];
                return typeof val === 'number' && val > 0.3;
            });
            if (!hasResource) return false;
        }

        return true;
    });
}