// web/static/js/modal/textures.js
import PlanetGenerator from '../planet_generator.js';
import { hashStringToNumber, getClimateId } from './utils.js';

let planetGen = null;
let climateData = null;
const textureCache = new Map();

export async function initTextureGenerator() {
    try {
        const response = await fetch('/static/js/climates.json');
        if (!response.ok) throw new Error('Failed to load climates.json');
        climateData = await response.json();
        planetGen = new PlanetGenerator({
            canvasSize: 64,
            maxCacheSize: 5000,
            climateData: climateData
        });
        console.log('✅ PlanetGenerator initialized');
    } catch (error) {
        console.error('❌ Failed to load climate data:', error);
        planetGen = new PlanetGenerator({
            canvasSize: 64,
            maxCacheSize: 5000,
            climateData: null
        });
    }
}

export function getPlanetTexture(planet, spectralClass, sizeMultiplier) {
    if (!planetGen) return null;
    const key = `planet_${planet.id || planet.orbit_index}_${spectralClass}`;
    if (textureCache.has(key)) return textureCache.get(key);

    const climateId = getClimateId(planet);
    const textureRadius = Math.min(30, 15 + (planet.size || 10) * 1.2);
    const seed = planet.id ? hashStringToNumber(planet.id) : Date.now() + planet.orbit_index;

    try {
        const result = planetGen.generate({
            radius: textureRadius,
            starType: spectralClass,
            climateId: climateId,
            seed: seed
        });
        textureCache.set(key, result.image);
        return result.image;
    } catch (e) {
        console.warn('Planet texture generation failed:', e);
        return null;
    }
}

export function clearTextureCache() {
    textureCache.clear();
}