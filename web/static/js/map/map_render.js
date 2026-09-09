// web/static/js/map/map_render.js
import { state, elements } from './config.js';
import { isFiniteNumber, worldToCanvas, getStarColor } from './utils.js';
import { CONFIG } from '../config.js';

const { map: mapCfg } = CONFIG;

export function resizeCanvas() {
    const wrapper = document.getElementById('canvas-wrapper');
    state.canvasWidth = wrapper.clientWidth;
    state.canvasHeight = wrapper.clientHeight;
    elements.canvas.width = state.canvasWidth;
    elements.canvas.height = state.canvasHeight;
    draw();
}

export function draw() {
    const { canvasWidth, canvasHeight, currentWorldId, isFlying, flyStartTime, flyDuration, flyFrom, flyTo, scale, offsetX, offsetY, filteredWorlds, worlds } = state;
    const { ctx, statusBar } = elements;

    const worldsToDraw = filteredWorlds || worlds || [];

    ctx.clearRect(0, 0, canvasWidth, canvasHeight);

    // Сетка
    ctx.strokeStyle = '#1e293b';
    ctx.lineWidth = 0.5;
    const gridStep = mapCfg.gridStep * scale;
    if (gridStep > mapCfg.gridDisplayThreshold) {
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

    // Миры
    worldsToDraw.forEach(w => {
        const pos = worldToCanvas(w);
        if (!isFiniteNumber(pos.x) || !isFiniteNumber(pos.y)) return;

        const spectralSizeMap = {
            'O': 21, 'B': 19.5, 'A': 18,
            'F': 16.5, 'G': 15,
            'K': 12, 'M': 9,
            'L': 7.5, 'T': 6,
            'Y': 5.4
        };
        let baseSize = spectralSizeMap[w.spectral_class] || 12;
        let hash = 0;
        for (let i = 0; i < w.id.length; i++) {
            hash = (hash * 31 + w.id.charCodeAt(i)) & 0xFFFFFFFF;
        }
        const variation = 0.9 + (hash % 20) / 100;
        const radius = Math.max(mapCfg.minRadius, baseSize * variation * scale);
        const color = getStarColor(w.spectral_class || 'G');

        const opacity = w.opacity !== undefined ? w.opacity : 1;
        ctx.globalAlpha = opacity;

        ctx.beginPath();
        ctx.arc(pos.x, pos.y, radius, 0, Math.PI * 2);
        ctx.fillStyle = color;
        ctx.fill();
        ctx.strokeStyle = '#0f172a';
        ctx.lineWidth = 1;
        ctx.stroke();

        if (w.id === currentWorldId) {
            try {
                const glow = ctx.createRadialGradient(pos.x, pos.y, Math.max(0, radius - 2), pos.x, pos.y, radius + 10);
                glow.addColorStop(0, 'rgba(251,191,36,0.3)');
                glow.addColorStop(1, 'rgba(251,191,36,0)');
                ctx.fillStyle = glow;
                ctx.beginPath();
                ctx.arc(pos.x, pos.y, radius + 10, 0, Math.PI * 2);
                ctx.fill();
            } catch (e) {
                ctx.beginPath();
                ctx.arc(pos.x, pos.y, radius + 6, 0, Math.PI * 2);
                ctx.strokeStyle = '#fbbf24';
                ctx.lineWidth = 2;
                ctx.stroke();
            }
        }

        if (w.id === state.hoveredWorldId) {
            ctx.save();
            ctx.shadowColor = 'rgba(255,255,255,0.3)';
            ctx.shadowBlur = 12;
            ctx.beginPath();
            ctx.arc(pos.x, pos.y, radius + 3, 0, Math.PI * 2);
            ctx.fillStyle = 'rgba(255,255,255,0.15)';
            ctx.fill();
            ctx.strokeStyle = 'rgba(255,255,255,0.6)';
            ctx.lineWidth = 2;
            ctx.stroke();
            ctx.restore();
        }

        ctx.globalAlpha = 1;
    });

    if (scale > mapCfg.nameDisplayThreshold) {
        worldsToDraw.forEach(w => {
            const pos = worldToCanvas(w);
            if (!isFiniteNumber(pos.x) || !isFiniteNumber(pos.y)) return;
            const spectralSizeMap = {
                'O': 21, 'B': 19.5, 'A': 18,
                'F': 16.5, 'G': 15,
                'K': 12, 'M': 9,
                'L': 7.5, 'T': 6,
                'Y': 5.4
            };
            let baseSize = spectralSizeMap[w.spectral_class] || 12;
            let hash = 0;
            for (let i = 0; i < w.id.length; i++) {
                hash = (hash * 31 + w.id.charCodeAt(i)) & 0xFFFFFFFF;
            }
            const variation = 0.9 + (hash % 20) / 100;
            const radius = Math.max(mapCfg.minRadius, baseSize * variation * scale);
            const opacity = w.opacity !== undefined ? w.opacity : 1;
            ctx.globalAlpha = opacity;
            ctx.fillStyle = '#94a3b8';
            ctx.font = `${Math.max(8, mapCfg.nameFontSize * scale)}px system-ui`;
            ctx.textAlign = 'center';
            ctx.fillText(w.name, pos.x, pos.y + radius + mapCfg.nameFontSize * scale);
            ctx.globalAlpha = 1;
        });
    }

    if (isFlying && flyFrom && flyTo) {
        const elapsed = (Date.now() - flyStartTime) / 1000;
        const progress = Math.min(elapsed / flyDuration, 1);
        const fromPos = worldToCanvas(flyFrom);
        const toPos = worldToCanvas(flyTo);
        if (isFiniteNumber(fromPos.x) && isFiniteNumber(fromPos.y) &&
            isFiniteNumber(toPos.x) && isFiniteNumber(toPos.y)) {
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
    }

    if (isFlying) {
        const elapsed = (Date.now() - flyStartTime) / 1000;
        const remaining = Math.max(0, flyDuration - elapsed);
        statusBar.textContent = `🚀 В полёте... осталось ${Math.ceil(remaining)} сек.`;
    } else {
        const total = worlds ? worlds.length : 0;
        const shown = worldsToDraw.length;
        if (shown === total) {
            statusBar.textContent = `${shown} миров на карте`;
        } else {
            statusBar.textContent = `Показано ${shown} из ${total} миров`;
        }
    }
}