import { state, elements } from './config.js';
import { worldToCanvas, isFiniteNumber } from './utils.js';
import { draw } from './render.js';
import { loadData } from './data.js';
import { centerOnAgent } from './navigation.js';
import { CONFIG } from '../config.js';

const { map: mapCfg, ui: uiCfg } = CONFIG;

// --- HOVER: определение мира под курсором ---
export function initHover() {
    elements.canvas.addEventListener('mousemove', (e) => {
        const rect = elements.canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
        const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

        let found = null;
        let minDist = mapCfg.minDistForClick;
        state.worlds.forEach(w => {
            const pos = worldToCanvas(w);
            if (!isFiniteNumber(pos.x) || !isFiniteNumber(pos.y)) return;
            const dist = Math.hypot(mouseX - pos.x, mouseY - pos.y);
            if (dist < minDist) {
                minDist = dist;
                found = w;
            }
        });

        // Обновляем состояние и перерисовываем только если изменилось
        if (found) {
            if (state.hoveredWorldId !== found.id) {
                state.hoveredWorldId = found.id;
                elements.canvas.style.cursor = 'pointer';
                draw();
            }
        } else {
            if (state.hoveredWorldId !== null) {
                state.hoveredWorldId = null;
                elements.canvas.style.cursor = 'crosshair';
                draw();
            }
        }
    });

    // При выходе мыши за пределы Canvas сбрасываем подсветку
    elements.canvas.addEventListener('mouseleave', () => {
        if (state.hoveredWorldId !== null) {
            state.hoveredWorldId = null;
            elements.canvas.style.cursor = 'crosshair';
            draw();
        }
    });
}

// --- КЛИК ---
export function handleCanvasClick(e) {
    const rect = elements.canvas.getBoundingClientRect();
    const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
    const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

    let found = null;
    let minDist = mapCfg.minDistForClick;
    state.worlds.forEach(w => {
        const pos = worldToCanvas(w);
        if (!isFiniteNumber(pos.x) || !isFiniteNumber(pos.y)) return;
        const dist = Math.hypot(mouseX - pos.x, mouseY - pos.y);
        if (dist < minDist) {
            minDist = dist;
            found = w;
        }
    });

    if (found) {
        // --- ОТКРЫВАЕМ МОДАЛКУ С СИСТЕМОЙ ---
        if (typeof window.openSystemModal === 'function') {
            window.openSystemModal(found.id, found.name, found.spectral_class || 'G');
        } else {
            console.warn('openSystemModal не загружена');
        }
        // Скрываем тултип
        elements.tooltip.classList.remove('active');
        state.selectedWorldId = found.id;
    } else {
        elements.tooltip.classList.remove('active');
        state.selectedWorldId = null;
    }
}

// --- КНОПКА "ЛЕТЕТЬ" (без изменений) ---
export function initFlyBtn() {
    elements.tooltipFlyBtn.addEventListener('click', async function(e) {
        e.stopPropagation();
        const worldId = this.dataset.worldId;
        if (!worldId) return;
        const token = localStorage.getItem('token');
        if (!token) {
            alert('Не авторизован');
            return;
        }
        if (state.isFlying) {
            alert('Уже в полёте');
            return;
        }
        try {
            const res = await fetch('/travel', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + token
                },
                body: JSON.stringify({ world_id: worldId })
            });
            const text = await res.text();
            if (!res.ok) {
                alert('Ошибка: ' + text);
                return;
            }
            const data = JSON.parse(text);
            const fromWorld = state.worlds.find(w => w.id === data.from);
            const toWorld = state.worlds.find(w => w.id === data.to);
            if (fromWorld && toWorld) {
                state.flyFrom = fromWorld;
                state.flyTo = toWorld;
                state.flyDuration = data.duration;
                state.flyStartTime = Date.now();
                state.isFlying = true;
                elements.tooltip.classList.remove('active');
                draw();
            }
        } catch (e) {
            alert('Ошибка: ' + e.message);
        }
    });
}

// --- ПАНОРАМИРОВАНИЕ И ЗУМ (с обновлением курсора) ---
export function initPanZoom() {
    elements.canvas.addEventListener('mousedown', (e) => {
        if (e.target === elements.canvas) {
            state.isDragging = true;
            state.dragStartX = e.clientX;
            state.dragStartY = e.clientY;
            state.dragStartOffsetX = state.offsetX;
            state.dragStartOffsetY = state.offsetY;
            elements.canvas.style.cursor = 'grabbing';
        }
    });

    window.addEventListener('mousemove', (e) => {
        if (state.isDragging) {
            const dx = e.clientX - state.dragStartX;
            const dy = e.clientY - state.dragStartY;
            state.offsetX = state.dragStartOffsetX + dx;
            state.offsetY = state.dragStartOffsetY + dy;
            draw();
        }
    });

    window.addEventListener('mouseup', () => {
        if (state.isDragging) {
            state.isDragging = false;
            // Возвращаем курсор на crosshair (если не hover)
            if (state.hoveredWorldId === null) {
                elements.canvas.style.cursor = 'crosshair';
            } else {
                elements.canvas.style.cursor = 'pointer';
            }
        }
    });

    elements.canvas.addEventListener('wheel', (e) => {
        e.preventDefault();
        const rect = elements.canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
        const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

        const delta = e.deltaY > 0 ? mapCfg.wheelSensitivity : 1 / mapCfg.wheelSensitivity;
        const newScale = Math.min(Math.max(state.scale * delta, mapCfg.minZoom), mapCfg.maxZoom);
        if (newScale === state.scale) return;

        const worldX = (mouseX - state.offsetX) / state.scale;
        const worldY = (mouseY - state.offsetY) / state.scale;
        state.scale = newScale;
        state.offsetX = mouseX - worldX * state.scale;
        state.offsetY = mouseY - worldY * state.scale;

        elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
        draw();
    }, { passive: false });

    document.getElementById('zoomInBtn').addEventListener('click', () => {
        const centerX = state.canvasWidth / 2;
        const centerY = state.canvasHeight / 2;
        const worldX = (centerX - state.offsetX) / state.scale;
        const worldY = (centerY - state.offsetY) / state.scale;
        state.scale = Math.min(state.scale * mapCfg.zoomStep, mapCfg.maxZoom);
        state.offsetX = centerX - worldX * state.scale;
        state.offsetY = centerY - worldY * state.scale;
        elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
        draw();
    });

    document.getElementById('zoomOutBtn').addEventListener('click', () => {
        const centerX = state.canvasWidth / 2;
        const centerY = state.canvasHeight / 2;
        const worldX = (centerX - state.offsetX) / state.scale;
        const worldY = (centerY - state.offsetY) / state.scale;
        state.scale = Math.max(state.scale / mapCfg.zoomStep, mapCfg.minZoom);
        state.offsetX = centerX - worldX * state.scale;
        state.offsetY = centerY - worldY * state.scale;
        elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
        draw();
    });

    document.getElementById('centerBtn').addEventListener('click', centerOnAgent);
}