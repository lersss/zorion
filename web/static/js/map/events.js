// web/static/js/map/events.js
import { state, elements } from './config.js';
import { worldToCanvas, isFiniteNumber } from './utils.js';
import { draw } from './map_render.js';
import { loadData } from './data.js';
import { centerOnAgent } from './navigation.js';
import { CONFIG } from '../config.js';
import { openSystemModal } from '../modal/index.js';

const { map: mapCfg, ui: uiCfg } = CONFIG;

// --- HOVER ---
export function initHover() {
    elements.canvas.addEventListener('mousemove', (e) => {
        const rect = elements.canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
        const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

        let found = null;
        let minDist = mapCfg.minDistForClick;
        const worldsToCheck = state.filteredWorlds || state.worlds || [];
        worldsToCheck.forEach(w => {
            const pos = worldToCanvas(w);
            if (!isFiniteNumber(pos.x) || !isFiniteNumber(pos.y)) return;
            const dist = Math.hypot(mouseX - pos.x, mouseY - pos.y);
            if (dist < minDist) {
                minDist = dist;
                found = w;
            }
        });

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

    elements.canvas.addEventListener('mouseleave', () => {
        if (state.hoveredWorldId !== null) {
            state.hoveredWorldId = null;
            elements.canvas.style.cursor = 'crosshair';
            draw();
        }
    });
}

// --- CLICK ---
export function handleCanvasClick(e) {
    // Если был drag (перемещение более чем на 5 пикселей), игнорируем клик
    if (state.isDragging) {
        return;
    }
    // Также проверяем, что мышь переместилась не слишком далеко от места нажатия
    if (state.dragStartX !== undefined && state.dragStartY !== undefined) {
        const dx = e.clientX - state.dragStartX;
        const dy = e.clientY - state.dragStartY;
        if (Math.hypot(dx, dy) > 5) {
            return;
        }
    }

    const rect = elements.canvas.getBoundingClientRect();
    const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
    const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

    let found = null;
    let minDist = mapCfg.minDistForClick;
    const worldsToCheck = state.filteredWorlds || state.worlds || [];
    worldsToCheck.forEach(w => {
        const pos = worldToCanvas(w);
        if (!isFiniteNumber(pos.x) || !isFiniteNumber(pos.y)) return;
        const dist = Math.hypot(mouseX - pos.x, mouseY - pos.y);
        if (dist < minDist) {
            minDist = dist;
            found = w;
        }
    });

    if (found) {
        if (typeof openSystemModal === 'function') {
            openSystemModal(found.id, found.name, found.spectral_class || 'G');
        } else {
            console.warn('openSystemModal not loaded');
        }
        elements.tooltip.classList.remove('active');
        state.selectedWorldId = found.id;
    } else {
        elements.tooltip.classList.remove('active');
        state.selectedWorldId = null;
    }
}

// --- КНОПКА "ЛЕТЕТЬ" ---
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

// --- PAN / ZOOM ---
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

    window.addEventListener('mouseup', (e) => {
        if (state.isDragging) {
            state.isDragging = false;
            // Сохраняем конечную позицию мыши для проверки клика
            state.dragEndX = e.clientX;
            state.dragEndY = e.clientY;
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