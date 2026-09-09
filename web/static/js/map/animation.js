import { state, elements } from './config.js';
import { draw } from './map_render.js';
import { loadData } from './data.js';

export function animationLoop() {
    if (state.isFlying) {
        const elapsed = (Date.now() - state.flyStartTime) / 1000;
        if (elapsed >= state.flyDuration) {
            state.isFlying = false;
            elements.statusBar.textContent = '✅ Прибытие!';
            loadData();
        } else {
            draw();
        }
    }
    requestAnimationFrame(animationLoop);
}