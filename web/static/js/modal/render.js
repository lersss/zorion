// web/static/js/modal/render.js
import { modalState } from './state.js';
import { drawMiniMap } from './minimap.js';
import { getPlanetTexture } from './textures.js';

export async function drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height) {
    const dpr = window.devicePixelRatio || 1;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    canvas.style.width = width + 'px';
    canvas.style.height = height + 'px';

    const ctx = canvas.getContext('2d');
    ctx.scale(dpr, dpr);

    const cx = width / 2;
    const cy = height / 2;
    const maxRadius = Math.min(width, height) * 0.4;

    const maxStarRadius = maxRadius * 0.25;
    const finalStarRadius = Math.min(starRadius, maxStarRadius);

    const planetCount = planets ? planets.length : 0;
    let sizeMultiplier = 1;
    let orbitSpacingMultiplier = 1;
    if (planetCount > 8 && planetCount <= 12) {
        sizeMultiplier = 0.85;
        orbitSpacingMultiplier = 0.9;
    } else if (planetCount > 12) {
        sizeMultiplier = 0.7;
        orbitSpacingMultiplier = 0.8;
    }

    const maxOrbit = planets ? planets.reduce((max, p) => Math.max(max, p.orbit_index), 0) : 0;
    const availableRadius = maxRadius - finalStarRadius * 1.8;
    const step = (maxOrbit > 0) ? (availableRadius / (maxOrbit + 1)) * orbitSpacingMultiplier : availableRadius / 3;

    ctx.save();
    ctx.translate(modalState.offsetX, modalState.offsetY);
    ctx.scale(modalState.zoom, modalState.zoom);

    // ---- СЛОЙ 1: ОРБИТЫ ----
    if (planets && planets.length > 0) {
        planets.forEach((p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = finalStarRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            ctx.save();
            ctx.strokeStyle = '#444';
            ctx.lineWidth = 1;
            ctx.setLineDash([3, 6]);
            ctx.beginPath();
            ctx.arc(cx, cy, orbitRadius, 0, 2 * Math.PI);
            ctx.stroke();
            ctx.restore();
        });
    }

    // ---- СЛОЙ 2: ЗВЕЗДА ----
    ctx.save();
    ctx.shadowColor = starColor;
    ctx.shadowBlur = 40;
    ctx.beginPath();
    ctx.arc(cx, cy, finalStarRadius, 0, 2 * Math.PI);
    ctx.fillStyle = starColor;
    ctx.fill();
    ctx.restore();

    // ---- СЛОЙ 3: ПЛАНЕТЫ (АСИНХРОННАЯ ЗАГРУЗКА) ----
    if (planets && planets.length > 0) {
        // Загружаем все текстуры параллельно
        const loadPromises = planets.map(async (p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = finalStarRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            const angle = (idx * 1.3 + 0.7) % (2 * Math.PI);
            const px = cx + orbitRadius * Math.cos(angle);
            const py = cy + orbitRadius * Math.sin(angle);

            let texture = null;
            try {
                texture = await getPlanetTexture(p, spectralClass, sizeMultiplier);
            } catch (e) {
                console.warn('Failed to load texture for planet', p.id, e);
            }

            const drawRadius = (10 + (p.size || 10) * 0.6) * sizeMultiplier;
            return { x: px, y: py, radius: drawRadius, texture, idx };
        });

        const loaded = await Promise.all(loadPromises);

        // Рисуем все планеты
        loaded.forEach(({ x, y, radius, texture, idx }) => {
            if (texture && texture.complete && texture.naturalWidth > 0) {
                ctx.save();
                ctx.shadowColor = 'rgba(255,255,255,0.1)';
                ctx.shadowBlur = 8;
                ctx.drawImage(texture, x - radius, y - radius, radius * 2, radius * 2);
                ctx.restore();
            } else {
                // fallback круг
                const p = planets[idx];
                let color = '#aaa';
                const type = (p.type || '').toLowerCase();
                if (type.includes('газовый') || type === 'gas_giant') color = '#e8a87c';
                else if (type.includes('землеподобная') || type === 'terran') color = '#6fcf97';
                else if (type.includes('пустынная') || type === 'desert') color = '#d4a373';
                else if (type.includes('ледяная') || type === 'ice') color = '#a8d8ea';
                else if (type.includes('вулканическая') || type === 'volcanic') color = '#e74c3c';
                else if (type.includes('океаническая') || type === 'ocean') color = '#3498db';
                ctx.save();
                ctx.shadowColor = color;
                ctx.shadowBlur = 15;
                ctx.beginPath();
                ctx.arc(x, y, radius, 0, 2 * Math.PI);
                ctx.fillStyle = color;
                ctx.fill();
                ctx.restore();
            }
        });

        // ---- ПОДСВЕТКА ПРИ ХОВЕРЕ (поверх планет) ----
        if (modalState.hoveredObject === 'star') {
            ctx.save();
            ctx.shadowColor = 'rgba(255,255,255,0.3)';
            ctx.shadowBlur = 25;
            ctx.beginPath();
            ctx.arc(cx, cy, finalStarRadius + 4, 0, 2 * Math.PI);
            ctx.fillStyle = 'rgba(255,255,255,0.15)';
            ctx.fill();
            ctx.strokeStyle = 'rgba(255,255,255,0.7)';
            ctx.lineWidth = 2;
            ctx.stroke();
            ctx.restore();
        } else if (modalState.hoveredObject && modalState.hoveredObject.type === 'planet') {
            const idx = modalState.hoveredObject.index;
            const p = loaded[idx];
            if (p) {
                ctx.save();
                ctx.shadowColor = 'rgba(255,255,255,0.3)';
                ctx.shadowBlur = 20;
                ctx.beginPath();
                ctx.arc(p.x, p.y, p.radius + 3, 0, 2 * Math.PI);
                ctx.fillStyle = 'rgba(255,255,255,0.15)';
                ctx.fill();
                ctx.strokeStyle = 'rgba(255,255,255,0.6)';
                ctx.lineWidth = 2;
                ctx.stroke();
                ctx.restore();
            }
        }
    } else {
        ctx.fillStyle = '#666';
        ctx.font = '16px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('Нет планет', cx, cy + 60);
    }

    ctx.restore(); // сброс трансформации

    // ---- МИНИ-КАРТА (поверх всего) ----
    drawMiniMap(ctx, cx, cy, maxRadius, finalStarRadius, planets, width, height);
}