// web/static/js/admin/auth.js
import { loadStats } from './stats.js';
import { loadWorlds } from './worlds.js';

let adminPassword = localStorage.getItem('adminPassword') || '';

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