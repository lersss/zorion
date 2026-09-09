// web/static/js/main.js
import { state, elements } from './map/config.js';
import { resizeCanvas } from './map/render.js';
import { loadData } from './map/data.js';
import { handleCanvasClick, initFlyBtn, initPanZoom, initHover } from './map/events.js';
import { animationLoop } from './map/animation.js';
import { filterWorlds } from './map/data.js';
import { applyFiltersFromUI, resetFilters } from './filters.js';

function init() {
    elements.canvas.addEventListener('click', handleCanvasClick);
    window.addEventListener('resize', resizeCanvas);
    initFlyBtn();
    initPanZoom();
    initHover();

    // --- Инициализация фильтров ---
    const applyBtn = document.getElementById('apply-filters');
    const resetBtn = document.getElementById('reset-filters');

    if (applyBtn) {
        applyBtn.addEventListener('click', () => {
            applyFiltersFromUI();
            state.filteredWorlds = filterWorlds(state.worlds);
            resizeCanvas();
            const count = state.filteredWorlds.length;
            const statusEl = document.getElementById('filter-status');
            if (statusEl) {
                statusEl.textContent = `Показано: ${count} из ${state.worlds ? state.worlds.length : 0}`;
            }
        });
    }

    if (resetBtn) {
        resetBtn.addEventListener('click', () => {
            resetFilters();
            const hasPlanets = document.getElementById('filter-has-planets');
            const hasLife = document.getElementById('filter-life');
            const hasHabitable = document.getElementById('filter-habitable');
            const planetType = document.getElementById('filter-planet-type');
            const resource = document.getElementById('filter-resource');
            if (hasPlanets) hasPlanets.checked = false;
            if (hasLife) hasLife.checked = false;
            if (hasHabitable) hasHabitable.checked = false;
            if (planetType) planetType.value = '';
            if (resource) resource.value = '';
            state.filteredWorlds = null;
            resizeCanvas();
            const statusEl = document.getElementById('filter-status');
            if (statusEl) {
                statusEl.textContent = '';
            }
        });
    }

    loadData().then(() => {
        state.filteredWorlds = null;
        animationLoop();
    });
}

init();