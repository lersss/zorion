import { openSystemModal } from './modal/index.js';
import { state, elements } from './map/config.js';
import { resizeCanvas } from './map/render.js';
import { loadData } from './map/data.js';
import { handleCanvasClick, initFlyBtn, initPanZoom } from './map/events.js';
import { animationLoop } from './map/animation.js';
import { initHover } from './map/events.js';

function init() {
    elements.canvas.addEventListener('click', handleCanvasClick);
    window.addEventListener('resize', resizeCanvas);
    initFlyBtn();
    initPanZoom();
    initHover();
    
    loadData().then(() => {
        animationLoop();
    });
}

init();