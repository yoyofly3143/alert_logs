/* ===================================================
   Alert Webhook Dashboard - JavaScript
   =================================================== */

const API = '/api';
let currentPage = 1;
let currentPageSize = 20;
let totalRecords = 0;

// ── Init ────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
    initDateDefaults();
    loadStats();
    loadAlerts(1);

    // Enter key 触发查询
    document.getElementById('filterAlertname').addEventListener('keydown', e => {
        if (e.key === 'Enter') loadAlerts(1);
    });

    // 每 30 秒自动刷新
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

function fmtDate(d) {
    return d.toISOString().split('T')[0];
}

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

// ── Stats ────────────────────────────────────────────
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

        document.getElementById('lastUpdate').textContent =
            '更新于 ' + new Date().toLocaleTimeString('zh-CN');
    } catch (e) {
        console.error('stats error', e);
    }
}

// ── Alerts List ────────────────────────────────────
async function loadAlerts(page = 1) {
    currentPage = page;
    currentPageSize = parseInt(document.getElementById('pageSizeSelect').value) || 20;

    const params = new URLSearchParams({
        page,
        page_size: currentPageSize
    });

    const status = document.getElementById('filterStatus').value;
    const severity = document.getElementById('filterSeverity').value;
    const alertname = document.getElementById('filterAlertname').value.trim();
    const startDate = document.getElementById('filterStartDate').value;
    const endDate = document.getElementById('filterEndDate').value;

    if (status) params.append('status', status);
    if (severity) params.append('severity', severity);
    if (alertname) params.append('alertname', alertname);
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    const tbody = document.getElementById('alertTableBody');
    tbody.innerHTML = `<tr><td colspan="9" class="loading-row"><div class="spinner"></div> 加载中...</td></tr>`;

    try {
        const r = await fetch(`${API}/alerts?${params}`);
        const d = await r.json();

        totalRecords = d.total || 0;
        const pages = Math.ceil(totalRecords / currentPageSize) || 1;

        document.getElementById('paginationInfo').textContent =
            `共 ${totalRecords} 条，第 ${page}/${pages} 页`;

        renderTable(d.alerts || []);
        renderPagination(page, pages);
    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="9" class="loading-row" style="color:var(--red)">⚠ 加载失败: ${esc(e.message)}</td></tr>`;
    }
}

function renderTable(alerts) {
    const tbody = document.getElementById('alertTableBody');
    if (!alerts.length) {
        tbody.innerHTML = `<tr><td colspan="9" class="loading-row">暂无符合条件的告警记录</td></tr>`;
        return;
    }

    tbody.innerHTML = alerts.map(a => {
        const statusBadge = a.status === 'firing'
            ? `<span class="badge badge-firing">● Firing</span>`
            : `<span class="badge badge-resolved">✓ Resolved</span>`;

        const sev = (a.severity || '').toLowerCase();
        let sevBadge = '';
        if (sev === 'critical') sevBadge = `<span class="badge badge-critical">Critical</span>`;
        else if (sev === 'warning') sevBadge = `<span class="badge badge-warning">Warning</span>`;
        else if (sev === 'info') sevBadge = `<span class="badge badge-info">Info</span>`;
        else if (sev) sevBadge = `<span class="badge badge-unknown">${esc(a.severity)}</span>`;
        else sevBadge = `<span class="badge badge-unknown">—</span>`;

        const instance = a.instance
            ? `<div>${esc(a.instance)}</div>`
            : '';
        const job = a.job
            ? `<div style="color:var(--text-subtle)">${esc(a.job)}</div>`
            : '';

        return `<tr>
            <td class="id-cell">#${a.id}</td>
            <td class="alertname-cell">${esc(a.alertname)}</td>
            <td>${statusBadge}</td>
            <td>${sevBadge}</td>
            <td class="instance-cell">${instance}${job}</td>
            <td class="summary-cell" title="${esc(a.summary)}">${esc(a.summary) || '<span style="color:var(--text-subtle)">—</span>'}</td>
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

// ── Pagination ────────────────────────────────────
function renderPagination(page, pages) {
    const el = document.getElementById('pagination');
    if (pages <= 1) { el.innerHTML = ''; return; }

    let html = '';

    html += `<button class="page-btn" ${page <= 1 ? 'disabled' : ''} onclick="loadAlerts(${page-1})">‹</button>`;

    const maxShow = 5;
    let start = Math.max(1, page - Math.floor(maxShow/2));
    let end = Math.min(pages, start + maxShow - 1);
    if (end - start < maxShow - 1) start = Math.max(1, end - maxShow + 1);

    if (start > 1) {
        html += `<button class="page-btn" onclick="loadAlerts(1)">1</button>`;
        if (start > 2) html += `<span class="page-ellipsis">…</span>`;
    }

    for (let i = start; i <= end; i++) {
        html += `<button class="page-btn ${i === page ? 'active' : ''}" onclick="loadAlerts(${i})">${i}</button>`;
    }

    if (end < pages) {
        if (end < pages - 1) html += `<span class="page-ellipsis">…</span>`;
        html += `<button class="page-btn" onclick="loadAlerts(${pages})">${pages}</button>`;
    }

    html += `<button class="page-btn" ${page >= pages ? 'disabled' : ''} onclick="loadAlerts(${page+1})">›</button>`;

    el.innerHTML = html;
}

function changePageSize() {
    loadAlerts(1);
}

function resetFilters() {
    document.getElementById('filterStatus').value = '';
    document.getElementById('filterSeverity').value = '';
    document.getElementById('filterAlertname').value = '';
    initDateDefaults();
    loadAlerts(1);
}

function forceRefresh() {
    const btn = document.getElementById('btnRefresh');
    btn.classList.add('spinning');
    Promise.all([loadStats(), loadAlerts(currentPage)]).finally(() => {
        btn.classList.remove('spinning');
    });
}

// ── Alert Detail Modal ────────────────────────────
async function showDetail(id) {
    openModal('加载中...', '');
    document.getElementById('modalBody').innerHTML =
        `<div class="loading-row"><div class="spinner"></div> 加载详情...</div>`;
    try {
        const r = await fetch(`${API}/alerts/${id}`);
        const a = await r.json();

        document.getElementById('modalSubtitle').textContent =
            `${a.alertname}  #${a.id}`;
        openModal(a.alertname, `#${a.id}`);

        const labels = a.labels || {};
        const annotations = a.annotations || {};

        document.getElementById('modalBody').innerHTML = `
            <div class="detail-section">
                <div class="detail-section-title">基本信息</div>
                <div class="detail-grid">
                    <div class="detail-field">
                        <span class="detail-key">告警名称</span>
                        <span class="detail-val"><strong>${esc(a.alertname)}</strong></span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">状态 / 级别</span>
                        <span class="detail-val">
                            ${a.status === 'firing'
                                ? '<span class="badge badge-firing">● Firing</span>'
                                : '<span class="badge badge-resolved">✓ Resolved</span>'}
                            ${a.severity ? `<span class="badge badge-${(a.severity||'').toLowerCase()}" style="margin-left:4px">${esc(a.severity)}</span>` : ''}
                        </span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">实例 (instance)</span>
                        <span class="detail-val mono" style="font-size:12px;padding:4px 8px">${esc(a.instance) || '—'}</span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">Job</span>
                        <span class="detail-val mono" style="font-size:12px;padding:4px 8px">${esc(a.job) || '—'}</span>
                    </div>
                    ${a.cluster ? `<div class="detail-field"><span class="detail-key">Cluster</span><span class="detail-val">${esc(a.cluster)}</span></div>` : ''}
                    ${a.env ? `<div class="detail-field"><span class="detail-key">环境</span><span class="detail-val">${esc(a.env)}</span></div>` : ''}
                    <div class="detail-field">
                        <span class="detail-key">Receiver</span>
                        <span class="detail-val">${esc(a.receiver) || '—'}</span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">Fingerprint</span>
                        <span class="detail-val mono" style="font-size:12px;padding:4px 8px">${esc(a.fingerprint)}</span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">开始时间</span>
                        <span class="detail-val">${fmtDateTime(a.starts_at)}</span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">结束时间</span>
                        <span class="detail-val">${a.ends_at ? fmtDateTime(a.ends_at) : '—'}</span>
                    </div>
                    <div class="detail-field">
                        <span class="detail-key">接收时间</span>
                        <span class="detail-val">${fmtDateTime(a.created_at)}</span>
                    </div>
                    ${a.generator_url ? `
                    <div class="detail-field detail-field-full">
                        <span class="detail-key">Generator URL</span>
                        <span class="detail-val"><a href="${esc(a.generator_url)}" target="_blank" class="detail-link">${esc(a.generator_url)}</a></span>
                    </div>` : ''}
                    ${a.runbook ? `
                    <div class="detail-field detail-field-full">
                        <span class="detail-key">Runbook</span>
                        <span class="detail-val"><a href="${esc(a.runbook)}" target="_blank" class="detail-link">${esc(a.runbook)}</a></span>
                    </div>` : ''}
                </div>
            </div>

            ${a.summary || a.description ? `
            <div class="detail-section">
                <div class="detail-section-title">告警描述</div>
                <div class="detail-grid">
                    ${a.summary ? `
                    <div class="detail-field detail-field-full">
                        <span class="detail-key">摘要 (summary)</span>
                        <span class="detail-val">${esc(a.summary)}</span>
                    </div>` : ''}
                    ${a.description ? `
                    <div class="detail-field detail-field-full">
                        <span class="detail-key">描述 (description)</span>
                        <span class="detail-val">${esc(a.description)}</span>
                    </div>` : ''}
                </div>
            </div>` : ''}

            <div class="detail-section">
                <div class="detail-section-title">Labels</div>
                <div class="detail-field">
                    <span class="detail-val mono">${JSON.stringify(labels, null, 2)}</span>
                </div>
            </div>

            <div class="detail-section">
                <div class="detail-section-title">Annotations</div>
                <div class="detail-field">
                    <span class="detail-val mono">${JSON.stringify(annotations, null, 2)}</span>
                </div>
            </div>
        `;
    } catch (e) {
        document.getElementById('modalBody').innerHTML =
            `<div class="loading-row" style="color:var(--red)">⚠ 加载失败: ${esc(e.message)}</div>`;
    }
}

function openModal(title, subtitle) {
    document.getElementById('modalBody').querySelector && null;
    document.querySelector('.modal-title').textContent = title || '告警详情';
    document.getElementById('modalSubtitle').textContent = subtitle || '';
    document.getElementById('modalOverlay').classList.add('open');
    document.body.style.overflow = 'hidden';
}

function closeModal() {
    document.getElementById('modalOverlay').classList.remove('open');
    document.body.style.overflow = '';
}

// ESC 键关闭弹窗
document.addEventListener('keydown', e => {
    if (e.key === 'Escape') closeModal();
});
