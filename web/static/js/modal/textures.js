// web/static/js/modal/textures.js

const textureCache = new Map();

export function getPlanetTexture(planet, spectralClass, sizeMultiplier) {
    const key = `planet_${planet.id || planet.orbit_index}_${spectralClass}`;
    if (textureCache.has(key)) return textureCache.get(key);

    // Параметры для запроса
    const seed = planet.id ? hashStringToNumber(planet.id) : Date.now() + planet.orbit_index;
    const climateId = getClimateId(planet);
    const radius = Math.min(30, 15 + (planet.size || 10) * 1.2);

    // Запрашиваем изображение с сервера
    const url = `/api/planet-image?seed=${seed}&starType=${spectralClass}&climateId=${climateId}&radius=${radius}`;

    // Используем fetch и создаём Image
    const img = new Image();
    img.crossOrigin = 'anonymous';
    img.src = url;

    // Кешируем изображение
    textureCache.set(key, img);
    return img;
}

export function clearTextureCache() {
    textureCache.clear();
}

// Вспомогательные функции (нужно их импортировать или определить здесь)
function hashStringToNumber(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        const char = str.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash |= 0;
    }
    return Math.abs(hash);
}

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