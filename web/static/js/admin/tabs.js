// web/static/js/admin/tabs.js
import { loadPlanetStats } from './stats.js';

export function initTabs() {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            this.classList.add('active');
            const tabId = this.dataset.tab;
            document.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
            document.getElementById(tabId).classList.add('active');
            if (tabId === 'tab-stats') {
                loadPlanetStats();
            }
        });
    });
}