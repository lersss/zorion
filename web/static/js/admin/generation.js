// web/static/js/admin/generation.js
import { fetchWithAuth } from './auth.js';
import { loadStats } from './stats.js';
import { loadWorlds } from './worlds.js';
import { pollJob } from './poll.js';

let pollIntervals = {};

export async function generateUniverse() {
    // ... (код из старого admin.js, без изменений)
}

export async function generatePlanets() {
    // ... (код из старого admin.js)
}

export async function generateFactions() {
    // ... (код из старого admin.js)
}

export async function generateResources() {
    // ... (код из старого admin.js)
}

export async function cancelGeneration(jobType) {
    // ... (код из старого admin.js)
}

export async function clearUniverse() {
    // ... (код из старого admin.js)
}