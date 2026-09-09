// minimap.js
import { modalState } from './state.js';

export function drawMiniMap(ctx, cx, cy, maxRadius, starRadius, planets, width, height) {
    const miniSize = 120;
    const miniX = width - miniSize - 20;
    const miniY = height - miniSize - 20;

    ctx.save();
    ctx.fillStyle = 'rgba(0,0,0,0.6)';
    ctx.strokeStyle = '#444';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(miniX, miniY, miniSize, miniSize, 8);
    ctx.fill();
    ctx.stroke();
    ctx.restore();

    let systemRadius = starRadius * 1.8;
    if (planets && planets.length > 0) {
        const maxOrbitIndex = planets.reduce((max, p) => Math.max(max, p.orbit_index), 0);
        const availableRadius = maxRadius - starRadius * 1.8;
        const step = (maxOrbitIndex > 0) ? (availableRadius / (maxOrbitIndex + 1)) : availableRadius / 3;
        const maxOrbitRadius = starRadius * 1.8 + (maxOrbitIndex + 1) * step * 1.1;
        systemRadius = Math.max(systemRadius, maxOrbitRadius);
    }

    const padding = 0.9;
    const miniScale = (miniSize * padding) / (systemRadius * 2);
    const centerX = miniX + miniSize / 2;
    const centerY = miniY + miniSize / 2;

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

    const viewScale = miniScale;
    const viewWidth = (width / modalState.zoom) * viewScale;
    const viewHeight = (height / modalState.zoom) * viewScale;
    const viewCenterX = centerX - (modalState.offsetX / modalState.zoom) * viewScale;
    const viewCenterY = centerY - (modalState.offsetY / modalState.zoom) * viewScale;
    const viewX = viewCenterX - viewWidth / 2;
    const viewY = viewCenterY - viewHeight / 2;

    ctx.save();
    ctx.strokeStyle = 'rgba(255,255,255,0.5)';
    ctx.lineWidth = 1;
    ctx.setLineDash([2, 3]);
    ctx.strokeRect(viewX, viewY, viewWidth, viewHeight);
    ctx.restore();
}

// roundRect polyfill (если не определён)
if (!CanvasRenderingContext2D.prototype.roundRect) {
    CanvasRenderingContext2D.prototype.roundRect = function (x, y, w, h, r) {
        if (w < 2 * r) r = w / 2;
        if (h < 2 * r) r = h / 2;
        this.moveTo(x + r, y);
        this.lineTo(x + w - r, y);
        this.quadraticCurveTo(x + w, y, x + w, y + r);
        this.lineTo(x + w, y + h - r);
        this.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
        this.lineTo(x + r, y + h);
        this.quadraticCurveTo(x, y + h, x, y + h - r);
        this.lineTo(x, y + r);
        this.quadraticCurveTo(x, y, x + r, y);
        return this;
    };
}