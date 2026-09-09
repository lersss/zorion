import { state, elements } from './config.js';
import { draw } from './map_render.js';

export function centerOnAgent() {
    console.log('centerOnAgent called, currentWorldId:', state.currentWorldId);
    if (!state.currentWorldId) {
        console.warn('currentWorldId is null or undefined');
        if (state.worlds && state.worlds.length > 0) {
            state.currentWorldId = state.worlds[0].id;
            elements.currentWorldNameEl.textContent = state.worlds[0].name;
            console.log('Auto-assigned first world locally:', state.currentWorldId);
        } else {
            alert('Нет миров для центрирования.');
            return;
        }
    }
    if (!state.worlds || state.worlds.length === 0) {
        console.warn('worlds is empty');
        alert('Нет миров для центрирования.');
        return;
    }
    const world = state.worlds.find(w => w.id === state.currentWorldId);
    if (!world) {
        console.warn('World with id', state.currentWorldId, 'not found in worlds list');
        alert('Мир не найден в списке. Возможно, данные не загружены.');
        return;
    }
    console.log('Centering on world:', world.name, world.coord_x, world.coord_y);
    const cx = world.coord_x;
    const cy = world.coord_y;
    const targetScale = Math.max(state.scale, 1.0);
    state.offsetX = state.canvasWidth / 2 - cx * targetScale;
    state.offsetY = state.canvasHeight / 2 - cy * targetScale;
    state.scale = targetScale;
    elements.zoomInfo.textContent = Math.round(state.scale * 100) + '%';
    draw();
}