import { CONFIG } from '../config.js';

export const state = {
    worlds: [],
    currentWorldId: null,
    selectedWorldId: null,
    isFlying: false,
    flyStartTime: 0,
    flyDuration: 0,
    flyFrom: null,
    flyTo: null,
    scale: 1.0,
    offsetX: 0,
    offsetY: 0,
    isDragging: false,
    dragStartX: 0,
    dragStartY: 0,
    dragStartOffsetX: 0,
    dragStartOffsetY: 0,
    canvasWidth: 0,
    canvasHeight: 0
};

export const elements = {
    canvas: document.getElementById('mapCanvas'),
    get ctx() { return this.canvas.getContext('2d'); },
    tooltip: document.getElementById('tooltip'),
    tooltipName: document.getElementById('tooltipName'),
    tooltipType: document.getElementById('tooltipType'),
    tooltipLevel: document.getElementById('tooltipLevel'),
    tooltipFlyBtn: document.getElementById('tooltipFlyBtn'),
    statusBar: document.getElementById('status-bar'),
    currentWorldNameEl: document.getElementById('currentWorldName'),
    loadingEl: document.getElementById('loading'),
    zoomInfo: document.getElementById('zoom-info')
};