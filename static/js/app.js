/* ===================================================
   Alert Webhook Dashboard - JavaScript (Strict 10-field Version)
   =================================================== */

const API = '/api';
let currentPage = 1;
let currentPageSize = 20;
let totalRecords = 0;

document.addEventListener('DOMContentLoaded', () => {
    initDateDefaults();
    loadStats();
    loadAlerts(1);

    document.getElementById('filterAlertname').addEventListener('keydown', e => {
        if (e.key === 'Enter') loadAlerts(1);
    });

    setInterval(() => {
        loadStats();
        loadAlerts(currentPage);
    }, 30000);
});

function initDateDefaults() {
    const end = new Date();
    const start = new Date();
    start.setDate(start.getDate() - 7);
    document.getElementById('filterEndDate').value = fmtDate(end);
    document.getElementById('filterStartDate').value = fmtDate(start);
}

function fmtDate(d) { return d.toISOString().split('T')[0]; }

function fmtDateTime(s) {
    if (!s) return '—';
    const d = new Date(s);
    return d.toLocaleString('zh-CN', {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit', second: '2-digit',
        hour12: false
    });
}

function esc(s) {
    if (!s) return '';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

async function loadStats() {
    try {
        const r = await fetch(`${API}/alerts/stats`);
        const d = await r.json();

        document.getElementById('totalAlerts').textContent = d.total ?? 0;
        document.getElementById('todayAlerts').textContent = d.today_alerts ?? 0;
        document.getElementById('recentFiring').textContent = d.recent_firing ?? 0;

        let firing = 0, resolved = 0;
        (d.by_status || []).forEach(s => {
            if (s.status === 'firing') firing = s.count;
            else if (s.status === 'resolved') resolved = s.count;
        });
        document.getElementById('firingAlerts').textContent = firing;
        document.getElementById('resolvedAlerts').textContent = resolved;

        document.getElementById('lastUpdate').textContent = '更新于 ' + new Date().toLocaleTimeString('zh-CN');
    } catch (e) { console.error('stats error', e); }
}

async function loadAlerts(page = 1) {
    currentPage = page;
    currentPageSize = parseInt(document.getElementById('pageSizeSelect').value) || 20;

    const params = new URLSearchParams({ page, page_size: currentPageSize });
    const status = document.getElementById('filterStatus').value;
    const alertname = document.getElementById('filterAlertname').value.trim();
    const startDate = document.getElementById('filterStartDate').value;
    const endDate = document.getElementById('filterEndDate').value;

    if (status) params.append('status', status);
    if (alertname) params.append('alert_name', alertname);
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    const tbody = document.getElementById('alertTableBody');
    tbody.innerHTML = `<tr><td colspan="7" class="loading-row"><div class="spinner"></div> 加载中...</td></tr>`;

    try {
        const r = await fetch(`${API}/alerts?${params}`);
        const d = await r.json();

        totalRecords = d.total || 0;
        const pages = Math.ceil(totalRecords / currentPageSize) || 1;
        document.getElementById('paginationInfo').textContent = `共 ${totalRecords} 条，第 ${page}/${pages} 页`;

        renderTable(d.alerts || []);
        renderPagination(page, pages);
    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="7" class="loading-row" style="color:var(--red)">⚠ 加载失败: ${esc(e.message)}</td></tr>`;
    }
}

function renderTable(alerts) {
    const tbody = document.getElementById('alertTableBody');
    if (!alerts.length) {
        tbody.innerHTML = `<tr><td colspan="7" class="loading-row">暂无符合条件的告警记录</td></tr>`;
        return;
    }

    tbody.innerHTML = alerts.map(a => {
        const statusBadge = a.status === 'firing'
            ? `<span class="badge badge-firing">● Firing</span>`
            : `<span class="badge badge-resolved">✓ Resolved</span>`;

        return `<tr>
            <td class="id-cell">#${a.id}</td>
            <td class="alertname-cell">${esc(a.alert_name)}</td>
            <td>${statusBadge}</td>
            <td class="instance-cell">${esc(a.instance) || '—'}</td>
            <td class="time-cell">${fmtDateTime(a.starts_at)}</td>
            <td class="time-cell">${a.ends_at ? fmtDateTime(a.ends_at) : '<span style="color:var(--text-subtle)">—</span>'}</td>
            <td>
                <button class="btn-detail" onclick="showDetail(${a.id})" title="查看详情">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                        <circle cx="12" cy="12" r="3"/>
                    </svg>
                </button>
            </td>
        </tr>`;
    }).join('');
}

function renderPagination(page, pages) {
    const el = document.getElementById('pagination');
    if (pages <= 1) { el.innerHTML = ''; return; }
    let html = `<button class="page-btn" ${page <= 1 ? 'disabled' : ''} onclick="loadAlerts(${page-1})">‹</button>`;
    for (let i = 1; i <= pages; i++) {
        if (i === 1 || i === pages || (i >= page - 2 && i <= page + 2)) {
            html += `<button class="page-btn ${i === page ? 'active' : ''}" onclick="loadAlerts(${i})">${i}</button>`;
        } else if (i === page - 3 || i === page + 3) {
            html += `<span class="page-ellipsis">…</span>`;
        }
    }
    html += `<button class="page-btn" ${page >= pages ? 'disabled' : ''} onclick="loadAlerts(${page+1})">›</button>`;
    el.innerHTML = html;
}

function changePageSize() { loadAlerts(1); }
function resetFilters() {
    document.getElementById('filterStatus').value = '';
    document.getElementById('filterAlertname').value = '';
    initDateDefaults();
    loadAlerts(1);
}

async function showDetail(id) {
    openModal('加载中...', '');
    try {
        const r = await fetch(`${API}/alerts/${id}`);
        const a = await r.json();
        openModal(a.alert_name, `#${a.id}`);

        document.getElementById('modalBody').innerHTML = `
            <div class="detail-section">
                <div class="detail-section-title">基本信息</div>
                <div class="detail-grid">
                    <div class="detail-field">
                        <span class="detail-key">告警名称</span>
                        <span class="detail-val"><strong>${esc(a.alert_name)}</strong></span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">状态</span>
                        <span class="detail-val">${a.status === 'firing' ? '<span class="badge badge-firing">● Firing</span>' : '<span class="badge badge-resolved">✓ Resolved</span>'}</span>
                    </div>
                    <div class="detail-field"><span class="detail-key">实例</span><span class="detail-val mono">${esc(a.instance) || '—'}</span></div>
                    <div class="detail-field"><span class="detail-key">指纹 (Fingerprint)</span><span class="detail-val mono">${esc(a.fingerprint)}</span></div>
                    <div class="detail-field"><span class="detail-key">开始时间</span><span class="detail-val">${fmtDateTime(a.starts_at)}</span></div>
                    <div class="detail-field"><span class="detail-key">结束时间</span><span class="detail-val">${fmtDateTime(a.ends_at)}</span></div>
                </div>
            </div>
            <div class="detail-section">
                <div class="detail-section-title">Labels & Annotations</div>
                <div class="detail-grid">
                    <div class="detail-field detail-field-full"><span class="detail-key">Labels</span><pre class="detail-val mono" style="background:#f8f9fa;padding:8px;border-radius:4px">${JSON.stringify(a.labels, null, 2)}</pre></div>
                    <div class="detail-field detail-field-full"><span class="detail-key">Annotations</span><pre class="detail-val mono" style="background:#f8f9fa;padding:8px;border-radius:4px">${JSON.stringify(a.annotations, null, 2)}</pre></div>
                </div>
            </div>
            <div class="detail-section">
                <div class="detail-section-title">Raw Content (Original JSON)</div>
                <div class="detail-field detail-field-full">
                    <pre class="detail-val mono" style="background:#212529;color:#f8f9fa;padding:12px;border-radius:4px;max-height:400px;overflow:auto">${JSON.stringify(a.raw_content, null, 2)}</pre>
                </div>
            </div>
        `;
    } catch (e) { document.getElementById('modalBody').innerHTML = `<div class="loading-row" style="color:var(--red)">⚠ 加载失败: ${esc(e.message)}</div>`; }
}

function openModal(title, subtitle) {
    document.querySelector('.modal-title').textContent = title || '告警详情';
    document.getElementById('modalSubtitle').textContent = subtitle || '';
    document.getElementById('modalOverlay').classList.add('open');
    document.body.style.overflow = 'hidden';
}

function closeModal() {
    document.getElementById('modalOverlay').classList.remove('open');
    document.body.style.overflow = '';
}

document.addEventListener('keydown', e => { if (e.key === 'Escape') closeModal(); });
