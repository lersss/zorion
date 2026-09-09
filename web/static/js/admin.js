// web/static/js/admin.js

// --- Глобальные ---
let adminPassword = localStorage.getItem('adminPassword') || '';
let currentPage = 1, currentLimit = 50, currentSearch = '';
let pollIntervals = {};

// --- Переключение вкладок ---
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

// --- Авторизация ---
export function setPassword() {
    const pwd = document.getElementById('adminPassword').value;
    if (pwd) {
        adminPassword = pwd;
        localStorage.setItem('adminPassword', pwd);
        alert('Пароль сохранён');
        loadStats();
        loadWorlds(1);
    }
}

export async function fetchWithAuth(url, options = {}) {
    const headers = options.headers || {};
    headers['X-Admin-Password'] = adminPassword;
    const res = await fetch(url, { ...options, headers });
    if (res.status === 401) {
        alert('Неверный пароль администратора');
        throw new Error('Unauthorized');
    }
    return res;
}

// --- Краткая статистика (для шапки) ---
export async function loadStats() {
    try {
        const res = await fetchWithAuth('/admin/stats');
        const data = await res.json();
        document.getElementById('statsWorlds').textContent = data.worlds || 0;
        document.getElementById('statsPlanets').textContent = data.planets || 0;
    } catch (e) { console.error(e); }
}

// --- Детальная статистика планет ---
export async function loadPlanetStats() {
    const container = document.getElementById('planetStatsContainer');
    container.innerHTML = '<p style="color:#94a3b8;">⏳ Загрузка статистики...</p>';
    try {
        const res = await fetchWithAuth('/admin/stats/planets');
        if (!res.ok) throw new Error('Failed to fetch stats');
        const stats = await res.json();
        renderPlanetStats(stats, container);
    } catch (e) {
        container.innerHTML = `<p style="color:#f87171;">❌ Ошибка загрузки: ${e.message}</p>`;
    }
}

export function renderPlanetStats(stats, container) {
    let html = '';

    // Общая сводка
    html += `<div class="stat-grid">`;
    html += `<div class="stat-card"><strong>Всего миров:</strong> ${stats.total_worlds}</div>`;
    html += `<div class="stat-card"><strong>Миров с планетами:</strong> ${stats.worlds_with_planets}</div>`;
    html += `<div class="stat-card"><strong>Всего планет:</strong> ${stats.total_planets}</div>`;
    html += `<div class="stat-card"><strong>Обитаемых:</strong> ${stats.habitable_count}</div>`;
    html += `<div class="stat-card"><strong>С жизнью:</strong> ${stats.life_count}</div>`;
    html += `<div class="stat-card"><strong>Средний размер:</strong> ${stats.avg_size.toFixed(2)}</div>`;
    html += `<div class="stat-card"><strong>Средняя масса:</strong> ${stats.avg_mass.toFixed(2)}</div>`;
    html += `<div class="stat-card"><strong>Средняя температура:</strong> ${stats.avg_temp.toFixed(0)}K</div>`;
    html += `<div class="stat-card"><strong>Средняя вода:</strong> ${stats.avg_water.toFixed(1)}%</div>`;
    html += `<div class="stat-card"><strong>Среднее население:</strong> ${stats.avg_population.toLocaleString()}</div>`;
    html += `</div>`;

    // Типы планет (сортировка по клику)
    html += `<h3 style="margin-top:20px;">Распределение по типам</h3>`;
    html += `<table class="stats-table" id="type-table">
        <thead><tr>
            <th class="sortable" data-sort="type" data-order="asc">Тип</th>
            <th class="sortable" data-sort="count" data-order="asc">Кол-во</th>
        </tr></thead>
        <tbody id="type-body"></tbody>
    </table>`;

    // По спектральным классам (сортировка по клику)
    html += `<h3 style="margin-top:20px;">По спектральным классам</h3>`;
    html += `<table class="stats-table" id="spectral-table">
        <thead><tr>
            <th class="sortable" data-sort="spectral" data-order="asc">Спектр</th>
            <th class="sortable" data-sort="types" data-order="asc">Типы (кол-во)</th>
            <th class="sortable" data-sort="total" data-order="asc">Всего</th>
        </tr></thead>
        <tbody id="spectral-body"></tbody>
    </table>`;

    // Аномалии
    html += `<h3 style="margin-top:20px;">⚠️ Аномалии</h3>`;
    if (stats.anomalies && stats.anomalies.length > 0) {
        html += `<ul>`;
        stats.anomalies.forEach(a => {
            const severityClass = a.severity === 'high' ? 'anomaly-high' : a.severity === 'medium' ? 'anomaly-medium' : 'anomaly-low';
            html += `<li class="${severityClass}">${a.description}: ${a.value.toFixed(1)} (ожидалось ~${a.expected.toFixed(1)})</li>`;
        });
        html += `</ul>`;
    } else {
        html += `<p style="color:#6fcf97;">✅ Аномалий не обнаружено</p>`;
    }

    container.innerHTML = html;

    // Заполняем таблицу типов
    const typeBody = document.getElementById('type-body');
    let typeData = Object.entries(stats.planets_by_type).map(([type, count]) => ({ type, count }));
    typeData.sort((a, b) => b.count - a.count);
    typeBody.innerHTML = typeData.map(d => `<tr><td>${d.type}</td><td>${d.count}</td></tr>`).join('');

    // Заполняем спектральную таблицу
    const spectralBody = document.getElementById('spectral-body');
    let spectralData = Object.entries(stats.planets_by_spectral).map(([spec, types]) => {
        const total = Object.values(types).reduce((sum, v) => sum + v, 0);
        const typesStr = Object.entries(types).map(([t, c]) => `${t}: ${c}`).join(', ');
        return { spec, typesStr, total };
    });
    spectralData.sort((a, b) => b.total - a.total);
    spectralBody.innerHTML = spectralData.map(d => `<tr><td>${d.spec}</td><td>${d.typesStr}</td><td>${d.total}</td></tr>`).join('');

    // Добавляем обработчики кликов для сортировки таблиц
    document.querySelectorAll('#type-table .sortable').forEach(th => {
        th.addEventListener('click', function() {
            const sortKey = this.dataset.sort;
            const currentOrder = this.dataset.order;
            const newOrder = currentOrder === 'asc' ? 'desc' : 'asc';
            this.dataset.order = newOrder;
            const tbody = document.getElementById('type-body');
            const rows = Array.from(tbody.querySelectorAll('tr'));
            rows.sort((a, b) => {
                let valA, valB;
                if (sortKey === 'type') {
                    valA = a.cells[0].textContent;
                    valB = b.cells[0].textContent;
                } else {
                    valA = parseInt(a.cells[1].textContent);
                    valB = parseInt(b.cells[1].textContent);
                }
                if (typeof valA === 'string') {
                    return newOrder === 'asc' ? valA.localeCompare(valB) : valB.localeCompare(valA);
                } else {
                    return newOrder === 'asc' ? valA - valB : valB - valA;
                }
            });
            rows.forEach(row => tbody.appendChild(row));
            document.querySelectorAll('#type-table .sortable').forEach(th => {
                th.textContent = th.textContent.replace(/ [▲▼]/, '');
            });
            this.textContent += newOrder === 'asc' ? ' ▲' : ' ▼';
        });
    });

    document.querySelectorAll('#spectral-table .sortable').forEach(th => {
        th.addEventListener('click', function() {
            const sortKey = this.dataset.sort;
            const currentOrder = this.dataset.order;
            const newOrder = currentOrder === 'asc' ? 'desc' : 'asc';
            this.dataset.order = newOrder;
            const tbody = document.getElementById('spectral-body');
            const rows = Array.from(tbody.querySelectorAll('tr'));
            rows.sort((a, b) => {
                let valA, valB;
                if (sortKey === 'spectral') {
                    valA = a.cells[0].textContent;
                    valB = b.cells[0].textContent;
                } else if (sortKey === 'total') {
                    valA = parseInt(a.cells[2].textContent);
                    valB = parseInt(b.cells[2].textContent);
                } else {
                    valA = parseInt(a.cells[2].textContent);
                    valB = parseInt(b.cells[2].textContent);
                }
                if (typeof valA === 'string') {
                    return newOrder === 'asc' ? valA.localeCompare(valB) : valB.localeCompare(valA);
                } else {
                    return newOrder === 'asc' ? valA - valB : valB - valA;
                }
            });
            rows.forEach(row => tbody.appendChild(row));
            document.querySelectorAll('#spectral-table .sortable').forEach(th => {
                th.textContent = th.textContent.replace(/ [▲▼]/, '');
            });
            this.textContent += newOrder === 'asc' ? ' ▲' : ' ▼';
        });
    });
}

// --- Список миров ---
export async function loadWorlds(page) {
    if (page) currentPage = page;
    currentLimit = parseInt(document.getElementById('limitSelect').value);
    currentSearch = document.getElementById('searchInput').value;

    try {
        const url = `/admin/worlds?page=${currentPage}&limit=${currentLimit}&search=${encodeURIComponent(currentSearch)}`;
        const res = await fetchWithAuth(url);
        const data = await res.json();

        const tbody = document.getElementById('worldsBody');
        tbody.innerHTML = '';
        if (data.data.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;color:#94a3b8;">Нет миров</td></tr>';
        } else {
            data.data.forEach(w => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${w.id}</td>
                    <td>${w.name}</td>
                    <td>${w.coord_x}</td>
                    <td>${w.coord_y}</td>
                    <td><button class="btn-small" onclick="deleteWorld('${w.id}')">Удалить</button></td>
                `;
                tbody.appendChild(row);
            });
        }

        const pagination = document.getElementById('paginationControls');
        pagination.innerHTML = '';
        if (data.total > 0) {
            const totalPages = Math.ceil(data.total / data.limit);
            const prev = document.createElement('button');
            prev.textContent = '◀';
            prev.onclick = () => loadWorlds(Math.max(1, currentPage - 1));
            pagination.appendChild(prev);
            const span = document.createElement('span');
            span.textContent = `Страница ${currentPage} из ${totalPages}`;
            pagination.appendChild(span);
            const next = document.createElement('button');
            next.textContent = '▶';
            next.onclick = () => loadWorlds(Math.min(totalPages, currentPage + 1));
            pagination.appendChild(next);
        }
    } catch (e) {
        console.error(e);
        document.getElementById('worldsBody').innerHTML = '<tr><td colspan="5" style="text-align:center;color:#f87171;">Ошибка загрузки</td></tr>';
    }
}

export async function deleteWorld(id) {
    if (!confirm('Удалить мир?')) return;
    try {
        const res = await fetchWithAuth('/admin/worlds/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id })
        });
        if (res.ok) { alert('Мир удалён'); loadStats(); loadWorlds(currentPage); }
        else { const text = await res.text(); alert('Ошибка: ' + text); }
    } catch (e) { alert('Ошибка: ' + e.message); }
}

export async function createWorld() {
    const name = document.getElementById('worldName').value.trim();
    const x = parseFloat(document.getElementById('coordX').value);
    const y = parseFloat(document.getElementById('coordY').value);
    if (!name || isNaN(x) || isNaN(y)) {
        document.getElementById('createResult').textContent = '❌ Заполните все поля';
        return;
    }
    try {
        const res = await fetchWithAuth('/admin/worlds/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, coord_x: x, coord_y: y })
        });
        if (res.ok) { document.getElementById('createResult').textContent = '✅ Мир создан'; loadStats(); loadWorlds(1); }
        else { const text = await res.text(); document.getElementById('createResult').textContent = '❌ ' + text; }
    } catch (e) { document.getElementById('createResult').textContent = '❌ ' + e.message; }
}

// --- Генерация вселенной ---
export async function generateUniverse() {
    const worlds = parseInt(document.getElementById('genWorlds').value);
    const clusters = parseInt(document.getElementById('genClusters').value);
    const mapSize = parseInt(document.getElementById('genMapSize').value);
    const minDist = parseInt(document.getElementById('genMinDist').value);
    const clusterRadius = parseInt(document.getElementById('genClusterRadius').value);
    const clusterSpacing = parseInt(document.getElementById('genClusterSpacing').value);
    const outlierPercent = parseInt(document.getElementById('genOutlierPercent').value);
    if (isNaN(worlds) || isNaN(clusters) || worlds <= 0 || clusters <= 0 || isNaN(mapSize) || mapSize <= 0 || isNaN(minDist) || minDist <= 0 || isNaN(clusterRadius) || clusterRadius <= 0 || isNaN(clusterSpacing) || clusterSpacing <= 0 || isNaN(outlierPercent) || outlierPercent < 0) {
        document.getElementById('genResult').textContent = '❌ Введите корректные числа';
        return;
    }
    if (!confirm(`Сгенерировать ${worlds} миров в ${clusters} кластерах?`)) return;

    document.getElementById('genResult').textContent = '⏳ Генерация запущена...';
    document.getElementById('genProgress').style.display = 'block';
    document.getElementById('genProgressBar').value = 0;
    document.getElementById('genProgressText').textContent = '0 из ' + worlds;
    document.getElementById('cancelUniverseBtn').style.display = 'inline-block';

    try {
        const res = await fetchWithAuth('/admin/generate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                world_count: worlds,
                cluster_count: clusters,
                map_size: mapSize,
                min_dist: minDist,
                cluster_radius: clusterRadius,
                cluster_spacing: clusterSpacing,
                outlier_percent: outlierPercent
            })
        });
        if (!res.ok) {
            const text = await res.text();
            document.getElementById('genResult').textContent = '❌ Ошибка: ' + text;
            document.getElementById('genProgress').style.display = 'none';
            document.getElementById('cancelUniverseBtn').style.display = 'none';
            return;
        }
        if (pollIntervals['universe']) clearInterval(pollIntervals['universe']);
        pollIntervals['universe'] = setInterval(() => pollJob('generate_universe', 'genProgress', 'genResult', 'cancelUniverseBtn'), 1500);
    } catch (e) {
        document.getElementById('genResult').textContent = '❌ ' + e.message;
        document.getElementById('genProgress').style.display = 'none';
        document.getElementById('cancelUniverseBtn').style.display = 'none';
    }
}

// --- Генерация планет ---
export async function generatePlanets() {
    if (!confirm('Сгенерировать планеты для всех миров?')) return;
    document.getElementById('planetResult').textContent = '⏳ Генерация запущена...';
    document.getElementById('planetProgress').style.display = 'block';
    document.getElementById('planetProgressBar').value = 0;
    document.getElementById('planetProgressText').textContent = 'Подготовка...';
    document.getElementById('cancelPlanetsBtn').style.display = 'inline-block';

    try {
        const res = await fetchWithAuth('/admin/generate-planets', { method: 'POST' });
        if (!res.ok) {
            const text = await res.text();
            document.getElementById('planetResult').textContent = '❌ Ошибка: ' + text;
            document.getElementById('planetProgress').style.display = 'none';
            document.getElementById('cancelPlanetsBtn').style.display = 'none';
            return;
        }
        if (pollIntervals['planets']) clearInterval(pollIntervals['planets']);
        pollIntervals['planets'] = setInterval(() => pollJob('generate_planets', 'planetProgress', 'planetResult', 'cancelPlanetsBtn'), 1500);
    } catch (e) {
        document.getElementById('planetResult').textContent = '❌ ' + e.message;
        document.getElementById('planetProgress').style.display = 'none';
        document.getElementById('cancelPlanetsBtn').style.display = 'none';
    }
}

// --- Генерация фракций ---
export async function generateFactions() {
    if (!confirm('Сгенерировать фракции для всех обитаемых планет?')) return;
    document.getElementById('factionResult').textContent = '⏳ Генерация запущена...';
    document.getElementById('factionProgress').style.display = 'block';
    document.getElementById('factionProgressBar').value = 0;
    document.getElementById('factionProgressText').textContent = 'Подготовка...';
    document.getElementById('cancelFactionsBtn').style.display = 'inline-block';

    try {
        const res = await fetchWithAuth('/admin/generate-factions', { method: 'POST' });
        if (!res.ok) {
            const text = await res.text();
            document.getElementById('factionResult').textContent = '❌ Ошибка: ' + text;
            document.getElementById('factionProgress').style.display = 'none';
            document.getElementById('cancelFactionsBtn').style.display = 'none';
            return;
        }
        if (pollIntervals['factions']) clearInterval(pollIntervals['factions']);
        pollIntervals['factions'] = setInterval(() => pollJob('generate_factions', 'factionProgress', 'factionResult', 'cancelFactionsBtn'), 1500);
    } catch (e) {
        document.getElementById('factionResult').textContent = '❌ ' + e.message;
        document.getElementById('factionProgress').style.display = 'none';
        document.getElementById('cancelFactionsBtn').style.display = 'none';
    }
}

// --- Общий опрос статуса ---
async function pollJob(jobType, progressId, resultId, cancelBtnId) {
    try {
        const res = await fetchWithAuth(`/admin/generate-status?job=${jobType}`);
        const data = await res.json();
        const progress = data.processed || 0;
        const total = data.total || 0;
        const status = data.status || 'running';

        const textEl = document.getElementById(progressId + 'Text');
        const barEl = document.getElementById(progressId + 'Bar');
        if (textEl) textEl.textContent = total > 0 ? `${progress} из ${total}` : 'Подготовка...';
        if (barEl) {
            const percent = total > 0 ? (progress / total * 100) : 0;
            barEl.value = percent;
        }

        const cancelBtn = document.getElementById(cancelBtnId);
        if (cancelBtn) {
            cancelBtn.style.display = (status === 'running') ? 'inline-block' : 'none';
        }

        if (status === 'done' || status === 'error' || status === 'canceled') {
            clearInterval(pollIntervals[jobType]);
            delete pollIntervals[jobType];
            document.getElementById(progressId).style.display = 'none';
            if (cancelBtn) cancelBtn.style.display = 'none';
            const resultEl = document.getElementById(resultId);
            if (status === 'done') {
                resultEl.textContent = `✅ Готово! (${total} объектов)`;
                if (jobType === 'generate_universe') {
                    loadStats();
                    loadWorlds(1);
                } else if (jobType === 'generate_planets') {
                    loadStats();
                }
            } else if (status === 'canceled') {
                resultEl.textContent = `⏹️ Остановлено пользователем`;
            } else if (status === 'error') {
                resultEl.textContent = `❌ Ошибка: ${data.error || 'неизвестная'}`;
            }
        }
    } catch (e) {
        console.error('Poll error:', e);
    }
}

// --- Отмена генерации ---
export async function cancelGeneration(jobType) {
    if (!confirm(`Остановить генерацию?`)) return;
    try {
        const res = await fetchWithAuth(`/admin/generate-cancel?job=${jobType}`, { method: 'POST' });
        if (res.ok) {
            alert('Остановка запрошена');
            const btnId = jobType === 'generate_universe' ? 'cancelUniverseBtn' :
                          jobType === 'generate_planets' ? 'cancelPlanetsBtn' : 'cancelFactionsBtn';
            document.getElementById(btnId).style.display = 'none';
        } else {
            const text = await res.text();
            alert('Ошибка: ' + text);
        }
    } catch (e) {
        alert('Ошибка: ' + e.message);
    }
}

// --- Очистка ---
export async function clearUniverse() {
    if (!confirm('Удалить ВСЕ миры?')) return;
    try {
        const res = await fetchWithAuth('/admin/clear', { method: 'POST', headers: { 'Content-Type': 'application/json' } });
        if (res.ok) { document.getElementById('clearResult').textContent = '✅ Вселенная очищена'; loadStats(); loadWorlds(1); }
        else { const text = await res.text(); document.getElementById('clearResult').textContent = '❌ Ошибка: ' + text; }
    } catch (e) { document.getElementById('clearResult').textContent = '❌ ' + e.message; }
}

// --- Инициализация ---
export function initAdmin() {
    initTabs();
    loadStats();
    loadWorlds(1);

    // Глобальные функции для onclick в HTML
    window.setPassword = setPassword;
    window.loadStats = loadStats;
    window.loadWorlds = loadWorlds;
    window.deleteWorld = deleteWorld;
    window.createWorld = createWorld;
    window.generateUniverse = generateUniverse;
    window.generatePlanets = generatePlanets;
    window.generateFactions = generateFactions;
    window.cancelGeneration = cancelGeneration;
    window.clearUniverse = clearUniverse;
    window.loadPlanetStats = loadPlanetStats;
}