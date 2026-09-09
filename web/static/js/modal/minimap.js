// web/static/js/modal/minimap.js
import { modalState } from './state.js';

export function drawMiniMap(ctx, cx, cy, maxRadius, starRadius, planets, width, height) {
    const miniSize = 120;
    const miniX = width - miniSize - 20;
    const miniY = height - miniSize - 20;

    // Фон
    ctx.save();
    ctx.fillStyle = 'rgba(0,0,0,0.6)';
    ctx.strokeStyle = '#444';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(miniX, miniY, miniSize, miniSize, 8);
    ctx.fill();
    ctx.stroke();
    ctx.restore();

    // ---- Вычисление радиуса системы ----
    let systemRadius = starRadius * 1.8;
    if (planets && planets.length > 0) {
        const maxOrbitIndex = planets.reduce((max, p) => Math.max(max, p.orbit_index), 0);
        const availableRadius = maxRadius - starRadius * 1.8;
        const step = (maxOrbitIndex > 0) ? (availableRadius / (maxOrbitIndex + 1)) : availableRadius / 3;
        const maxOrbitRadius = starRadius * 1.8 + (maxOrbitIndex + 1) * step * 1.1;
        systemRadius = Math.max(systemRadius, maxOrbitRadius);
    }

    // ---- Масштаб мини-карты (чтобы вся система помещалась) ----
    const padding = 0.9;
    const miniScale = (miniSize * padding) / (systemRadius * 2);
    const centerX = miniX + miniSize / 2;
    const centerY = miniY + miniSize / 2;

    // ---- Рисуем звезду и планеты ----
    ctx.save();
    ctx.beginPath();
    ctx.arc(centerX, centerY, Math.max(2, starRadius * miniScale), 0, 2 * Math.PI);
    ctx.fillStyle = '#fff4a3';
    ctx.fill();
    ctx.restore();

    if (planets && planets.length > 0) {
        const maxOrbitIndex = planets.reduce((max, p) => Math.max(max, p.orbit_index), 0);
        const availableRadius = maxRadius - starRadius * 1.8;
        const step = (maxOrbitIndex > 0) ? (availableRadius / (maxOrbitIndex + 1)) : availableRadius / 3;
        planets.forEach((p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = starRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            const angle = (idx * 1.3 + 0.7) % (2 * Math.PI);
            const px = centerX + orbitRadius * Math.cos(angle) * miniScale;
            const py = centerY + orbitRadius * Math.sin(angle) * miniScale;
            ctx.save();
            ctx.beginPath();
            ctx.arc(px, py, Math.max(1.5, 3 * miniScale), 0, 2 * Math.PI);
            ctx.fillStyle = '#6fcf97';
            ctx.fill();
            ctx.restore();
        });
    }

    // ---- РАМКА ВИДИМОЙ ОБЛАСТИ (ОГРАНИЧЕННАЯ) ----
    // Вычисляем размер рамки в координатах мини-карты
    const viewScale = miniScale;
    let viewWidth = (width / modalState.zoom) * viewScale;
    let viewHeight = (height / modalState.zoom) * viewScale;

    // Ограничиваем рамку размерами мини-карты (чтобы не вылезала за границы)
    viewWidth = Math.min(viewWidth, miniSize);
    viewHeight = Math.min(viewHeight, miniSize);

    // Вычисляем центр рамки (с учётом панорамирования)
    let viewCenterX = centerX - (modalState.offsetX / modalState.zoom) * viewScale;
    let viewCenterY = centerY - (modalState.offsetY / modalState.zoom) * viewScale;

    // Ограничиваем центр рамки, чтобы она не выходила за границы мини-карты
    const minX = miniX + viewWidth / 2;
    const maxX = miniX + miniSize - viewWidth / 2;
    const minY = miniY + viewHeight / 2;
    const maxY = miniY + miniSize - viewHeight / 2;
    viewCenterX = Math.max(minX, Math.min(maxX, viewCenterX));
    viewCenterY = Math.max(minY, Math.min(maxY, viewCenterY));

    const viewX = viewCenterX - viewWidth / 2;
    const viewY = viewCenterY - viewHeight / 2;

    // ---- Рисуем рамку ----
    ctx.save();
    ctx.strokeStyle = 'rgba(255,255,255,0.5)';
    ctx.lineWidth = 1;
    ctx.setLineDash([2, 3]);
    ctx.strokeRect(viewX, viewY, viewWidth, viewHeight);
    ctx.restore();
}