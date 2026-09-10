// web/static/js/modal/panel.js
import { modalState } from './state.js';
import { drawSystem } from './modal_render.js';
import { renderTabContent } from './tabs.js';

// Перевод Кельвинов в Цельсии (для таблицы планет)
function kelvinToCelsius(k) {
    if (typeof k !== 'number' || isNaN(k)) return '—';
    return (k - 273.15).toFixed(1);
}

// renderRightPanel — рисует правую панель модалки:
// список планет (selectedIndex === null) или карточку планеты.
export function renderRightPanel(planets, selectedIndex) {
    const panel = document.getElementById('right-panel');
    if (!panel) return;

    if (selectedIndex === null || selectedIndex === undefined) {
        renderList(panel, planets);
    } else {
        renderCard(panel, planets, selectedIndex);
    }
}

// ---------- СПИСОК ПЛАНЕТ ----------

function renderList(panel, planets) {
    panel.innerHTML = `
        <h3 style="margin: 0 0 8px 0; font-size: 1rem; color: #aaa;">Планеты</h3>
        <table style="width:100%; border-collapse: collapse; font-size: 0.8rem;">
            <thead>
                <tr>
                    <th style="text-align:left; color:#888; border-bottom:1px solid #333;">#</th>
                    <th style="text-align:left; color:#888; border-bottom:1px solid #333;">Тип</th>
                    <th style="text-align:left; color:#888; border-bottom:1px solid #333;">Размер</th>
                    <th style="text-align:left; color:#888; border-bottom:1px solid #333;">T</th>
                </tr>
            </thead>
            <tbody id="planet-list-body"></tbody>
        </table>
    `;

    const tbody = panel.querySelector('#planet-list-body');
    if (!planets || planets.length === 0) {
        tbody.innerHTML = `<tr><td colspan="4" style="text-align:center;color:#666;">Нет планет</td></tr>`;
        return;
    }

    planets.forEach((p, idx) => {
        const tr = document.createElement('tr');
        tr.dataset.index = idx;
        tr.style.cssText = `border-bottom: 1px solid #1a1a2e; cursor: pointer;`;
        tr.innerHTML = `
            <td>${idx + 1}</td>
            <td>${p.type || 'неизвестно'}</td>
            <td>${p.size ? p.size.toFixed(2) : '-'}</td>
            <td>${p.temperature ? kelvinToCelsius(p.temperature) + ' °C' : '-'}</td>
        `;
        tr.addEventListener('mouseenter', () => { tr.style.background = '#1f1f3a'; });
        tr.addEventListener('mouseleave', () => { tr.style.background = 'transparent'; });
        tbody.appendChild(tr);
    });
}

// ---------- КАРТОЧКА ПЛАНЕТЫ ----------

function renderCard(panel, planets, selectedIndex) {
    const planet = planets[selectedIndex];
    if (!planet) {
        renderList(panel, planets);
        return;
    }

    panel.innerHTML = `
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
            <h3 style="margin: 0; font-size: 1rem; color: #aaa;">Планета #${selectedIndex + 1}</h3>
            <button id="back-to-list-btn" style="background: #2a2a4a; border: none; color: #aaa; padding: 4px 12px; border-radius: 4px; cursor: pointer;">← Назад</button>
        </div>
        <div style="display: flex; gap: 8px; margin-bottom: 12px; border-bottom: 1px solid #333; padding-bottom: 8px;">
            <button class="tab-btn" data-tab="general" style="background: none; border: none; color: #888; padding: 4px 12px; cursor: pointer; font-size: 0.8rem; border-radius: 4px;">Общее</button>
            <button class="tab-btn" data-tab="resources" style="background: none; border: none; color: #888; padding: 4px 12px; cursor: pointer; font-size: 0.8rem; border-radius: 4px;">Ресурсы</button>
            <button class="tab-btn" data-tab="settlements" style="background: none; border: none; color: #888; padding: 4px 12px; cursor: pointer; font-size: 0.8rem; border-radius: 4px;">Поселения</button>
            <button class="tab-btn" data-tab="factions" style="background: none; border: none; color: #888; padding: 4px 12px; cursor: pointer; font-size: 0.8rem; border-radius: 4px;">Фракции</button>
        </div>
        <div id="tab-content" style="font-size: 0.85rem; line-height: 1.6;"></div>
    `;

    const tabBtns = panel.querySelectorAll('.tab-btn');
    const tabContent = panel.querySelector('#tab-content');

    function switchTab(tab) {
        tabBtns.forEach(btn => {
            btn.style.color = btn.dataset.tab === tab ? '#fff' : '#888';
            btn.style.background = btn.dataset.tab === tab ? '#2a2a4a' : 'none';
        });
        renderTabContent(tab, planet, tabContent);
    }

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => switchTab(btn.dataset.tab));
    });

    switchTab('general');

    const backBtn = panel.querySelector('#back-to-list-btn');
    if (backBtn) {
        backBtn.addEventListener('click', () => {
            modalState.selectedPlanetIndex = null;
            renderRightPanel(planets, null);
            const canvas = document.getElementById('system-canvas');
            if (canvas) {
                drawSystem(
                    canvas,
                    modalState.spectralClass,
                    planets,
                    modalState.starRadius,
                    modalState.starColor,
                    modalState.canvasWidth,
                    modalState.canvasHeight
                );
            }
        });
    }
}