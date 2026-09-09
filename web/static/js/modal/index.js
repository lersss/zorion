// index.js
import { modalState, resetState } from './state.js';
import { drawSystem } from './render.js';
import { initEvents } from './events.js';
import { initTextureGenerator, clearTextureCache } from './textures.js';
import { getStarColor, getStarSize } from './utils.js';

// Инициализируем генератор текстур при загрузке модуля
await initTextureGenerator();

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

    // Создаём оверлей и модалку (код идентичен предыдущему, я не буду дублировать его здесь полностью)
    // Вставьте сюда полный код создания оверлея, модалки, заголовка, кнопки закрытия, таблицы и т.д.
    // Я приведу полный файл в конце, чтобы не перегружать это сообщение.
    // Но структура: оверлей -> модалка -> заголовок + кнопка закрытия -> canvasWrapper + tableWrapper.

    // ... (весь код рендеринга модалки, который был в renderModal, перенесите сюда)

    // После создания canvas, инициализируем события
    const canvas = document.getElementById('system-canvas');
    const canvasWrapper = canvas.parentElement;
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

    requestAnimationFrame(() => {
        drawSystem(canvas, spectralClass, planets, starRadius, starColor, width, height);
    });

    initEvents(canvas, spectralClass, planets, starRadius, starColor, width, height);

    // Resize observer (если нужно)
    const resizeObserver = new ResizeObserver(() => {
        const newRect = canvasWrapper.getBoundingClientRect();
        modalState.canvasWidth = newRect.width;
        modalState.canvasHeight = newRect.height;
        drawSystem(canvas, spectralClass, planets, starRadius, starColor, newRect.width, newRect.height);
    });
    resizeObserver.observe(canvasWrapper);

    // Закрытие по клику на оверлей
    const overlay = document.getElementById('system-modal-overlay');
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeModal();
    });
}

function closeModal() {
    const overlay = document.getElementById('system-modal-overlay');
    if (overlay) overlay.remove();
    resetState();
}

// Экспортируем функцию для глобального использования
window.openSystemModal = openSystemModal;