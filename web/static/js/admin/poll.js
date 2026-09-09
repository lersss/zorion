// web/static/js/admin/poll.js
import { fetchWithAuth } from './auth.js';
import { loadStats } from './stats.js';
import { loadWorlds } from './worlds.js';

let pollIntervals = {};

export async function pollJob(jobType, progressId, resultId, cancelBtnId) {
    // ... (код из старого admin.js)
}