// web/static/js/modal/events.js
import { modalState } from './state.js';
import { drawSystem } from './modal_render.js';

export function initEvents(canvas, spectralClass, planets, starRadius, starColor, width, height) {
    // Получаем DPR (устройство пикселей на CSS-пиксель)
    const dpr = window.devicePixelRatio || 1;

    // ---- HOVER ----
    canvas.addEventListener('mousemove', (e) => {
        const rect = canvas.getBoundingClientRect();
        // Преобразуем координаты мыши в логические (без учёта dpr)
        // canvas.width уже = width * dpr, поэтому mouseX в физических пикселях
        // делим на dpr, чтобы получить логические
        const mouseX = (e.clientX - rect.left) * (canvas.width / rect.width) / dpr;
        const mouseY = (e.clientY - rect.top) * (canvas.height / rect.height) / dpr;

        const worldX = (mouseX - modalState.offsetX) / modalState.zoom;
        const worldY = (mouseY - modalState.offsetY) / modalState.zoom;

        const cx = width / 2;
        const cy = height / 2;

        // Проверка попадания в звезду
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

        // ---- Вычисляем параметры, как в drawSystem ----
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

        const maxRadius = Math.min(width, height) * 0.4;
        const finalStarRadius = Math.min(starRadius, maxRadius * 0.25);
        const availableRadius = maxRadius - finalStarRadius * 1.8;
        const maxOrbit = planets.reduce((max, p) => Math.max(max, p.orbit_index), 0);
        const step = (maxOrbit > 0) ? (availableRadius / (maxOrbit + 1)) * orbitSpacingMultiplier : availableRadius / 3;

        let foundPlanet = null;
        planets.forEach((p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = finalStarRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            const angle = (idx * 1.3 + 0.7) % (2 * Math.PI);
            const px = cx + orbitRadius * Math.cos(angle);
            const py = cy + orbitRadius * Math.sin(angle);

            let planetRadius = 10 * sizeMultiplier;
            const type = (p.type || '').toLowerCase();
            if (type.includes('газовый') || type === 'gas_giant') planetRadius = 22 * sizeMultiplier;
            else if (type.includes('землеподобная') || type === 'terran' || type === 'earthlike') planetRadius = 12 * sizeMultiplier;
            else if (type.includes('пустынная') || type === 'desert') planetRadius = 10 * sizeMultiplier;
            else if (type.includes('ледяная') || type === 'ice') planetRadius = 10 * sizeMultiplier;
            else if (type.includes('вулканическая') || type === 'volcanic') planetRadius = 10 * sizeMultiplier;
            else if (type.includes('океаническая') || type === 'ocean') planetRadius = 12 * sizeMultiplier;
            else planetRadius = 8 * sizeMultiplier;

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

    // ---- WHEEL ZOOM ----
    canvas.addEventListener('wheel', (e) => {
        e.preventDefault();
        const rect = canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rect.left) * (canvas.width / rect.width) / dpr;
        const mouseY = (e.clientY - rect.top) * (canvas.height / rect.height) / dpr;

        const delta = e.deltaY > 0 ? 0.9 : 1.1;
        const newZoom = Math.min(Math.max(modalState.zoom * delta, 0.3), 5);

        const worldX = (mouseX - modalState.offsetX) / modalState.zoom;
        const worldY = (mouseY - modalState.offsetY) / modalState.zoom;
        modalState.zoom = newZoom;
        modalState.offsetX = mouseX - worldX * modalState.zoom;
        modalState.offsetY = mouseY - worldY * modalState.zoom;

        drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
    }, { passive: false });

    // ---- DRAG ----
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
            const dx = (e.clientX - modalState.dragStartX) / dpr;
            const dy = (e.clientY - modalState.dragStartY) / dpr;
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

    // ---- CLICK ----
    canvas.addEventListener('click', (e) => {
        const rect = canvas.getBoundingClientRect();
        const mouseX = (e.clientX - rect.left) * (canvas.width / rect.width) / dpr;
        const mouseY = (e.clientY - rect.top) * (canvas.height / rect.height) / dpr;

        const worldX = (mouseX - modalState.offsetX) / modalState.zoom;
        const worldY = (mouseY - modalState.offsetY) / modalState.zoom;

        const cx = width / 2;
        const cy = height / 2;
        const maxRadius = Math.min(width, height) * 0.4;
        const finalStarRadius = Math.min(starRadius, maxRadius * 0.25);

        // Проверка клика по звезде — сбрасываем выбор
        const distToStar = Math.hypot(worldX - cx, worldY - cy);
        if (distToStar < finalStarRadius + 8) {
            if (modalState.selectedPlanetIndex !== null) {
                modalState.selectedPlanetIndex = null;
                drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
                if (typeof window.updateRightPanel === 'function') {
                    window.updateRightPanel(null);
                }
            }
            return;
        }

        // Проверка клика по планетам
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

        const availableRadius = maxRadius - finalStarRadius * 1.8;
        const maxOrbit = planets.reduce((max, p) => Math.max(max, p.orbit_index), 0);
        const step = (maxOrbit > 0) ? (availableRadius / (maxOrbit + 1)) * orbitSpacingMultiplier : availableRadius / 3;

        let foundPlanet = null;
        planets.forEach((p, idx) => {
            const randomOffset = (idx * 1.7) % 0.2 - 0.1;
            const orbitRadius = finalStarRadius * 1.8 + (p.orbit_index + 1) * step * (1 + randomOffset);
            const angle = (idx * 1.3 + 0.7) % (2 * Math.PI);
            const px = cx + orbitRadius * Math.cos(angle);
            const py = cy + orbitRadius * Math.sin(angle);

            let planetRadius = 10 * sizeMultiplier;
            const type = (p.type || '').toLowerCase();
            if (type.includes('газовый') || type === 'gas_giant') planetRadius = 22 * sizeMultiplier;
            else if (type.includes('землеподобная') || type === 'terran' || type === 'earthlike') planetRadius = 12 * sizeMultiplier;
            else if (type.includes('пустынная') || type === 'desert') planetRadius = 10 * sizeMultiplier;
            else if (type.includes('ледяная') || type === 'ice') planetRadius = 10 * sizeMultiplier;
            else if (type.includes('вулканическая') || type === 'volcanic') planetRadius = 10 * sizeMultiplier;
            else if (type.includes('океаническая') || type === 'ocean') planetRadius = 12 * sizeMultiplier;
            else planetRadius = 8 * sizeMultiplier;

            const dist = Math.hypot(worldX - px, worldY - py);
            if (dist < planetRadius + 6) {
                foundPlanet = idx;
            }
        });

        if (foundPlanet !== null) {
            if (modalState.selectedPlanetIndex !== foundPlanet) {
                modalState.selectedPlanetIndex = foundPlanet;
                drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
                if (typeof window.updateRightPanel === 'function') {
                    window.updateRightPanel(foundPlanet);
                }
            }
        } else {
            // Клик по пустому месту — сбрасываем выбор
            if (modalState.selectedPlanetIndex !== null) {
                modalState.selectedPlanetIndex = null;
                drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
                if (typeof window.updateRightPanel === 'function') {
                    window.updateRightPanel(null);
                }
            }
        }
    });
}