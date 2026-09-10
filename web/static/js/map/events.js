// web/static/js/map/events.js
import { state, elements } from './config.js';
import { isFiniteNumber } from './utils.js';
import { draw, clusterScreenRadius } from './map_render.js';
import { scheduleReload } from './data.js';
import { centerOnAgent } from './navigation.js';
import { CONFIG } from '../config.js';
import { openSystemModal } from '../modal/index.js';

const { map: mapCfg } = CONFIG;

// ==================== СОХРАНЕНИЕ VIEWPORT ====================

function saveViewport() {
    try {
        sessionStorage.setItem('viewport', JSON.stringify({
            offsetX: state.offsetX,
            offsetY: state.offsetY,
            scale: state.scale
        }));
    } catch (e) { /* ignore */ }
}

// ==================== ПОИСК КЛАСТЕРА ПОД КУРСОРОМ ====================

function findClusterAt(mouseX, mouseY) {
    const clusters = state.clusters || [];
    let found = null;
    let bestDist = Infinity;

    for (const c of clusters) {
        const px = c.x * state.scale + state.offsetX;
        const py = c.y * state.scale + state.offsetY;
        if (!isFiniteNumber(px) || !isFiniteNumber(py)) continue;

        const radius = clusterScreenRadius(c);
        const dist = Math.hypot(mouseX - px, mouseY - py);
        const hitRadius = Math.max(radius + 3, mapCfg.minDistForClick);

        if (dist < hitRadius && dist < bestDist) {
            bestDist = dist;
            found = { cluster: c, screenX: px, screenY: py };
        }
    }
    return found;
}

// ==================== HOVER ====================

export function initHover() {
    elements.canvas.addEventListener('mousemove', (e) => {
        const rect = elements.canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
        const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

        const hit = findClusterAt(mouseX, mouseY);
        const newHoveredId = hit && hit.cluster.cnt === 1 ? hit.cluster.sid : null;

        if (state.hoveredWorldId !== newHoveredId) {
            state.hoveredWorldId = newHoveredId;
            elements.canvas.style.cursor = hit ? 'pointer' : 'crosshair';
            draw();
        } else if (hit) {
            elements.canvas.style.cursor = 'pointer';
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

// ==================== CLICK ====================

export function handleCanvasClick(e) {
    if (state.isDragging) return;
    if (state.dragStartX !== undefined && state.dragStartY !== undefined) {
        const dx = e.clientX - state.dragStartX;
        const dy = e.clientY - state.dragStartY;
        if (Math.hypot(dx, dy) > 5) return;
    }

    const rect = elements.canvas.getBoundingClientRect();
    const mouseX = (e.clientX - rect.left) * (elements.canvas.width / rect.width);
    const mouseY = (e.clientY - rect.top) * (elements.canvas.height / rect.height);

    const hit = findClusterAt(mouseX, mouseY);

    if (!hit) {
        elements.tooltip.classList.remove('active');
        state.selectedWorldId = null;
        return;
    }

    const c = hit.cluster;

    if (c.cnt === 1) {
        // Одиночный мир — открываем модалку
        if (typeof openSystemModal === 'function') {
            openSystemModal(c.sid, c.sname || '—', c.sspec || 'G');
        }
        elements.tooltip.classList.remove('active');
        state.selectedWorldId = c.sid;
    } else {
        // Кластер — зуммируем к его центру
        const targetScale = Math.min(state.scale * 2, mapCfg.maxZoom);
        const worldX = c.x;
        const worldY = c.y;
        state.scale = targetScale;
        state.offsetX = state.canvasWidth / 2 - worldX * targetScale;
        state.offsetY = state.canvasHeight / 2 - worldY * targetScale;
        if (elements.zoomInfo) {
            elements.zoomInfo.textContent = Math.round(targetScale * 100) + '%';
        }
        draw();
        saveViewport();
        scheduleReload();
    }
}

// ==================== КНОПКА "ЛЕТЕТЬ" ====================

export function initFlyBtn() {
    elements.tooltipFlyBtn.addEventListener('click', async function (e) {
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

// ==================== PAN / ZOOM ====================

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
            if (state.hoveredWorldId === null) {
                elements.canvas.style.cursor = 'crosshair';
            } else {
                elements.canvas.style.cursor = 'pointer';
            }
            saveViewport();
            scheduleReload();
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

        if (elements.zoomInfo) elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
        draw();
        saveViewport();
        scheduleReload();
    }, { passive: false });

    document.getElementById('zoomInBtn').addEventListener('click', () => {
        const centerX = state.canvasWidth / 2;
        const centerY = state.canvasHeight / 2;
        const worldX = (centerX - state.offsetX) / state.scale;
        const worldY = (centerY - state.offsetY) / state.scale;
        state.scale = Math.min(state.scale * mapCfg.zoomStep, mapCfg.maxZoom);
        state.offsetX = centerX - worldX * state.scale;
        state.offsetY = centerY - worldY * state.scale;
        if (elements.zoomInfo) elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
        draw();
        saveViewport();
        scheduleReload();
    });

    document.getElementById('zoomOutBtn').addEventListener('click', () => {
        const centerX = state.canvasWidth / 2;
        const centerY = state.canvasHeight / 2;
        const worldX = (centerX - state.offsetX) / state.scale;
        const worldY = (centerY - state.offsetY) / state.scale;
        state.scale = Math.max(state.scale / mapCfg.zoomStep, mapCfg.minZoom);
        state.offsetX = centerX - worldX * state.scale;
        state.offsetY = centerY - worldY * state.scale;
        if (elements.zoomInfo) elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
        draw();
        saveViewport();
        scheduleReload();
    });

    document.getElementById('centerBtn').addEventListener('click', centerOnAgent);
}