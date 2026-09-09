// web/static/js/modal/index.js
import { modalState, resetState } from './state.js';
import { drawSystem } from './render.js';
import { initEvents } from './events.js';
import { clearTextureCache } from './textures.js';
import { getStarColor, getStarSize } from './utils.js';

export function openSystemModal(worldId, worldName, spectralClass) {
    const token = localStorage.getItem('token');
    if (!token) {
        alert('Не авторизован. Пожалуйста, войдите в систему.');
        return;
    }

    fetch(`/api/worlds/${worldId}/planets`, {
        headers: { 'Authorization': 'Bearer ' + token }
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
    if (document.getElementById('system-modal-overlay')) return;

    resetState();

    const starColor = getStarColor(spectralClass);
    const starRadius = getStarSize(spectralClass);

    // Создаём оверлей
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

    // Модальное окно
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

    // Заголовок
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

    // Контент: Canvas + таблица
    const content = document.createElement('div');
    content.style.cssText = `
        display: flex;
        flex: 1;
        gap: 20px;
        min-height: 0;
    `;

    // Canvas
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

    // Таблица
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

    // Закрытие по клику на оверлей
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeModal();
    });

    // Размеры
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

    clearTextureCache();

    // Отрисовка
    requestAnimationFrame(() => {
        drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
    });

    // Инициализация событий
    initEvents(canvas, spectralClass, planets, starRadius, starColor, width, height);

    // Resize observer
    const resizeObserver = new ResizeObserver(() => {
        const newRect = canvasWrapper.getBoundingClientRect();
        modalState.canvasWidth = newRect.width;
        modalState.canvasHeight = newRect.height;
        drawSystem(canvas, spectralClass, planets, starRadius, starColor, newRect.width, newRect.height);
    });
    resizeObserver.observe(canvasWrapper);
}

function closeModal() {
    const overlay = document.getElementById('system-modal-overlay');
    if (overlay) overlay.remove();
    resetState();
}

// Добавляем стиль анимации
if (!document.getElementById('modal-fade-style')) {
    const style = document.createElement('style');
    style.id = 'modal-fade-style';
    style.textContent = `
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
    `;
    document.head.appendChild(style);
}

window.openSystemModal = openSystemModal;