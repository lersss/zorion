// web/static/js/main.js
import { state, elements } from './map/config.js';
import { resizeCanvas } from './map/map_render.js';
import { handleCanvasClick, initFlyBtn, initPanZoom, initHover } from './map/events.js';
import { animationLoop } from './map/animation.js';
import { centerOnAgent } from './map/navigation.js';
import { applyFiltersFromUI, resetFilters, filterState } from './filters.js';
import { openSystemModal } from './modal/index.js';
import { worldToCanvas, isFiniteNumber } from './map/utils.js';
import { draw } from './map/map_render.js';

// --- Восстановление вьюпорта из sessionStorage ---
function restoreViewport() {
    try {
        const saved = sessionStorage.getItem('viewport');
        if (saved) {
            const vp = JSON.parse(saved);
            state.offsetX = vp.offsetX || 0;
            state.offsetY = vp.offsetY || 0;
            state.scale = vp.scale || 1;
            return true;
        }
    } catch (e) { /* ignore */ }
    return false;
}

// --- Загрузка миров с фильтрацией ---
async function loadWorldsWithFilters(filters) {
    const token = localStorage.getItem('token');
    if (!token) {
        throw new Error('No token');
    }
    let url = '/api/worlds/filter';
    if (filters) {
        const params = new URLSearchParams();
        if (filters.hasPlanets) params.append('has_planets', 'true');
        if (filters.hasLife) params.append('has_life', 'true');
        if (filters.hasHabitable) params.append('has_habitable', 'true');
        if (filters.planetType) params.append('planet_type', filters.planetType);
        if (filters.resourceCategory) params.append('resource_category', filters.resourceCategory);
        const query = params.toString();
        if (query) url += '?' + query;
    }
    const res = await fetch(url, {
        headers: { 'Authorization': 'Bearer ' + token }
    });
    if (!res.ok) {
        throw new Error('Failed to fetch worlds');
    }
    return res.json();
}

// --- Загрузка всех данных (миры + пользователь) ---
async function loadAllData(filters, keepViewport = false) {
    const token = localStorage.getItem('token');
    if (!token) {
        console.warn('No token, redirect to login');
        window.location.href = '/login-page';
        return;
    }
    elements.loading.style.display = 'block';
    elements.statusBar.textContent = '⏳ Загрузка данных...';

    const savedOffsetX = keepViewport ? state.offsetX : null;
    const savedOffsetY = keepViewport ? state.offsetY : null;
    const savedScale = keepViewport ? state.scale : null;

    try {
        const worlds = await loadWorldsWithFilters(filters || null);
        state.worlds = worlds;
        state.filteredWorlds = null;

        await loadUserData();

        if (savedOffsetX !== null && savedOffsetY !== null && savedScale !== null) {
            state.offsetX = savedOffsetX;
            state.offsetY = savedOffsetY;
            state.scale = savedScale;
        } else {
            // Если не нужно сохранять вьюпорт, либо центрируем, либо восстанавливаем из sessionStorage
            if (!restoreViewport()) {
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
            }
        }

        resizeCanvas();
        elements.loading.style.display = 'none';
        elements.statusBar.textContent = `${state.worlds.length} миров загружено`;
    } catch (err) {
        console.error('Load data error:', err);
        elements.loading.textContent = '❌ Ошибка загрузки данных';
        elements.statusBar.textContent = '❌ Ошибка';
        if (err.message === 'No token') {
            window.location.href = '/login-page';
        }
        throw err;
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

function init() {
    elements.canvas.addEventListener('click', handleCanvasClick);
    window.addEventListener('resize', resizeCanvas);
    initFlyBtn();
    initPanZoom();
    initHover();

    // --- Автоматическое применение фильтров ---
    const filterInputs = document.querySelectorAll('#filters-bar input, #filters-bar select');
    let timeoutId = null;

    async function applyFilters() {
        applyFiltersFromUI();
        const filters = {};
        if (filterState.hasPlanets) filters.hasPlanets = true;
        if (filterState.hasLife) filters.hasLife = true;
        if (filterState.hasHabitable) filters.hasHabitable = true;
        if (filterState.planetType) filters.planetType = filterState.planetType;
        if (filterState.resourceCategory) filters.resourceCategory = filterState.resourceCategory;

        const spinner = document.getElementById('filter-spinner');
        const countEl = document.getElementById('filter-count');
        if (spinner) spinner.style.display = 'inline-block';
        if (countEl) countEl.textContent = '...';

        try {
            await loadAllData(filters, true);
            if (countEl) {
                const activeCount = Object.keys(filters).length;
                countEl.textContent = activeCount > 0 ? `(${activeCount})` : '';
            }
        } catch (e) {
            console.error('Filter apply error:', e);
            if (countEl) countEl.textContent = '❌';
        } finally {
            if (spinner) spinner.style.display = 'none';
        }
    }

    filterInputs.forEach(el => {
        el.addEventListener('change', () => {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(applyFilters, 300);
        });
    });

    const resetBtn = document.getElementById('reset-filters');
    if (resetBtn) {
        resetBtn.addEventListener('click', async () => {
            resetFilters();
            document.getElementById('filter-has-planets').checked = false;
            document.getElementById('filter-life').checked = false;
            document.getElementById('filter-habitable').checked = false;
            document.getElementById('filter-planet-type').value = '';
            document.getElementById('filter-resource').value = '';
            const spinner = document.getElementById('filter-spinner');
            const countEl = document.getElementById('filter-count');
            if (spinner) spinner.style.display = 'inline-block';
            if (countEl) countEl.textContent = '...';

            try {
                await loadAllData(null, true);
                if (countEl) countEl.textContent = '';
            } catch (e) {
                console.error('Reset filter error:', e);
                if (countEl) countEl.textContent = '❌';
            } finally {
                if (spinner) spinner.style.display = 'none';
            }
        });
    }

    loadAllData(null, false).then(() => {
        animationLoop();
    });
}

init();