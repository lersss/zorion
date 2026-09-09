// web/static/js/modal.js

// --- Импорт генератора планет ---
import PlanetGenerator from './planet-generator.js';
import climateData from './climates.json' assert { type: 'json' };

// --- Инициализация генератора ---
const planetGen = new PlanetGenerator({
    canvasSize: 64,          // размер текстуры планеты
    maxCacheSize: 5000,
    climateData: climateData
});

// --- Кеш для текстур (чтобы не генерировать повторно) ---
const textureCache = new Map();

// --- Хеш-функция для строки -> число (детерминированно) ---
function hashStringToNumber(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        const char = str.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash |= 0; // Convert to 32bit integer
    }
    return Math.abs(hash);
}

// --- Состояние модалки ---
let modalState = {
    zoom: 1,
    offsetX: 0,
    offsetY: 0,
    isDragging: false,
    dragStartX: 0,
    dragStartY: 0,
    dragStartOffsetX: 0,
    dragStartOffsetY: 0,
    hoveredObject: null
};

// --- Функция определения климата на основе данных планеты ---
function getClimateId(planet) {
    const type = (planet.type || '').toLowerCase();
    const temp = planet.temperature || 0;
    const water = planet.water_percent || 0;

    if (type.includes('вулканическая') || type.includes('лавовая')) return 'extreme';
    if (type.includes('пустынная') && temp > 300) return 'hot';
    if (type.includes('ледяная') || temp < 200) return 'cold';
    if (type.includes('океаническая') || water > 60) return 'temperate';
    if (temp > 350) return 'hot';
    if (temp > 200) return 'temperate';
    if (temp > 100) return 'cold';
    return 'variable';
}

// --- Открытие модалки ---
function openSystemModal(worldId, worldName, spectralClass) {
    const token = localStorage.getItem('token');
    if (!token) {
        alert('Не авторизован. Пожалуйста, войдите в систему.');
        return;
    }

    fetch(`/api/worlds/${worldId}/planets`, {
        headers: {
            'Authorization': 'Bearer ' + token
        }
    })
    .then(response => {
        if (!response.ok) {
            if (response.status === 401 || response.status === 403) {
                alert('Сессия истекла. Пожалуйста, войдите заново.');
                return;
            }
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        console.log('Planets data:', data);
        renderModal(worldName, spectralClass, data.planets);
    })
    .catch(error => {
        console.error('Error loading planets:', error);
        alert('Ошибка загрузки планет: ' + error.message);
    });
}

function renderModal(worldName, spectralClass, planets) {
    if (document.getElementById('system-modal-overlay')) {
        return;
    }

    modalState.zoom = 1;
    modalState.offsetX = 0;
    modalState.offsetY = 0;
    modalState.hoveredObject = null;

    const starColors = {
        'O': '#9bb0ff', 'B': '#aabfff', 'A': '#cad7ff',
        'F': '#f8f7ff', 'G': '#fff4a3', 'K': '#ffd2a1',
        'M': '#ffb47c', 'L': '#ff8c5a', 'T': '#d95c14',
        'Y': '#9e4b2c'
    };
    const starColor = starColors[spectralClass] || '#ffffff';

    const starSizeMap = {
        'O': 120, 'B': 105, 'A': 90,
        'F': 75, 'G': 60,
        'K': 48, 'M': 36,
        'L': 30, 'T': 24,
        'Y': 18
    };
    const starRadius = starSizeMap[spectralClass] || 60;

    const overlay = document.createElement('div');
    overlay.id = 'system-modal-overlay';
    overlay.style.cssText = `
        position: fixed;
        top: 0; left: 0; width: 100%; height: 100%;
        background: rgba(0,0,0,0.7);
        z-index: 1000;
        display: flex;
        justify-content: center;
        align-items: center;
        backdrop-filter: blur(4px);
        animation: fadeIn 0.2s ease;
    `;

    const modal = document.createElement('div');
    modal.style.cssText = `
        background: #1a1a2e;
        color: #e0e0e0;
        border-radius: 16px;
        padding: 20px;
        width: 90%;
        max-width: 1100px;
        height: 85vh;
        max-height: 800px;
        display: flex;
        flex-direction: column;
        box-shadow: 0 8px 32px rgba(0,0,0,0.5);
        position: relative;
    `;

    const header = document.createElement('div');
    header.style.cssText = `
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;
    `;
    const title = document.createElement('h2');
    title.textContent = `${worldName} (${spectralClass})`;
    title.style.cssText = `
        margin: 0;
        font-size: 1.5rem;
        color: ${starColor};
    `;
    const closeBtn = document.createElement('button');
    closeBtn.innerHTML = '✕';
    closeBtn.style.cssText = `
        background: none;
        border: none;
        color: #aaa;
        font-size: 1.8rem;
        cursor: pointer;
        padding: 0 8px;
    `;
    closeBtn.onclick = () => closeModal();
    header.appendChild(title);
    header.appendChild(closeBtn);
    modal.appendChild(header);

    const content = document.createElement('div');
    content.style.cssText = `
        display: flex;
        flex: 1;
        gap: 20px;
        min-height: 0;
    `;

    const canvasWrapper = document.createElement('div');
    canvasWrapper.style.cssText = `
        flex: 2;
        min-width: 0;
        background: #0d0d1a;
        border-radius: 12px;
        position: relative;
        overflow: hidden;
        cursor: grab;
    `;
    const canvas = document.createElement('canvas');
    canvas.id = 'system-canvas';
    canvas.style.cssText = `
        width: 100%;
        height: 100%;
        display: block;
    `;
    canvasWrapper.appendChild(canvas);
    content.appendChild(canvasWrapper);

    const tableWrapper = document.createElement('div');
    tableWrapper.style.cssText = `
        flex: 1;
        background: #12121f;
        border-radius: 12px;
        padding: 12px;
        overflow-y: auto;
        min-width: 200px;
        max-height: 100%;
    `;
    const tableTitle = document.createElement('h3');
    tableTitle.textContent = 'Планеты';
    tableTitle.style.cssText = `margin: 0 0 8px 0; font-size: 1rem; color: #aaa;`;
    tableWrapper.appendChild(tableTitle);

    const table = document.createElement('table');
    table.style.cssText = `
        width: 100%;
        border-collapse: collapse;
        font-size: 0.8rem;
    `;
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>#</th>
            <th>Тип</th>
            <th>Размер</th>
            <th>Температура</th>
        </tr>
    `;
    thead.style.cssText = `text-align: left; color: #888; border-bottom: 1px solid #333;`;
    table.appendChild(thead);

    const tbody = document.createElement('tbody');
    if (planets && planets.length > 0) {
        planets.forEach((p, idx) => {
            const tr = document.createElement('tr');
            tr.style.cssText = `border-bottom: 1px solid #1a1a2e;`;
            tr.innerHTML = `
                <td>${idx + 1}</td>
                <td>${p.type || 'неизвестно'}</td>
                <td>${p.size ? p.size.toFixed(1) : '-'}</td>
                <td>${p.temperature ? p.temperature.toFixed(0) + 'K' : '-'}</td>
            `;
            tbody.appendChild(tr);
        });
    } else {
        const tr = document.createElement('tr');
        tr.innerHTML = `<td colspan="4" style="text-align:center;color:#666;">Нет планет</td>`;
        tbody.appendChild(tr);
    }
    table.appendChild(tbody);
    tableWrapper.appendChild(table);
    content.appendChild(tableWrapper);

    modal.appendChild(content);
    overlay.appendChild(modal);
    document.body.appendChild(overlay);

    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeModal();
    });

    const rect = canvasWrapper.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;

    modalState.canvasWidth = width;
    modalState.canvasHeight = height;
    modalState.starRadius = starRadius;
    modalState.starColor = starColor;
    modalState.planets = planets;
    modalState.canvas = canvas;
    modalState.canvasWrapper = canvasWrapper;
    modalState.spectralClass = spectralClass;

    // Очищаем кеш текстур при открытии новой модалки, чтобы не было конфликтов
    textureCache.clear();

    requestAnimationFrame(() => {
        drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
    });

    // --- Обработчики событий (остаются без изменений) ---
    canvas.addEventListener('mousemove', (e) => {
        const rectCanvas = canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rectCanvas.left) * (canvas.width / rectCanvas.width);
        const mouseY = (e.clientY - rectCanvas.top) * (canvas.height / rectCanvas.height);

        const worldX = (mouseX - modalState.offsetX) / modalState.zoom;
        const worldY = (mouseY - modalState.offsetY) / modalState.zoom;

        const cx = width / 2;
        const cy = height / 2;

        const distToStar = Math.hypot(worldX - cx, worldY - cy);
        const maxStarRadius = Math.min(starRadius, Math.min(width, height) * 0.4 * 0.25);
        if (distToStar < maxStarRadius + 8) {
            if (modalState.hoveredObject !== 'star') {
                modalState.hoveredObject = 'star';
                canvas.style.cursor = 'pointer';
                drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
            }
            return;
        }

        let foundPlanet = null;
        const maxRadius = Math.min(width, height) * 0.4;
        const finalStarRadius = Math.min(starRadius, maxRadius * 0.25);
        const availableRadius = maxRadius - finalStarRadius * 1.8;
        const maxOrbit = planets.reduce((max, p) => Math.max(max, p.orbit_index), 0);
        const step = (maxOrbit > 0) ? (availableRadius / (maxOrbit + 1)) : availableRadius / 3;

        planets.forEach((p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = finalStarRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            const angle = (idx * 1.3 + 0.7) % (2 * Math.PI);
            const px = cx + orbitRadius * Math.cos(angle);
            const py = cy + orbitRadius * Math.sin(angle);

            let planetRadius = 10;
            const type = (p.type || '').toLowerCase();
            if (type.includes('газовый') || type === 'gas_giant') planetRadius = 22;
            else if (type.includes('землеподобная') || type === 'terran' || type === 'earthlike') planetRadius = 12;
            else if (type.includes('пустынная') || type === 'desert') planetRadius = 10;
            else if (type.includes('ледяная') || type === 'ice') planetRadius = 10;
            else if (type.includes('вулканическая') || type === 'volcanic') planetRadius = 10;
            else if (type.includes('океаническая') || type === 'ocean') planetRadius = 12;
            else planetRadius = 8;

            const planetCount = planets.length;
            let sizeMultiplier = 1;
            if (planetCount > 8 && planetCount <= 12) sizeMultiplier = 0.85;
            else if (planetCount > 12) sizeMultiplier = 0.7;
            planetRadius *= sizeMultiplier;

            const dist = Math.hypot(worldX - px, worldY - py);
            if (dist < planetRadius + 6) {
                foundPlanet = idx;
            }
        });

        if (foundPlanet !== null) {
            if (modalState.hoveredObject === null || modalState.hoveredObject.type !== 'planet' || modalState.hoveredObject.index !== foundPlanet) {
                modalState.hoveredObject = { type: 'planet', index: foundPlanet };
                canvas.style.cursor = 'pointer';
                drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
            }
        } else {
            if (modalState.hoveredObject !== null) {
                modalState.hoveredObject = null;
                canvas.style.cursor = modalState.isDragging ? 'grabbing' : 'default';
                drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
            }
        }
    });

    canvas.addEventListener('mouseleave', () => {
        if (modalState.hoveredObject !== null) {
            modalState.hoveredObject = null;
            canvas.style.cursor = modalState.isDragging ? 'grabbing' : 'default';
            drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
        }
    });

    canvas.addEventListener('wheel', (e) => {
        e.preventDefault();
        const rectCanvas = canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rectCanvas.left) * (canvas.width / rectCanvas.width);
        const mouseY = (e.clientY - rectCanvas.top) * (canvas.height / rectCanvas.height);

        const delta = e.deltaY > 0 ? 0.9 : 1.1;
        const newZoom = Math.min(Math.max(modalState.zoom * delta, 0.3), 5);

        const worldX = (mouseX - modalState.offsetX) / modalState.zoom;
        const worldY = (mouseY - modalState.offsetY) / modalState.zoom;
        modalState.zoom = newZoom;
        modalState.offsetX = mouseX - worldX * modalState.zoom;
        modalState.offsetY = mouseY - worldY * modalState.zoom;

        drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
    }, { passive: false });

    canvas.addEventListener('mousedown', (e) => {
        modalState.isDragging = true;
        modalState.dragStartX = e.clientX;
        modalState.dragStartY = e.clientY;
        modalState.dragStartOffsetX = modalState.offsetX;
        modalState.dragStartOffsetY = modalState.offsetY;
        canvas.style.cursor = 'grabbing';
    });

    window.addEventListener('mousemove', (e) => {
        if (modalState.isDragging) {
            const dx = e.clientX - modalState.dragStartX;
            const dy = e.clientY - modalState.dragStartY;
            modalState.offsetX = modalState.dragStartOffsetX + dx;
            modalState.offsetY = modalState.dragStartOffsetY + dy;
            drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
        }
    });

    window.addEventListener('mouseup', () => {
        if (modalState.isDragging) {
            modalState.isDragging = false;
            canvas.style.cursor = modalState.hoveredObject ? 'pointer' : 'default';
        }
    });

    const resizeObserver = new ResizeObserver(() => {
        const newRect = canvasWrapper.getBoundingClientRect();
        modalState.canvasWidth = newRect.width;
        modalState.canvasHeight = newRect.height;
        drawSystem(canvas, spectralClass, planets, starRadius, starColor, newRect.width, newRect.height);
    });
    resizeObserver.observe(canvasWrapper);
}

// --- Отрисовка системы с использованием генератора текстур и seed ---
function drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height) {
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

    // ---- СЛОЙ 3: ПЛАНЕТЫ (с генерацией текстур) ----
    if (planets && planets.length > 0) {
        const planetData = [];
        planets.forEach((p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = finalStarRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            const angle = (idx * 1.3 + 0.7) % (2 * Math.PI);
            const px = cx + orbitRadius * Math.cos(angle);
            const py = cy + orbitRadius * Math.sin(angle);

            // ---- Генерация текстуры планеты с seed на основе ID ----
            const textureKey = `planet_${p.id || idx}_${spectralClass}`;
            let texture = textureCache.get(textureKey);
            if (!texture) {
                const climateId = getClimateId(p);
                const textureRadius = Math.min(30, 15 + (p.size || 10) * 1.2);
                // Используем p.id как seed
                const seed = p.id ? hashStringToNumber(p.id) : (Date.now() + idx);
                try {
                    const result = planetGen.generate({
                        radius: textureRadius,
                        starType: spectralClass,
                        climateId: climateId,
                        seed: seed  // <-- детерминизм!
                    });
                    texture = result.image;
                    textureCache.set(textureKey, texture);
                } catch (e) {
                    console.warn('Planet texture generation failed:', e);
                    texture = null;
                }
            }

            // ---- Рисуем текстуру или запасной круг ----
            if (texture) {
                const drawRadius = (10 + (p.size || 10) * 0.6) * sizeMultiplier;
                ctx.save();
                ctx.shadowColor = 'rgba(255,255,255,0.1)';
                ctx.shadowBlur = 8;
                ctx.drawImage(texture, px - drawRadius, py - drawRadius, drawRadius * 2, drawRadius * 2);
                ctx.restore();
            } else {
                // Запасной вариант: цветной круг
                let color = '#aaa';
                const type = (p.type || '').toLowerCase();
                if (type.includes('газовый') || type === 'gas_giant') color = '#e8a87c';
                else if (type.includes('землеподобная') || type === 'terran') color = '#6fcf97';
                else if (type.includes('пустынная') || type === 'desert') color = '#d4a373';
                else if (type.includes('ледяная') || type === 'ice') color = '#a8d8ea';
                else if (type.includes('вулканическая') || type === 'volcanic') color = '#e74c3c';
                else if (type.includes('океаническая') || type === 'ocean') color = '#3498db';
                const drawRadius = (10 + (p.size || 10) * 0.6) * sizeMultiplier;
                ctx.save();
                ctx.shadowColor = color;
                ctx.shadowBlur = 15;
                ctx.beginPath();
                ctx.arc(px, py, drawRadius, 0, 2 * Math.PI);
                ctx.fillStyle = color;
                ctx.fill();
                ctx.restore();
            }

            planetData.push({ x: px, y: py, radius: (10 + (p.size || 10) * 0.6) * sizeMultiplier, color: '#aaa', idx });
        });

        // ---- СЛОЙ 4: ПОДСВЕТКА (hover) ----
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
            const p = planetData[idx];
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

    // ---- МИНИ-КАРТА ----
    drawMiniMap(ctx, cx, cy, maxRadius, finalStarRadius, planets, width, height);
}

// --- МИНИ-КАРТА (без изменений) ---
function drawMiniMap(ctx, cx, cy, maxRadius, starRadius, planets, width, height) {
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

function closeModal() {
    const overlay = document.getElementById('system-modal-overlay');
    if (overlay) overlay.remove();
    modalState.zoom = 1;
    modalState.offsetX = 0;
    modalState.offsetY = 0;
    modalState.hoveredObject = null;
}

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

const style = document.createElement('style');
style.textContent = `
    @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
    }
`;
document.head.appendChild(style);

window.openSystemModal = openSystemModal;