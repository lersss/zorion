/**
 * planet-generator.js
 * Модуль для генерации процедурных планет с учётом климатических данных.
 * Поддерживает типы: rocky, earth, gas, ice, lava (выбираются на основе климата)
 * Принимает JSON с климатами, содержащий:
 *   - id, name, weight (по спектральным типам OBAFGKM),
 *   - allowed_surfaces, allowed_hydrospheres, allowed_atmospheres, allowed_biospheres,
 *   - temperature_min/max, water_chance, life_chance.
 * 
 * Экспортирует класс PlanetGenerator
 * 
 * Пример:
 *   import { PlanetGenerator } from './planet-generator.js';
 *   import climateData from './climates.json' assert { type: 'json' };
 *   const gen = new PlanetGenerator({ canvasSize: 64, climateData });
 *   const planet = gen.generate({ radius: 40, starType: 'G', climateId: 'temperate' });
 *   ctx.drawImage(planet.image, x, y, 40, 40);
 */

// ----- Вспомогательные функции (внутренние) -----

/**
 * Простой генератор псевдослучайных чисел с seed
 * @param {number} seed - начальное число
 * @returns {function(): number} функция, возвращающая float 0..1
 */
function createRNG(seed) {
    let s = seed || Date.now() & 0x7fffffff;
    return function() {
        s = (s * 9301 + 49297) & 0x7fffffff;
        return s / 0x7fffffff;
    };
}

/**
 * Хеширует объект параметров для ключа кэша
 * @param {object} params - параметры планеты
 * @returns {string} хеш-строка
 */
function hashParams(params) {
    return JSON.stringify(params);
}

/**
 * Парсит CSS-цвет в массив [r,g,b,a] (0-255)
 */
function parseColor(colorStr) {
    const temp = document.createElement('div');
    temp.style.color = colorStr;
    document.body.appendChild(temp);
    const style = getComputedStyle(temp);
    const rgb = style.color.match(/\d+/g);
    document.body.removeChild(temp);
    if (rgb) {
        let r = parseInt(rgb[0]), g = parseInt(rgb[1]), b = parseInt(rgb[2]);
        let a = 1;
        if (rgb.length > 3) a = parseFloat(rgb[3]);
        return [r, g, b, a * 255];
    }
    return [100, 150, 255, 50];
}

/**
 * Наложение сферической тени, блика и атмосферы на ImageData
 * @param {ImageData} imageData - исходное изображение (планета без эффектов)
 * @param {number} size - размер стороны квадрата
 * @param {object} options - параметры эффектов
 * @param {string} options.atmosphereColor - цвет атмосферы (CSS-цвет)
 * @param {number} options.atmosphereIntensity - 0..1
 * @param {number} options.highlightIntensity - 0..1 (яркость блика)
 * @param {string} options.highlightColor - цвет блика (обычно белый)
 * @param {number} options.shadowIntensity - 0..1 (тень)
 */
function applyPostProcessing(imageData, size, options) {
    const data = imageData.data;
    const w = size, h = size;
    const cx = w / 2, cy = h / 2;
    const radius = Math.min(w, h) / 2 - 2;

    const atmColor = parseColor(options.atmosphereColor || 'rgba(100,150,255,0.2)');
    const highlightColor = parseColor(options.highlightColor || 'rgba(255,255,255,0.9)');
    const shadowStrength = Math.min(1, Math.max(0, options.shadowIntensity || 0.6));
    const highlightStrength = Math.min(1, Math.max(0, options.highlightIntensity || 0.3));
    const atmStrength = Math.min(1, Math.max(0, options.atmosphereIntensity || 0.2));

    for (let y = 0; y < h; y++) {
        for (let x = 0; x < w; x++) {
            const dx = x - cx, dy = y - cy;
            const dist = Math.sqrt(dx*dx + dy*dy);
            const normDist = dist / radius;
            const idx = (y * w + x) * 4;

            if (normDist > 1.2) {
                // Вне зоны свечения - прозрачный
                data[idx+3] = 0;
                continue;
            }
            if (normDist > 1) {
                // Атмосферное свечение снаружи
                const outer = (normDist - 1) / 0.2;
                const alpha = atmStrength * (1 - outer) * 0.5;
                data[idx] = (data[idx] * (1 - alpha) + atmColor[0] * alpha) | 0;
                data[idx+1] = (data[idx+1] * (1 - alpha) + atmColor[1] * alpha) | 0;
                data[idx+2] = (data[idx+2] * (1 - alpha) + atmColor[2] * alpha) | 0;
                data[idx+3] = (data[idx+3] * (1 - alpha) + 255 * alpha) | 0;
                continue;
            }

            // ---- Сферическая тень ----
            const lightX = -0.5, lightY = -0.4;
            const len = Math.sqrt(lightX*lightX + lightY*lightY);
            const nx = lightX / len, ny = lightY / len;
            const z = Math.sqrt(Math.max(0, radius*radius - dx*dx - dy*dy));
            const normLen = Math.sqrt(dx*dx + dy*dy + z*z);
            const normDx = dx / normLen, normDy = dy / normLen, normDz = z / normLen;
            let diffuse = normDx * nx + normDy * ny + normDz * 0.2;
            diffuse = Math.max(0, Math.min(1, diffuse));
            const shadow = 1 - shadowStrength * (1 - diffuse);
            data[idx] = (data[idx] * shadow) | 0;
            data[idx+1] = (data[idx+1] * shadow) | 0;
            data[idx+2] = (data[idx+2] * shadow) | 0;

            // ---- Блик ----
            const reflectX = 2 * diffuse * normDx - nx;
            const reflectY = 2 * diffuse * normDy - ny;
            const reflectZ = 2 * diffuse * normDz - 0.2;
            const spec = Math.max(0, reflectZ);
            const specIntensity = Math.pow(spec, 20) * highlightStrength * 2;
            if (specIntensity > 0.01) {
                data[idx] = Math.min(255, data[idx] + (highlightColor[0] - data[idx]) * specIntensity) | 0;
                data[idx+1] = Math.min(255, data[idx+1] + (highlightColor[1] - data[idx+1]) * specIntensity) | 0;
                data[idx+2] = Math.min(255, data[idx+2] + (highlightColor[2] - data[idx+2]) * specIntensity) | 0;
            }

            // ---- Внутренняя атмосфера ----
            if (atmStrength > 0 && normDist > 0.7) {
                const edge = (normDist - 0.7) / 0.3;
                const atmAlpha = atmStrength * edge * 0.8;
                data[idx] = (data[idx] * (1 - atmAlpha) + atmColor[0] * atmAlpha) | 0;
                data[idx+1] = (data[idx+1] * (1 - atmAlpha) + atmColor[1] * atmAlpha) | 0;
                data[idx+2] = (data[idx+2] * (1 - atmAlpha) + atmColor[2] * atmAlpha) | 0;
            }
        }
    }
}

// ----- Генераторы для разных типов планет (базовые, без пост-эффектов) -----

function generateRocky(size, rng, params) {
    const imageData = new ImageData(size, size);
    const data = imageData.data;
    const radius = size / 2 - 2;
    const cx = size / 2, cy = size / 2;
    const baseR = 100 + rng() * 60;
    const baseG = 80 + rng() * 50;
    const baseB = 60 + rng() * 40;
    const craterCount = 3 + Math.floor(rng() * 8);
    const craters = [];
    for (let i = 0; i < craterCount; i++) {
        const angle = rng() * 2 * Math.PI;
        const dist = rng() * radius * 0.8;
        const crx = cx + Math.cos(angle) * dist;
        const cry = cy + Math.sin(angle) * dist;
        const crRadius = 2 + rng() * 6;
        craters.push({ x: crx, y: cry, r: crRadius, depth: 0.3 + rng() * 0.5 });
    }
    for (let y = 0; y < size; y++) {
        for (let x = 0; x < size; x++) {
            const dx = x - cx, dy = y - cy;
            const dist = Math.sqrt(dx*dx + dy*dy);
            if (dist > radius) {
                const idx = (y * size + x) * 4;
                data[idx+3] = 0;
                continue;
            }
            const noise1 = Math.sin(x * 0.2 + y * 0.3) * 10 + Math.cos(x * 0.4 - y * 0.5) * 8;
            const noise2 = Math.sin(x * 0.7 + y * 0.9) * 5;
            let r = baseR + noise1 + noise2;
            let g = baseG + noise1 * 0.8 + noise2 * 0.6;
            let b = baseB + noise1 * 0.5 + noise2 * 0.4;
            for (const cr of craters) {
                const ddx = x - cr.x, ddy = y - cr.y;
                const d = Math.sqrt(ddx*ddx + ddy*ddy);
                if (d < cr.r) {
                    const factor = 1 - d / cr.r;
                    const darken = factor * cr.depth * 30;
                    r -= darken; g -= darken; b -= darken;
                    if (d > cr.r * 0.7) {
                        const rim = (d - cr.r * 0.7) / (cr.r * 0.3) * 10;
                        r += rim; g += rim; b += rim;
                    }
                }
            }
            r = Math.max(0, Math.min(255, r));
            g = Math.max(0, Math.min(255, g));
            b = Math.max(0, Math.min(255, b));
            const idx = (y * size + x) * 4;
            data[idx] = r | 0;
            data[idx+1] = g | 0;
            data[idx+2] = b | 0;
            data[idx+3] = 255;
        }
    }
    return imageData;
}

function generateEarth(size, rng, params) {
    const imageData = new ImageData(size, size);
    const data = imageData.data;
    const radius = size / 2 - 2;
    const cx = size / 2, cy = size / 2;
    const oceanR = 20 + rng() * 30;
    const oceanG = 60 + rng() * 50;
    const oceanB = 120 + rng() * 60;
    const continents = [];
    const numContinents = 3 + Math.floor(rng() * 4);
    for (let i = 0; i < numContinents; i++) {
        const angle = rng() * 2 * Math.PI;
        const dist = rng() * radius * 0.7;
        const cx2 = cx + Math.cos(angle) * dist;
        const cy2 = cy + Math.sin(angle) * dist;
        const rad = 5 + rng() * 15;
        const colorR = 50 + rng() * 80;
        const colorG = 100 + rng() * 70;
        const colorB = 30 + rng() * 50;
        continents.push({ x: cx2, y: cy2, r: rad, color: [colorR, colorG, colorB] });
    }
    const clouds = [];
    const numClouds = 2 + Math.floor(rng() * 5);
    for (let i = 0; i < numClouds; i++) {
        const angle = rng() * 2 * Math.PI;
        const dist = rng() * radius * 0.8;
        const cx2 = cx + Math.cos(angle) * dist;
        const cy2 = cy + Math.sin(angle) * dist;
        const rad = 4 + rng() * 12;
        clouds.push({ x: cx2, y: cy2, r: rad, alpha: 0.1 + rng() * 0.3 });
    }
    for (let y = 0; y < size; y++) {
        for (let x = 0; x < size; x++) {
            const dx = x - cx, dy = y - cy;
            const dist = Math.sqrt(dx*dx + dy*dy);
            if (dist > radius) {
                const idx = (y * size + x) * 4;
                data[idx+3] = 0;
                continue;
            }
            let r = oceanR, g = oceanG, b = oceanB;
            for (const cont of continents) {
                const ddx = x - cont.x, ddy = y - cont.y;
                const d = Math.sqrt(ddx*ddx + ddy*ddy);
                if (d < cont.r) {
                    const factor = 1 - d / cont.r;
                    const blend = factor * factor * 0.9 + 0.1;
                    r = r * (1 - blend) + cont.color[0] * blend;
                    g = g * (1 - blend) + cont.color[1] * blend;
                    b = b * (1 - blend) + cont.color[2] * blend;
                }
            }
            const noise = (Math.sin(x * 0.5 + y * 0.7) * 5 + Math.cos(x * 0.9 - y * 0.3) * 4);
            r += noise; g += noise; b += noise;
            for (const cl of clouds) {
                const ddx = x - cl.x, ddy = y - cl.y;
                const d = Math.sqrt(ddx*ddx + ddy*ddy);
                if (d < cl.r) {
                    const factor = 1 - d / cl.r;
                    const alpha = factor * cl.alpha;
                    r = r * (1 - alpha) + 255 * alpha;
                    g = g * (1 - alpha) + 255 * alpha;
                    b = b * (1 - alpha) + 255 * alpha;
                }
            }
            r = Math.max(0, Math.min(255, r));
            g = Math.max(0, Math.min(255, g));
            b = Math.max(0, Math.min(255, b));
            const idx = (y * size + x) * 4;
            data[idx] = r | 0;
            data[idx+1] = g | 0;
            data[idx+2] = b | 0;
            data[idx+3] = 255;
        }
    }
    return imageData;
}

function generateGas(size, rng, params) {
    const imageData = new ImageData(size, size);
    const data = imageData.data;
    const radius = size / 2 - 2;
    const cx = size / 2, cy = size / 2;
    const numBands = 4 + Math.floor(rng() * 6);
    const bands = [];
    let baseHue = 10 + rng() * 40;
    for (let i = 0; i < numBands; i++) {
        const yPos = -radius + (i / (numBands-1)) * 2 * radius;
        const width = 3 + rng() * 8;
        const color = `hsl(${baseHue + rng()*30 - 15}, 80%, ${40 + rng()*30}%)`;
        const parsed = parseColor(color);
        bands.push({ y: yPos, w: width, color: parsed });
    }
    const spotX = cx + (rng() - 0.5) * radius * 0.8;
    const spotY = cy + (rng() - 0.5) * radius * 0.6;
    const spotR = 3 + rng() * 8;
    const spotColor = parseColor(`hsl(${baseHue + 20}, 90%, 60%)`);
    for (let y = 0; y < size; y++) {
        for (let x = 0; x < size; x++) {
            const dx = x - cx, dy = y - cy;
            const dist = Math.sqrt(dx*dx + dy*dy);
            if (dist > radius) {
                const idx = (y * size + x) * 4;
                data[idx+3] = 0;
                continue;
            }
            let r = 100, g = 80, b = 60;
            let found = false;
            for (const band of bands) {
                const dy2 = y - (cy + band.y);
                if (Math.abs(dy2) < band.w/2) {
                    const factor = 1 - Math.abs(dy2) / (band.w/2);
                    const mix = factor * 0.8 + 0.2;
                    r = r * (1 - mix) + band.color[0] * mix;
                    g = g * (1 - mix) + band.color[1] * mix;
                    b = b * (1 - mix) + band.color[2] * mix;
                    found = true;
                }
            }
            if (!found) { r = 40; g = 30; b = 20; }
            const dSpot = Math.sqrt((x - spotX)*(x - spotX) + (y - spotY)*(y - spotY));
            if (dSpot < spotR) {
                const factor = 1 - dSpot / spotR;
                r = r * (1 - factor) + spotColor[0] * factor;
                g = g * (1 - factor) + spotColor[1] * factor;
                b = b * (1 - factor) + spotColor[2] * factor;
            }
            r = Math.max(0, Math.min(255, r));
            g = Math.max(0, Math.min(255, g));
            b = Math.max(0, Math.min(255, b));
            const idx = (y * size + x) * 4;
            data[idx] = r | 0;
            data[idx+1] = g | 0;
            data[idx+2] = b | 0;
            data[idx+3] = 255;
        }
    }
    return imageData;
}

function generateIce(size, rng, params) {
    const imageData = new ImageData(size, size);
    const data = imageData.data;
    const radius = size / 2 - 2;
    const cx = size / 2, cy = size / 2;
    const baseR = 180 + rng() * 40;
    const baseG = 200 + rng() * 40;
    const baseB = 220 + rng() * 30;
    const crackCount = 5 + Math.floor(rng() * 10);
    const cracks = [];
    for (let i = 0; i < crackCount; i++) {
        const angle = rng() * 2 * Math.PI;
        const dist = rng() * radius * 0.8;
        const x1 = cx + Math.cos(angle) * dist;
        const y1 = cy + Math.sin(angle) * dist;
        const angle2 = angle + (rng() - 0.5) * 1.2;
        const dist2 = dist + (rng() - 0.5) * radius * 0.5;
        const x2 = cx + Math.cos(angle2) * dist2;
        const y2 = cy + Math.sin(angle2) * dist2;
        cracks.push({ x1, y1, x2, y2, width: 1 + rng() * 2 });
    }
    for (let y = 0; y < size; y++) {
        for (let x = 0; x < size; x++) {
            const dx = x - cx, dy = y - cy;
            const dist = Math.sqrt(dx*dx + dy*dy);
            if (dist > radius) {
                const idx = (y * size + x) * 4;
                data[idx+3] = 0;
                continue;
            }
            let r = baseR, g = baseG, b = baseB;
            const noise = (Math.sin(x * 0.3 + y * 0.5) * 8 + Math.cos(x * 0.7 - y * 0.2) * 6);
            r += noise; g += noise; b += noise;
            for (const cr of cracks) {
                const dx1 = x - cr.x1, dy1 = y - cr.y1;
                const dx2 = cr.x2 - cr.x1, dy2 = cr.y2 - cr.y1;
                const len2 = dx2*dx2 + dy2*dy2;
                if (len2 === 0) continue;
                let t = (dx1*dx2 + dy1*dy2) / len2;
                t = Math.max(0, Math.min(1, t));
                const projX = cr.x1 + t * dx2;
                const projY = cr.y1 + t * dy2;
                const d = Math.sqrt((x - projX)*(x - projX) + (y - projY)*(y - projY));
                if (d < cr.width) {
                    const factor = 1 - d / cr.width;
                    const darken = factor * 30;
                    r -= darken; g -= darken; b -= darken;
                }
            }
            r = Math.max(0, Math.min(255, r));
            g = Math.max(0, Math.min(255, g));
            b = Math.max(0, Math.min(255, b));
            const idx = (y * size + x) * 4;
            data[idx] = r | 0;
            data[idx+1] = g | 0;
            data[idx+2] = b | 0;
            data[idx+3] = 255;
        }
    }
    return imageData;
}

function generateLava(size, rng, params) {
    const imageData = new ImageData(size, size);
    const data = imageData.data;
    const radius = size / 2 - 2;
    const cx = size / 2, cy = size / 2;
    const baseR = 20 + rng() * 20;
    const baseG = 10 + rng() * 10;
    const baseB = 10 + rng() * 10;
    const lavaCount = 4 + Math.floor(rng() * 6);
    const lavas = [];
    for (let i = 0; i < lavaCount; i++) {
        const startAngle = rng() * 2 * Math.PI;
        const startDist = rng() * radius * 0.9;
        const x1 = cx + Math.cos(startAngle) * startDist;
        const y1 = cy + Math.sin(startAngle) * startDist;
        const points = [{x: x1, y: y1}];
        let curX = x1, curY = y1;
        for (let j = 0; j < 4 + Math.floor(rng()*3); j++) {
            const angle = Math.atan2(curY - cy, curX - cx) + (rng() - 0.5) * 1.5;
            const dist = Math.sqrt((curX-cx)*(curX-cx) + (curY-cy)*(curY-cy));
            const newDist = dist + (rng() - 0.5) * radius * 0.3;
            const newX = cx + Math.cos(angle) * newDist;
            const newY = cy + Math.sin(angle) * newDist;
            if (Math.sqrt((newX-cx)*(newX-cx) + (newY-cy)*(newY-cy)) > radius) break;
            points.push({x: newX, y: newY});
            curX = newX; curY = newY;
        }
        lavas.push(points);
    }
    for (let y = 0; y < size; y++) {
        for (let x = 0; x < size; x++) {
            const dx = x - cx, dy = y - cy;
            const dist = Math.sqrt(dx*dx + dy*dy);
            if (dist > radius) {
                const idx = (y * size + x) * 4;
                data[idx+3] = 0;
                continue;
            }
            let r = baseR, g = baseG, b = baseB;
            let lavaGlow = 0;
            for (const pts of lavas) {
                for (let i = 0; i < pts.length - 1; i++) {
                    const p1 = pts[i], p2 = pts[i+1];
                    const dx1 = x - p1.x, dy1 = y - p1.y;
                    const dx2 = p2.x - p1.x, dy2 = p2.y - p1.y;
                    const len2 = dx2*dx2 + dy2*dy2;
                    if (len2 === 0) continue;
                    let t = (dx1*dx2 + dy1*dy2) / len2;
                    t = Math.max(0, Math.min(1, t));
                    const projX = p1.x + t * dx2;
                    const projY = p1.y + t * dy2;
                    const d = Math.sqrt((x - projX)*(x - projX) + (y - projY)*(y - projY));
                    if (d < 2.5) {
                        const factor = 1 - d / 2.5;
                        lavaGlow = Math.max(lavaGlow, factor);
                    }
                }
            }
            if (lavaGlow > 0) {
                const lavaR = 255, lavaG = 150 + rng()*50, lavaB = 20 + rng()*30;
                const mix = lavaGlow * 0.9 + 0.1;
                r = r * (1 - mix) + lavaR * mix;
                g = g * (1 - mix) + lavaG * mix;
                b = b * (1 - mix) + lavaB * mix;
            }
            const noise = (Math.sin(x * 0.5 + y * 0.6) * 5 + Math.cos(x * 0.8 - y * 0.4) * 4);
            r += noise; g += noise; b += noise;
            r = Math.max(0, Math.min(255, r));
            g = Math.max(0, Math.min(255, g));
            b = Math.max(0, Math.min(255, b));
            const idx = (y * size + x) * 4;
            data[idx] = r | 0;
            data[idx+1] = g | 0;
            data[idx+2] = b | 0;
            data[idx+3] = 255;
        }
    }
    return imageData;
}

// ----- Основной класс PlanetGenerator с поддержкой климатов -----

export class PlanetGenerator {
    /**
     * @param {object} options
     * @param {number} options.canvasSize - внутренний размер квадрата для генерации (по умолчанию 64)
     * @param {number} options.maxCacheSize - максимальное количество кэшированных планет (0 = без кэша)
     * @param {boolean} options.enableCache - включить кэширование (по умолчанию true)
     * @param {object} options.climateData - объект с климатами (как в вашем JSON)
     */
    constructor(options = {}) {
        this.canvasSize = options.canvasSize || 64;
        this.maxCacheSize = options.maxCacheSize || 2000;
        this.enableCache = (options.enableCache !== undefined) ? options.enableCache : true;
        this.cache = new Map();
        this.cacheOrder = [];

        // Сохраняем климатические данные
        this.climateData = options.climateData || null;
        if (this.climateData && this.climateData.climates) {
            this.climates = this.climateData.climates;
        } else {
            this.climates = [];
        }

        // Генераторы типов (визуальные)
        this.generators = {
            rocky: generateRocky,
            earth: generateEarth,
            gas: generateGas,
            ice: generateIce,
            lava: generateLava,
        };
    }

    /**
     * Выбирает климат на основе спектрального типа звезды и весов.
     * @param {string} starType - 'O','B','A','F','G','K','M' (если не указан, выбирается случайный)
     * @param {Function} rng - функция случайности
     * @returns {object} объект климата
     */
    _selectClimateByStarType(starType, rng) {
        if (!this.climates.length) {
            // Если нет данных, возвращаем фиктивный климат
            return {
                id: 'default',
                name: 'Стандартный',
                allowed_surfaces: ['скалистая'],
                allowed_hydrospheres: ['сухая'],
                allowed_atmospheres: ['разряженная'],
                allowed_biospheres: ['стерильная'],
                temperature_min: 200,
                temperature_max: 400,
                water_chance: 0,
                life_chance: 0
            };
        }

        // Если starType не задан, выбираем случайный спектральный класс с равной вероятностью
        const starTypes = ['O','B','A','F','G','K','M'];
        let type = starType;
        if (!type || !starTypes.includes(type)) {
            type = starTypes[Math.floor(rng() * starTypes.length)];
        }

        // Собираем веса для каждого климата для данного типа
        const weighted = [];
        for (const climate of this.climates) {
            const weight = climate.weight && climate.weight[type] ? climate.weight[type] : 0;
            if (weight > 0) {
                weighted.push({ climate, weight });
            }
        }
        if (weighted.length === 0) {
            // Если весов нет, берём первый климат
            return this.climates[0];
        }

        // Выбираем случайный климат с учётом весов
        const totalWeight = weighted.reduce((sum, w) => sum + w.weight, 0);
        let rand = rng() * totalWeight;
        for (const item of weighted) {
            rand -= item.weight;
            if (rand <= 0) {
                return item.climate;
            }
        }
        return weighted[weighted.length-1].climate;
    }

    /**
     * Выбирает случайный элемент из массива
     */
    _pickRandom(arr, rng) {
        if (!arr || arr.length === 0) return null;
        return arr[Math.floor(rng() * arr.length)];
    }

    /**
     * Определяет визуальный тип планеты на основе климатических характеристик.
     * @param {object} climate - объект климата
     * @param {string} surface - выбранная поверхность
     * @param {string} hydrosphere - выбранная гидросфера
     * @param {number} temperature - средняя температура (можно взять среднюю между min и max)
     * @param {Function} rng
     * @returns {string} один из: 'rocky', 'earth', 'gas', 'ice', 'lava'
     */
    _determineVisualType(climate, surface, hydrosphere, temperature, rng) {
        // Приоритет: лавовая, ледяная, водная, иначе скалистая
        if (surface && surface.includes('лавовая')) {
            return 'lava';
        }
        if (surface && surface.includes('ледяная')) {
            return 'ice';
        }
        if (temperature < 200) {
            return 'ice';
        }
        // Если есть вода (океаны, озёра) и не сухая
        if (hydrosphere && (hydrosphere.includes('океаны') || hydrosphere.includes('озёра')) && !hydrosphere.includes('сухая')) {
            return 'earth';
        }
        // Если атмосфера плотная и есть признаки жизни -> earth, иначе rocky
        // Также можно добавить газовые гиганты, но в данных их нет, поэтому игнорируем.
        return 'rocky';
    }

    /**
     * Генерирует планету с учётом климатических данных.
     * @param {object} params
     * @param {number} params.radius - отображаемый радиус (20-50)
     * @param {string} params.starType - спектральный тип звезды (O,B,A,F,G,K,M) – опционально
     * @param {string} params.climateId - конкретный id климата (если указан, starType игнорируется)
     * @param {string} params.surface - конкретная поверхность (если не указана, выбирается случайная из allowed)
     * @param {string} params.hydrosphere - конкретная гидросфера
     * @param {string} params.atmosphere - конкретная атмосфера
     * @param {string} params.biosphere - конкретная биосфера
     * @param {number} params.seed - для детерминизма
     * @param {boolean} params.hasAtmosphere - переопределить наличие атмосферы (если не указано, берётся из климата)
     * @param {boolean} params.hasRings - добавить кольца (случайно для газовых)
     * @returns {object} { image: HTMLCanvasElement, meta: object }
     */
    generate(params = {}) {
        // Создаём RNG
        const seed = params.seed || (Date.now() + Math.random() * 100000) & 0x7fffffff;
        const rng = createRNG(seed);

        // ---- Выбор климата ----
        let climate = null;
        if (params.climateId) {
            // Ищем климат по id
            climate = this.climates.find(c => c.id === params.climateId);
            if (!climate) {
                console.warn(`Климат с id "${params.climateId}" не найден, используется случайный.`);
            }
        }
        if (!climate) {
            // Если климат не задан или не найден, выбираем на основе starType
            const starType = params.starType || null;
            climate = this._selectClimateByStarType(starType, rng);
        }

        // ---- Выбор характеристик из allowed ----
        const surface = params.surface || this._pickRandom(climate.allowed_surfaces, rng) || 'скалистая';
        const hydrosphere = params.hydrosphere || this._pickRandom(climate.allowed_hydrospheres, rng) || 'сухая';
        const atmosphere = params.atmosphere || this._pickRandom(climate.allowed_atmospheres, rng) || 'разряженная';
        const biosphere = params.biosphere || this._pickRandom(climate.allowed_biospheres, rng) || 'стерильная';

        // ---- Вычисление температуры (средняя) ----
        const tempMin = climate.temperature_min || 200;
        const tempMax = climate.temperature_max || 400;
        const temperature = tempMin + rng() * (tempMax - tempMin);

        // ---- Определение визуального типа ----
        let visualType = this._determineVisualType(climate, surface, hydrosphere, temperature, rng);

        // Если мы хотим поддержать газовые гиганты, но их нет в данных, то можно добавить случайно, но мы не будем.

        // ---- Дополнительные параметры ----
        const hasAtmosphere = (params.hasAtmosphere !== undefined) ? params.hasAtmosphere : (atmosphere !== 'разряженная' && atmosphere !== '');
        // Кольца: добавляем для газовых гигантов, но у нас их нет, поэтому только если явно запрошено
        let hasRings = params.hasRings || false;
        if (visualType === 'gas' && rng() > 0.6) hasRings = true;

        // ---- Генерация изображения ----
        const size = this.canvasSize;
        const generator = this.generators[visualType] || this.generators.rocky;
        let imageData = generator(size, rng, { ...params, visualType });

        // ---- Пост-обработка ----
        let atmosphereColor = 'rgba(100,150,255,0.2)';
        let highlightIntensity = 0.3;
        let shadowIntensity = 0.6;
        let atmosphereIntensity = hasAtmosphere ? 0.2 : 0;

        // Настройка под тип
        if (visualType === 'ice') {
            atmosphereColor = 'rgba(200,230,255,0.15)';
            highlightIntensity = 0.8;
            shadowIntensity = 0.4;
        } else if (visualType === 'lava') {
            atmosphereColor = 'rgba(255,100,50,0.25)';
            highlightIntensity = 0.2;
            atmosphereIntensity = hasAtmosphere ? 0.15 : 0;
        } else if (visualType === 'earth') {
            atmosphereColor = 'rgba(70,150,255,0.2)';
            highlightIntensity = 0.3;
        } else if (visualType === 'gas') {
            atmosphereColor = 'rgba(200,180,150,0.15)';
            highlightIntensity = 0.2;
        }

        // Применяем пост-обработку
        applyPostProcessing(imageData, size, {
            atmosphereColor,
            atmosphereIntensity,
            highlightColor: 'rgba(255,255,255,0.9)',
            highlightIntensity,
            shadowIntensity,
        });

        // ---- Создание канваса ----
        const canvas = document.createElement('canvas');
        canvas.width = size;
        canvas.height = size;
        const ctx = canvas.getContext('2d');
        ctx.putImageData(imageData, 0, 0);

        // ---- Кольца (если есть) ----
        if (hasRings) {
            this._drawRings(ctx, size, 'rgba(200,180,150,0.5)', rng);
        }

        // ---- Масштабирование до требуемого радиуса ----
        const displaySize = params.radius ? params.radius * 2 : size;
        const finalCanvas = document.createElement('canvas');
        finalCanvas.width = displaySize;
        finalCanvas.height = displaySize;
        const finalCtx = finalCanvas.getContext('2d');
        finalCtx.imageSmoothingEnabled = true;
        finalCtx.imageSmoothingQuality = 'high';
        finalCtx.drawImage(canvas, 0, 0, displaySize, displaySize);

        // ---- Формирование результата ----
        const result = {
            image: finalCanvas,
            meta: {
                type: visualType,
                radius: params.radius || size/2,
                seed: seed,
                climate: {
                    id: climate.id,
                    name: climate.name,
                    temperature: temperature,
                },
                surface,
                hydrosphere,
                atmosphere,
                biosphere,
                hasAtmosphere,
                hasRings,
                starType: params.starType || null,
            }
        };

        // ---- Кэширование ----
        if (this.enableCache) {
            const cacheKey = hashParams({
                visualType,
                radius: params.radius,
                seed,
                climate: climate.id,
                surface,
                hydrosphere,
                atmosphere,
                biosphere,
                hasAtmosphere,
                hasRings,
            });
            if (this.cache.size >= this.maxCacheSize && this.maxCacheSize > 0) {
                const oldest = this.cacheOrder.shift();
                this.cache.delete(oldest);
            }
            this.cache.set(cacheKey, result);
            this.cacheOrder.push(cacheKey);
        }

        return result;
    }

    /**
     * Рисует кольца на канвасе (поверх планеты)
     */
    _drawRings(ctx, size, color, rng) {
        const cx = size/2, cy = size/2;
        const radius = size/2 - 2;
        const ringOuter = radius * 1.6;
        const ringInner = radius * 1.1;
        const tilt = 0.3 + rng() * 0.4;
        ctx.save();
        ctx.translate(cx, cy);
        ctx.scale(1, tilt);
        ctx.rotate(0.2 + rng() * 0.3);
        const baseColor = parseColor(color);
        const alpha = 0.3 + rng() * 0.3;
        const grad = ctx.createRadialGradient(0, 0, ringInner, 0, 0, ringOuter);
        grad.addColorStop(0, `rgba(${baseColor[0]},${baseColor[1]},${baseColor[2]},0)`);
        grad.addColorStop(0.3, `rgba(${baseColor[0]},${baseColor[1]},${baseColor[2]},${alpha})`);
        grad.addColorStop(0.7, `rgba(${baseColor[0]},${baseColor[1]},${baseColor[2]},${alpha*0.8})`);
        grad.addColorStop(1, `rgba(${baseColor[0]},${baseColor[1]},${baseColor[2]},0)`);
        ctx.beginPath();
        ctx.ellipse(0, 0, ringOuter, ringOuter * 0.25, 0, 0, Math.PI * 2);
        ctx.fillStyle = grad;
        ctx.fill();
        const grad2 = ctx.createRadialGradient(0, 0, ringInner*0.9, 0, 0, ringInner*1.1);
        grad2.addColorStop(0, 'rgba(0,0,0,0)');
        grad2.addColorStop(0.5, `rgba(${baseColor[0]},${baseColor[1]},${baseColor[2]},${alpha*0.3})`);
        grad2.addColorStop(1, 'rgba(0,0,0,0)');
        ctx.beginPath();
        ctx.ellipse(0, 0, ringInner*1.1, ringInner*1.1*0.25, 0, 0, Math.PI * 2);
        ctx.fillStyle = grad2;
        ctx.fill();
        ctx.restore();
    }

    /**
     * Генерирует множество планет (для предварительной загрузки)
     * @param {number} count - количество
     * @param {function} onProgress - колбэк (progress) => void
     * @param {object} baseParams - общие параметры для всех (можно передать starType, climateId и т.д.)
     * @returns {array} массив результатов
     */
    generateMany(count, onProgress = null, baseParams = {}) {
        const results = [];
        for (let i = 0; i < count; i++) {
            const params = {
                ...baseParams,
                seed: (baseParams.seed || 0) + i * 1000 + Math.random() * 1000,
                radius: baseParams.radius || (20 + Math.random() * 30),
            };
            if (!params.climateId && !params.starType) {
                // Если ничего не задано, можно выбрать случайный климат (без привязки к звезде)
                // Но мы можем просто не передавать starType, тогда _selectClimateByStarType выберет случайный тип звезды
            }
            const planet = this.generate(params);
            results.push(planet);
            if (onProgress) {
                onProgress((i+1) / count);
            }
        }
        return results;
    }

    /**
     * Очищает кэш
     */
    clearCache() {
        this.cache.clear();
        this.cacheOrder = [];
    }

    _updateCacheOrder(key) {
        const index = this.cacheOrder.indexOf(key);
        if (index > -1) {
            this.cacheOrder.splice(index, 1);
            this.cacheOrder.push(key);
        }
    }
}

export default PlanetGenerator;