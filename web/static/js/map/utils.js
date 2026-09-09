import { state } from './config.js';
import { CONFIG } from '../config.js';

export function isFiniteNumber(v) {
    return typeof v === 'number' && isFinite(v);
}

export function worldToCanvas(world) {
    return {
        x: world.coord_x * state.scale + state.offsetX,
        y: world.coord_y * state.scale + state.offsetY
    };
}

export function getStarColor(spectralClass) {
    const colors = CONFIG.map.starColors;
    return colors[spectralClass] || colors.default;
}

// для обратной совместимости (пока не удалим старый код)
export function getColorByType(type) {
    return CONFIG.map.starColors.default;
}