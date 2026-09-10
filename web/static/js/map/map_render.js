// web/static/js/map/map_render.js
import { state, elements } from './config.js';
import { isFiniteNumber, worldToCanvas, getStarColor } from './utils.js';
import { CONFIG } from '../config.js';

const { map: mapCfg } = CONFIG;

// Размеры звёзд по спектральному классу — вынесено из циклов.
export const SPECTRAL_SIZE = {
    'O': 21, 'B': 19.5, 'A': 18,
    'F': 16.5, 'G': 15,
    'K': 12, 'M': 9,
    'L': 7.5, 'T': 6,
    'Y': 5.4,
};

// ==================== RESIZE ====================

export function resizeCanvas() {
    const wrapper = document.getElementById('canvas-wrapper');
    state.canvasWidth = wrapper.clientWidth;
    state.canvasHeight = wrapper.clientHeight;
    elements.canvas.width = state.canvasWidth;
    elements.canvas.height = state.canvasHeight;
    draw();
}

// ==================== РАЗМЕР КЛАСТЕРА НА ЭКРАНЕ ====================

// Используется и в рендере, и в hover-детекции (из events.js).
export function clusterScreenRadius(c) {
    if (c.cnt === 1) {
        const baseSize = SPECTRAL_SIZE[c.sspec] || 12;
        const hash = hashString(c.sid || '');
        const variation = 0.9 + (hash % 20) / 100;
        return Math.max(mapCfg.minRadius, baseSize * variation * state.scale);
    }
    return Math.max(10, Math.min(30, 8 + Math.log2(c.cnt) * 3));
}

// ==================== DRAW ====================

export function draw() {
    const {
        canvasWidth, canvasHeight, currentWorldId, hoveredWorldId,
        isFlying, flyStartTime, flyDuration, flyFrom, flyTo,
        scale, offsetX, offsetY, clusters
    } = state;
    const { ctx, statusBar } = elements;

    ctx.clearRect(0, 0, canvasWidth, canvasHeight);

    drawGrid(ctx, canvasWidth, canvasHeight, scale, offsetX, offsetY);

    // --- Отрисовка кластеров ---
    const visibleClusters = clusters || [];
    const singles = [];

    for (const c of visibleClusters) {
        const px = c.x * scale + offsetX;
        const py = c.y * scale + offsetY;

        // Пропускаем то, что вне экрана
        if (!isFiniteNumber(px) || !isFiniteNumber(py)) continue;
        if (px < -50 || py < -50 || px > canvasWidth + 50 || py > canvasHeight + 50) continue;

        if (c.cnt === 1) {
            drawSingleStar(ctx, c, px, py, scale, currentWorldId, hoveredWorldId);
            singles.push({ c, x: px, y: py });
        } else {
            drawCluster(ctx, c, px, py);
        }
    }

    // --- Подписи одиночных миров при достаточном зуме ---
    if (scale > mapCfg.nameDisplayThreshold) {
        drawNames(ctx, singles, scale);
    }

    // --- Анимация полёта ---
    if (isFlying && flyFrom && flyTo) {
        drawFlight(ctx, scale, flyFrom, flyTo, flyStartTime, flyDuration);
    }

    updateStatusBar(statusBar, visibleClusters, isFlying, flyStartTime, flyDuration);
}

// ==================== СЕТКА ====================

function drawGrid(ctx, canvasWidth, canvasHeight, scale, offsetX, offsetY) {
    ctx.strokeStyle = '#1e293b';
    ctx.lineWidth = 0.5;
    const gridStep = mapCfg.gridStep * scale;
    if (gridStep <= mapCfg.gridDisplayThreshold) return;

    for (let x = -100; x <= 100; x += mapCfg.gridStep) {
        const px = x * scale + offsetX;
        if (!isFiniteNumber(px)) continue;
        ctx.beginPath();
        ctx.moveTo(px, 0);
        ctx.lineTo(px, canvasHeight);
        ctx.stroke();
    }
    for (let y = -100; y <= 100; y += mapCfg.gridStep) {
        const py = y * scale + offsetY;
        if (!isFiniteNumber(py)) continue;
        ctx.beginPath();
        ctx.moveTo(0, py);
        ctx.lineTo(canvasWidth, py);
        ctx.stroke();
    }
}

// ==================== ОТРИСОВКА ЭЛЕМЕНТОВ ====================

function drawSingleStar(ctx, c, x, y, scale, currentWorldId, hoveredWorldId) {
    const radius = clusterScreenRadius(c);
    const color = getStarColor(c.sspec || 'G');

    ctx.globalAlpha = 1;
    ctx.beginPath();
    ctx.arc(x, y, radius, 0, Math.PI * 2);
    ctx.fillStyle = color;
    ctx.fill();
    ctx.strokeStyle = '#0f172a';
    ctx.lineWidth = 1;
    ctx.stroke();

    if (c.sid === currentWorldId) {
        try {
            const glow = ctx.createRadialGradient(x, y, Math.max(0, radius - 2), x, y, radius + 10);
            glow.addColorStop(0, 'rgba(251,191,36,0.3)');
            glow.addColorStop(1, 'rgba(251,191,36,0)');
            ctx.fillStyle = glow;
            ctx.beginPath();
            ctx.arc(x, y, radius + 10, 0, Math.PI * 2);
            ctx.fill();
        } catch (e) {
            ctx.beginPath();
            ctx.arc(x, y, radius + 6, 0, Math.PI * 2);
            ctx.strokeStyle = '#fbbf24';
            ctx.lineWidth = 2;
            ctx.stroke();
        }
    }

    if (c.sid === hoveredWorldId) {
        ctx.save();
        ctx.shadowColor = 'rgba(255,255,255,0.3)';
        ctx.shadowBlur = 12;
        ctx.beginPath();
        ctx.arc(x, y, radius + 3, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(255,255,255,0.15)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(255,255,255,0.6)';
        ctx.lineWidth = 2;
        ctx.stroke();
        ctx.restore();
    }
}

function drawCluster(ctx, c, x, y) {
    const radius = clusterScreenRadius(c);
    const count = c.cnt;

    // Цвет: от голубого (мало) к фиолетовому (много)
    const t = Math.min(1, Math.log10(count) / 3);
    const r = Math.round(74 + t * 100);
    const g = Math.round(158 - t * 80);
    const b = Math.round(255 - t * 40);
    const fillColor = `rgba(${r},${g},${b},0.85)`;
    const strokeColor = `rgba(${r},${g},${b},1)`;

    ctx.beginPath();
    ctx.arc(x, y, radius, 0, Math.PI * 2);
    ctx.fillStyle = fillColor;
    ctx.fill();
    ctx.strokeStyle = strokeColor;
    ctx.lineWidth = 2;
    ctx.stroke();

    ctx.beginPath();
    ctx.arc(x, y, Math.max(1, radius - 3), 0, Math.PI * 2);
    ctx.strokeStyle = 'rgba(255,255,255,0.25)';
    ctx.lineWidth = 1;
    ctx.stroke();

    const label = formatCount(count);
    const fontSize = Math.max(10, Math.min(14, radius));
    ctx.fillStyle = '#0f172a';
    ctx.font = `bold ${fontSize}px system-ui`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(label, x, y);
    ctx.textBaseline = 'alphabetic';
}

function drawNames(ctx, singles, scale) {
    ctx.fillStyle = '#94a3b8';
    ctx.font = `${Math.max(8, mapCfg.nameFontSize * scale)}px system-ui`;
    ctx.textAlign = 'center';

    for (const s of singles) {
        const radius = clusterScreenRadius(s.c);
        ctx.fillText(s.c.sname || '—', s.x, s.y + radius + mapCfg.nameFontSize * scale);
    }
}

// ==================== ПОЛЁТ ====================

function drawFlight(ctx, scale, flyFrom, flyTo, flyStartTime, flyDuration) {
    const elapsed = (Date.now() - flyStartTime) / 1000;
    const progress = Math.min(elapsed / flyDuration, 1);
    const fromPos = worldToCanvas(flyFrom);
    const toPos = worldToCanvas(flyTo);
    if (!isFiniteNumber(fromPos.x) || !isFiniteNumber(fromPos.y)) return;
    if (!isFiniteNumber(toPos.x) || !isFiniteNumber(toPos.y)) return;

    const x = fromPos.x + (toPos.x - fromPos.x) * progress;
    const y = fromPos.y + (toPos.y - fromPos.y) * progress;

    ctx.beginPath();
    ctx.moveTo(fromPos.x, fromPos.y);
    ctx.lineTo(toPos.x, toPos.y);
    ctx.strokeStyle = 'rgba(251,191,36,0.3)';
    ctx.lineWidth = 2;
    ctx.setLineDash([6, 6]);
    ctx.stroke();
    ctx.setLineDash([]);

    ctx.save();
    const angle = Math.atan2(toPos.y - fromPos.y, toPos.x - fromPos.x);
    ctx.translate(x, y);
    ctx.rotate(angle);
    const shipSize = mapCfg.shipSize * scale;
    ctx.beginPath();
    ctx.moveTo(shipSize, 0);
    ctx.lineTo(-shipSize * 0.7, -shipSize * 0.5);
    ctx.lineTo(-shipSize * 0.7, shipSize * 0.5);
    ctx.closePath();
    ctx.fillStyle = '#60a5fa';
    ctx.fill();
    ctx.strokeStyle = '#1e293b';
    ctx.lineWidth = 1;
    ctx.stroke();
    ctx.restore();
}

// ==================== СТАТУС-БАР ====================

function updateStatusBar(statusBar, clusters, isFlying, flyStartTime, flyDuration) {
    if (isFlying) {
        const elapsed = (Date.now() - flyStartTime) / 1000;
        const remaining = Math.max(0, flyDuration - elapsed);
        statusBar.textContent = `🚀 В полёте... осталось ${Math.ceil(remaining)} сек.`;
        return;
    }

    const totalClusters = clusters.length;
    let totalWorlds = 0;
    let singles = 0;
    for (const c of clusters) {
        totalWorlds += c.cnt;
        if (c.cnt === 1) singles++;
    }
    statusBar.textContent = `${totalWorlds} миров в кадре (${singles} одиночных, ${totalClusters - singles} кластеров)`;
}

// ==================== УТИЛИТЫ ====================

export function hashString(s) {
    let hash = 0;
    for (let i = 0; i < s.length; i++) {
        hash = (hash * 31 + s.charCodeAt(i)) & 0xFFFFFFFF;
    }
    return hash;
}

function formatCount(n) {
    if (n < 1000) return String(n);
    if (n < 10000) return (n / 1000).toFixed(1) + 'k';
    if (n < 1000000) return Math.round(n / 1000) + 'k';
    return (n / 1000000).toFixed(1) + 'M';
}