// web/static/js/modal/state.js
export const modalState = {
    zoom: 1,
    offsetX: 0,
    offsetY: 0,
    isDragging: false,
    dragStartX: 0,
    dragStartY: 0,
    dragStartOffsetX: 0,
    dragStartOffsetY: 0,
    hoveredObject: null,
    selectedPlanetIndex: null, // <-- добавили
    canvasWidth: 0,
    canvasHeight: 0,
    starRadius: 60,
    starColor: '#fff4a3',
    planets: [],
    canvas: null,
    canvasWrapper: null,
    spectralClass: 'G'
};

export function resetState() {
    modalState.zoom = 1;
    modalState.offsetX = 0;
    modalState.offsetY = 0;
    modalState.isDragging = false;
    modalState.hoveredObject = null;
    modalState.selectedPlanetIndex = null; // <-- сбрасываем
}