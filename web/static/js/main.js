// web/static/js/main.js
import { state, elements } from './map/config.js';
import { resizeCanvas } from './map/render.js';
import { handleCanvasClick, initFlyBtn, initPanZoom, initHover } from './map/events.js';
import { animationLoop } from './map/animation.js';
import { centerOnAgent } from './map/navigation.js';
import { applyFiltersFromUI, resetFilters, filterState } from './filters.js';
import { openSystemModal } from './modal/index.js';
import { worldToCanvas, isFiniteNumber } from './map/utils.js';

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
async function loadAllData(filters) {
    const token = localStorage.getItem('token');
    if (!token) {
        console.warn('No token, redirect to login');
        window.location.href = '/login-page';
        return;
    }
    elements.loading.style.display = 'block';
    elements.statusBar.textContent = '⏳ Загрузка данных...';

    try {
        const worlds = await loadWorldsWithFilters(filters || null);
        state.worlds = worlds;
        state.filteredWorlds = null;

        await loadUserData();

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

    // --- Инициализация фильтров ---
    const applyBtn = document.getElementById('apply-filters');
    const resetBtn = document.getElementById('reset-filters');

    // Автоматическое применение фильтров при изменении любого элемента
    const filterInputs = document.querySelectorAll('#filters-bar input, #filters-bar select');
    filterInputs.forEach(el => {
        el.addEventListener('change', () => {
            // Триггерим применение фильтров
            applyFilters();
        });
    });

    async function applyFilters() {
        applyFiltersFromUI();
        const filters = {};
        if (filterState.hasPlanets) filters.hasPlanets = true;
        if (filterState.hasLife) filters.hasLife = true;
        if (filterState.hasHabitable) filters.hasHabitable = true;
        if (filterState.planetType) filters.planetType = filterState.planetType;
        if (filterState.resourceCategory) filters.resourceCategory = filterState.resourceCategory;

        // Показываем спиннер
        const spinner = document.getElementById('filter-spinner');
        const countEl = document.getElementById('filter-count');
        if (spinner) spinner.style.display = 'inline-block';
        if (countEl) countEl.textContent = '...';

        try {
            await loadAllData(filters);
            // Обновляем счётчик
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
                await loadAllData(null);
                if (countEl) countEl.textContent = '';
            } catch (e) {
                console.error('Reset filter error:', e);
                if (countEl) countEl.textContent = '❌';
            } finally {
                if (spinner) spinner.style.display = 'none';
            }
        });
    }

    // Первоначальная загрузка без фильтров
    loadAllData(null).then(() => {
        animationLoop();
    });
}

init();