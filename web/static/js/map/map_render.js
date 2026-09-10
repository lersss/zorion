// web/static/js/map/map_render.js
import { state, elements } from './config.js';
import { isFiniteNumber, worldToCanvas, getStarColor } from './utils.js';
import { CONFIG } from '../config.js';

const { map: mapCfg } = CONFIG;

// ==================== КОНСТАНТЫ ====================

// Размер ячейки кластеризации в пикселях экрана.
// Чем больше — тем крупнее кластеры. 40 — компромисс: не слипается, но заметно.
const CLUSTER_CELL = 40;

// Если миров в ячейке <= этому числу, показываем их как отдельные точки
// с индивидуальной подписью (только при достаточном зуме).
const CLUSTER_SINGLE_LIMIT = 1;

// Размеры звёзд по спектральному классу. Вынесено из цикла — было 100000× пересоздание.
const SPECTRAL_SIZE = {
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

// ==================== DRAW ====================

export function draw() {
    const {
        canvasWidth, canvasHeight, currentWorldId, hoveredWorldId,
        isFlying, flyStartTime, flyDuration, flyFrom, flyTo,
        scale, offsetX, offsetY, filteredWorlds, worlds
    } = state;
    const { ctx, statusBar } = elements;

    const worldsToDraw = filteredWorlds || worlds || [];

    ctx.clearRect(0, 0, canvasWidth, canvasHeight);

    drawGrid(ctx, canvasWidth, canvasHeight, scale, offsetX, offsetY);

    // --- ШАГ 1: разбить миры на ячейки ---
    const clusters = buildClusters(worldsToDraw, canvasWidth, canvasHeight, scale, offsetX, offsetY);

    // --- ШАГ 2: нарисовать точки и кластеры ---
    const singles = [];
    for (const cluster of clusters.values()) {
        if (cluster.count === CLUSTER_SINGLE_LIMIT) {
            const w = cluster.worlds[0];
            const pos = cluster;
            drawSingleStar(ctx, w, pos.x, pos.y, scale, currentWorldId, hoveredWorldId);
            singles.push({ world: w, x: pos.x, y: pos.y });
        } else {
            drawCluster(ctx, cluster, scale);
        }
    }

    // --- ШАГ 3: подписи для одиночных миров при достаточном зуме ---
    if (scale > mapCfg.nameDisplayThreshold) {
        drawNames(ctx, singles, scale);
    }

    // --- ШАГ 4: анимация полёта ---
    if (isFlying && flyFrom && flyTo) {
        drawFlight(ctx, scale, flyFrom, flyTo, flyStartTime, flyDuration);
    }

    // --- Статус-бар ---
    updateStatusBar(statusBar, worlds, worldsToDraw, clusters, isFlying, flyStartTime, flyDuration);
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

// ==================== КЛАСТЕРИЗАЦИЯ ====================

// buildClusters — один проход по мирам, раскладывает их по ячейкам сетки.
// Возвращает Map<key, {x, y, count, worlds, avgR, avgG, avgB}>.
function buildClusters(worlds, canvasWidth, canvasHeight, scale, offsetX, offsetY) {
    const clusters = new Map();

    for (let i = 0; i < worlds.length; i++) {
        const w = worlds[i];
        const px = w.coord_x * scale + offsetX;
        const py = w.coord_y * scale + offsetY;

        // За пределами экрана (с запасом) — не считаем.
        if (px < -CLUSTER_CELL || py < -CLUSTER_CELL ||
            px > canvasWidth + CLUSTER_CELL || py > canvasHeight + CLUSTER_CELL) {
            continue;
        }

        const cellX = Math.floor(px / CLUSTER_CELL);
        const cellY = Math.floor(py / CLUSTER_CELL);
        const key = cellX * 100000 + cellY;

        let cluster = clusters.get(key);
        if (!cluster) {
            cluster = {
                x: 0, y: 0, count: 0, worlds: [],
            };
            clusters.set(key, cluster);
        }

        cluster.count++;
        cluster.worlds.push(w);
        cluster.x += px;
        cluster.y += py;
    }

    // Усредняем координаты кластера.
    for (const cluster of clusters.values()) {
        cluster.x /= cluster.count;
        cluster.y /= cluster.count;
    }

    return clusters;
}

// ==================== ОТРИСОВКА ====================

function drawSingleStar(ctx, w, x, y, scale, currentWorldId, hoveredWorldId) {
    const baseSize = SPECTRAL_SIZE[w.spectral_class] || 12;
    const hash = hashString(w.id);
    const variation = 0.9 + (hash % 20) / 100;
    const radius = Math.max(mapCfg.minRadius, baseSize * variation * scale);
    const color = getStarColor(w.spectral_class || 'G');

    ctx.globalAlpha = 1;
    ctx.beginPath();
    ctx.arc(x, y, radius, 0, Math.PI * 2);
    ctx.fillStyle = color;
    ctx.fill();
    ctx.strokeStyle = '#0f172a';
    ctx.lineWidth = 1;
    ctx.stroke();

    // Подсветка текущего мира.
    if (w.id === currentWorldId) {
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

    // Подсветка hover.
    if (w.id === hoveredWorldId) {
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

// drawCluster — рисует кластер: круг с числом.
// Размер зависит от логарифма count — большие кластеры больше, но не гигантские.
function drawCluster(ctx, cluster, scale) {
    const count = cluster.count;
    const radius = Math.max(10, Math.min(30, 8 + Math.log2(count) * 3));

    // Цвет — от синего (мало) к фиолетовому (много).
    const t = Math.min(1, Math.log10(count) / 3);
    const r = Math.round(74 + t * 100);
    const g = Math.round(158 - t * 80);
    const b = Math.round(255 - t * 40);
    const fillColor = `rgba(${r},${g},${b},0.85)`;
    const strokeColor = `rgba(${r},${g},${b},1)`;

    ctx.beginPath();
    ctx.arc(cluster.x, cluster.y, radius, 0, Math.PI * 2);
    ctx.fillStyle = fillColor;
    ctx.fill();
    ctx.strokeStyle = strokeColor;
    ctx.lineWidth = 2;
    ctx.stroke();

    // Внутренний светлый кант.
    ctx.beginPath();
    ctx.arc(cluster.x, cluster.y, radius - 3, 0, Math.PI * 2);
    ctx.strokeStyle = 'rgba(255,255,255,0.25)';
    ctx.lineWidth = 1;
    ctx.stroke();

    // Число внутри.
    const label = formatCount(count);
    const fontSize = Math.max(10, Math.min(14, radius));
    ctx.fillStyle = '#0f172a';
    ctx.font = `bold ${fontSize}px system-ui`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(label, cluster.x, cluster.y);
    ctx.textBaseline = 'alphabetic';
}

function drawNames(ctx, singles, scale) {
    ctx.fillStyle = '#94a3b8';
    ctx.font = `${Math.max(8, mapCfg.nameFontSize * scale)}px system-ui`;
    ctx.textAlign = 'center';

    for (const s of singles) {
        const baseSize = SPECTRAL_SIZE[s.world.spectral_class] || 12;
        const hash = hashString(s.world.id);
        const variation = 0.9 + (hash % 20) / 100;
        const radius = Math.max(mapCfg.minRadius, baseSize * variation * scale);
        ctx.fillText(s.world.name, s.x, s.y + radius + mapCfg.nameFontSize * scale);
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

function updateStatusBar(statusBar, worlds, worldsToDraw, clusters, isFlying, flyStartTime, flyDuration) {
    if (isFlying) {
        const elapsed = (Date.now() - flyStartTime) / 1000;
        const remaining = Math.max(0, flyDuration - elapsed);
        statusBar.textContent = `🚀 В полёте... осталось ${Math.ceil(remaining)} сек.`;
        return;
    }

    const total = worlds ? worlds.length : 0;
    const visible = worldsToDraw.length;
    const shownOnScreen = Array.from(clusters.values()).reduce((sum, c) => sum + c.count, 0);
    const clustersCount = clusters.size;

    if (total === 0) {
        statusBar.textContent = 'Нет миров';
        return;
    }
    if (visible === total && shownOnScreen === total) {
        statusBar.textContent = `${total} миров на карте (${clustersCount} объектов)`;
    } else {
        statusBar.textContent = `${shownOnScreen} из ${total} миров (${clustersCount} объектов)`;
    }
}

// ==================== УТИЛИТЫ ====================

function hashString(s) {
    let hash = 0;
    for (let i = 0; i < s.length; i++) {
        hash = (hash * 31 + s.charCodeAt(i)) & 0xFFFFFFFF;
    }
    return hash;
}

// formatCount — 1250 → «1.3k», 12500 → «13k», 125000 → «125k».
function formatCount(n) {
    if (n < 1000) return String(n);
    if (n < 10000) return (n / 1000).toFixed(1) + 'k';
    if (n < 1000000) return Math.round(n / 1000) + 'k';
    return (n / 1000000).toFixed(1) + 'M';
}