// web/static/js/main.js
import { state, elements } from './map/config.js';
import { resizeCanvas } from './map/map_render.js';
import { handleCanvasClick, initFlyBtn, initPanZoom, initHover } from './map/events.js';
import { animationLoop } from './map/animation.js';
import { centerOnAgent } from './map/navigation.js';
import { applyFiltersFromUI, resetFilters, filterState } from './filters.js';
import { loadClusters, loadUserData } from './map/data.js';
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
            if (elements.zoomInfo) {
                elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
            }
            return true;
        }
    } catch (e) { /* ignore */ }
    return false;
}

// --- Инициализация карты ---
async function initMap() {
    resizeCanvas();

    // 1. Пользователь (current_world_id)
    await loadUserData();

    // 2. Восстановление вьюпорта
    const restored = restoreViewport();
    if (!restored) {
        // Первый заход — центрируемся на игроке, если знаем мир
        if (state.currentWorldId) {
            centerOnAgent();
        }
    }

    // 3. Первая загрузка кластеров
    if (elements.loading) elements.loading.style.display = 'block';
    if (elements.statusBar) elements.statusBar.textContent = '⏳ Загрузка карты...';

    await loadClusters();

    if (elements.loading) elements.loading.style.display = 'none';
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

        const spinner = document.getElementById('filter-spinner');
        const countEl = document.getElementById('filter-count');
        if (spinner) spinner.style.display = 'inline-block';
        if (countEl) countEl.textContent = '...';

        try {
            await loadClusters();
            if (countEl) {
                const activeCount = [
                    filterState.hasPlanets,
                    filterState.hasLife,
                    filterState.hasHabitable,
                    filterState.planetType,
                    filterState.resourceCategory,
                ].filter(Boolean).length;
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
                await loadClusters();
                if (countEl) countEl.textContent = '';
            } catch (e) {
                console.error('Reset filter error:', e);
                if (countEl) countEl.textContent = '❌';
            } finally {
                if (spinner) spinner.style.display = 'none';
            }
        });
    }

    // --- Первая загрузка + запуск цикла анимации ---
    initMap().then(() => {
        animationLoop();
    }).catch(err => {
        console.error('initMap error:', err);
        if (elements.statusBar) elements.statusBar.textContent = '❌ Ошибка загрузки';
    });
}

init();