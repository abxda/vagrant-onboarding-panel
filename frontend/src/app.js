/* =========================================================================
 *  Panel de Onboarding (Vagrant) — frontend
 *  Vanilla ES2020. Wails injects window.go.main.App.* and window.runtime.*
 *  Wizard de pasos secuenciales con elevación nativa puntual.
 * ========================================================================= */

/* ---------- ICONS (Lucide inline) ---------------------------------------- */
const ICONS = {
    box:      '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>',
    check:    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>',
    alert:    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><line x1="12" x2="12" y1="9" y2="13"/><line x1="12" x2="12.01" y1="17" y2="17"/></svg>',
    info:     '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>',
    play:     '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="6 3 20 12 6 21 6 3"/></svg>',
    refresh:  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"/><path d="M8 16H3v5"/></svg>',
    shield:   '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/></svg>',
    lock:     '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>',
    trash:    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>',
};
function icon(name) { return ICONS[name] || ICONS.info; }
function esc(s) { return String(s == null ? '' : s).replace(/[&<>"]/g, c => ({ '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;' }[c])); }

/* ---------- state -------------------------------------------------------- */
let STEPS = [];
let CURRENT = 0;          // index of selected step
let ENV = {};
const STATUS_BADGE = {
    ok:      { cls: 'ok',     icon: 'check', label: 'Listo' },
    warn:    { cls: 'warn',   icon: 'alert', label: 'Con avisos' },
    error:   { cls: 'danger', icon: 'alert', label: 'Requiere atención' },
    running: { cls: 'info',   icon: 'refresh', label: 'Ejecutando' },
    unknown: { cls: 'muted',  icon: 'info',  label: 'Sin verificar' },
};

/* ---------- boot --------------------------------------------------------- */
window.addEventListener('DOMContentLoaded', async () => {
    document.getElementById('brandLogo').innerHTML = icon('box');
    ENV = await window.go.main.App.GetEnvInfo();
    STEPS = await window.go.main.App.GetSteps();
    renderFooter();
    renderNav();
    renderStep();
    wireEvents();
    // initial log paint
    const snap = await window.go.main.App.GetLogSnapshot();
    snap.forEach(appendLog);
});

function wireEvents() {
    window.runtime.EventsOn('log', appendLog);
    window.runtime.EventsOn('step:status', ({ id, status }) => {
        const s = STEPS.find(x => x.id === id);
        if (s) s.status = status;
        renderNav();
        if (STEPS[CURRENT] && STEPS[CURRENT].id === id) renderStep();
    });
}

/* ---------- global progress --------------------------------------------- */
function progressPct() {
    if (!STEPS.length) return 0;
    const done = STEPS.filter(s => s.status === 'ok').length;
    return Math.round((done / STEPS.length) * 100);
}

/* ---------- sidebar stepper --------------------------------------------- */
function renderNav() {
    const nav = document.getElementById('stepNav');
    // A step is unlocked if it's the first, or the previous one is ok.
    nav.innerHTML = STEPS.map((s, i) => {
        const st = s.status || 'unknown';
        const numCls = st === 'ok' ? 'c-step-nav__num--ok'
            : st === 'warn' ? 'c-step-nav__num--warn'
            : st === 'error' ? 'c-step-nav__num--error'
            : st === 'running' ? 'c-step-nav__num--running' : '';
        const numInner = st === 'ok' ? icon('check') : (i + 1);
        const active = i === CURRENT ? ' is-active' : '';
        const sub = STATUS_BADGE[st] ? STATUS_BADGE[st].label : '';
        return `<button class="c-step-nav${active}" data-idx="${i}">
            <span class="c-step-nav__num ${numCls}">${numInner}</span>
            <span class="c-step-nav__body">
                <span class="c-step-nav__title">${esc(s.title)}</span>
                <span class="c-step-nav__sub">${esc(sub)}</span>
            </span>
        </button>`;
    }).join('');
    nav.querySelectorAll('button[data-idx]').forEach(b => {
        b.addEventListener('click', () => { CURRENT = parseInt(b.dataset.idx, 10); renderNav(); renderStep(); });
    });
}

/* ---------- header actions ---------------------------------------------- */
function renderHeaderActions() {
    const el = document.getElementById('viewActions');
    el.innerHTML =
        `<button class="c-btn" id="btnTestElev">${icon('shield')} Probar elevación</button>` +
        `<button class="c-btn" id="btnReset">${icon('trash')} Reiniciar progreso</button>`;
    document.getElementById('btnTestElev').addEventListener('click', testElevation);
    document.getElementById('btnReset').addEventListener('click', async () => {
        if (!confirm('¿Reiniciar el progreso del asistente? No se desinstala nada; solo se olvidan los pasos marcados como completos.')) return;
        await window.go.main.App.ResetWizard();
        STEPS = await window.go.main.App.GetSteps();
        CURRENT = 0; renderNav(); renderStep();
    });
}

/* ---------- step detail -------------------------------------------------- */
function renderStep() {
    renderHeaderActions();
    const s = STEPS[CURRENT];
    if (!s) return;
    document.getElementById('viewTitle').textContent = `Paso ${s.index} de ${STEPS.length} · ${s.title}`;

    const st = s.status || 'unknown';
    const badge = STATUS_BADGE[st] || STATUS_BADGE.unknown;
    const running = st === 'running';

    const pct = progressPct();
    const globalBar = `<div class="c-globalbar">
        <div class="c-globalbar__track"><div class="c-globalbar__fill" style="width:${pct}%"></div></div>
        <span class="c-globalbar__pct">${pct}%</span>
    </div>`;

    let elevBlock = '';
    if (s.needsElevation) {
        elevBlock = `<div class="c-elev" id="elevBlock">
            <div class="c-elev__head">${icon('lock')} Esta acción pedirá privilegios de administrador (${esc(ENV.mechanism || '')})</div>
            <div style="font-size:13px;color:#92400e">Antes de ejecutar verás el comando exacto. Aprobarás el diálogo nativo de tu sistema en ese momento; la app nunca corre como administrador completa.</div>
            <pre class="c-elev__cmd" id="elevCmd">cargando…</pre>
        </div>`;
    }

    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="c-card">
            <div class="c-step-meta">
                <span class="c-badge c-badge--${badge.cls}">${icon(badge.icon)} ${esc(badge.label)}</span>
                ${s.needsElevation ? `<span class="c-badge c-badge--warn">${icon('lock')} Requiere administrador</span>` : `<span class="c-badge c-badge--muted">Sin administrador</span>`}
                <div style="flex:1"></div>
                ${globalBar}
            </div>
            <p class="c-step-why">${esc(s.why)}</p>
            ${elevBlock}
            <div style="display:flex;gap:12px;flex-wrap:wrap">
                <button class="c-btn c-btn--primary" id="btnRun" ${running ? 'disabled' : ''}>
                    ${running ? '<span class="c-spinner-sm"></span>' : icon('play')} ${esc(s.actionLabel)}
                </button>
                <button class="c-btn" id="btnCheck" ${running ? 'disabled' : ''}>${icon('refresh')} Verificar estado</button>
            </div>
            <div id="stepResult"></div>
            <div id="diagProbes"></div>
        </div>

        <div class="c-console">
            <div class="c-console__toolbar">
                <span class="c-console__title">Registro en vivo</span>
                <button class="c-btn" id="btnClearLog">${icon('trash')} Limpiar</button>
            </div>
            <pre class="c-console__pane" id="logPane"></pre>
        </div>
    `;

    // re-paint existing logs into the new pane
    window.go.main.App.GetLogSnapshot().then(snap => { snap.forEach(appendLog); });

    if (s.needsElevation) {
        window.go.main.App.PreviewElevation(s.id).then(cmd => {
            const el = document.getElementById('elevCmd');
            if (el) el.textContent = cmd || '(no disponible)';
        });
    }

    document.getElementById('btnRun').addEventListener('click', () => runStep(s));
    document.getElementById('btnCheck').addEventListener('click', () => checkStep(s));

    // The diagnostic step renders a probe table; run it automatically the
    // first time it's viewed if it hasn't been checked yet.
    if (s.id === 'diagnostico') {
        if (s.status === 'unknown' || s.status === 'running') runDiagnostics();
        else if (LAST_DIAG) renderDiagnostics(LAST_DIAG);
    }
    document.getElementById('btnClearLog').addEventListener('click', async () => {
        await window.go.main.App.ClearLog();
        document.getElementById('logPane').innerHTML = '';
    });
}

async function runStep(s) {
    setResult('', '');
    if (s.id === 'diagnostico') { await runDiagnostics(); return; }
    const res = await window.go.main.App.RunStep(s.id);
    showActionResult(res);
    STEPS = await window.go.main.App.GetSteps();
    renderNav();
}

let LAST_DIAG = null;
async function runDiagnostics() {
    const btn = document.getElementById('btnRun');
    if (btn) { btn.disabled = true; btn.innerHTML = '<span class="c-spinner-sm"></span> Diagnosticando…'; }
    try {
        const rep = await window.go.main.App.GetDiagnostics();
        LAST_DIAG = rep;
        renderDiagnostics(rep);
        const s = STEPS.find(x => x.id === 'diagnostico');
        if (s) s.status = rep.overall;
        renderNav();
        if (rep.overall === 'ok') setResult('ok', `${icon('check')} Diagnóstico OK. Tu equipo está listo para virtualizar.`);
        else if (rep.overall === 'warn') setResult('warn', `${icon('alert')} Diagnóstico con avisos. Puedes continuar, pero revisa los puntos en amarillo.`);
        else setResult('err', `${icon('alert')} El diagnóstico encontró un bloqueo. Revisa los puntos en rojo antes de continuar.`);
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('play')} Ejecutar diagnóstico`; }
    }
}

function renderDiagnostics(rep) {
    const host = document.getElementById('diagProbes');
    if (!host || !rep) return;
    const rows = (rep.probes || []).map(p => {
        const badge = p.level === 'ok' ? 'ok' : p.level === 'warn' ? 'warn' : 'danger';
        const bicon = p.level === 'ok' ? 'check' : 'alert';
        const advice = (p.advice && p.level !== 'ok')
            ? `<div class="c-diag__advice">${icon('info')} ${esc(p.advice)}</div>` : '';
        return `<div class="c-diag__row c-diag__row--${badge}">
            <div class="c-diag__head">
                <span class="c-badge c-badge--${badge}">${icon(bicon)} ${p.level.toUpperCase()}</span>
                <span class="c-diag__label">${esc(p.label)}</span>
                <span class="c-diag__value">${esc(p.value)}</span>
            </div>
            ${advice}
        </div>`;
    }).join('');
    host.innerHTML = `<div class="c-diag">${rows}</div>`;
}

async function checkStep(s) {
    const status = await window.go.main.App.CheckStep(s.id);
    s.status = status;
    renderNav(); renderStep();
}

function showActionResult(res) {
    if (!res) return;
    if (res.ok) setResult('ok', `${icon('check')} ${esc(res.message || 'Listo')}`);
    else if (res.cancelled) setResult('warn', `${icon('alert')} ${esc(res.message || 'Cancelado')}`);
    else setResult('err', `${icon('alert')} ${esc(res.message || 'Falló')}${res.detail ? ' — ' + esc(res.detail) : ''}`);
}
function setResult(kind, html) {
    const el = document.getElementById('stepResult');
    if (!el) return;
    if (!kind) { el.innerHTML = ''; return; }
    el.innerHTML = `<div class="c-result c-result--${kind}">${html}</div>`;
}

async function testElevation() {
    const btn = document.getElementById('btnTestElev');
    if (btn) { btn.disabled = true; btn.innerHTML = '<span class="c-spinner-sm"></span> Probando…'; }
    try {
        const res = await window.go.main.App.TestElevation();
        showActionResult(res);
        alert(res.ok
            ? '✓ Elevación funcionando: ' + (res.detail ? '' : '') + (res.message || '')
            : (res.cancelled ? 'Cancelaste el diálogo de elevación.' : 'La elevación no se confirmó: ' + (res.message || '')));
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('shield')} Probar elevación`; }
    }
}

/* ---------- live log ----------------------------------------------------- */
function appendLog(ln) {
    const pane = document.getElementById('logPane');
    if (!pane) return;
    const lv = (ln.level || 'INFO').toLowerCase();
    const cls = lv === 'error' ? 'error' : lv === 'warn' ? 'warn' : 'info';
    const line = document.createElement('div');
    line.className = 'c-log c-log--' + cls;
    line.innerHTML = `<span class="c-log__ts">${esc(ln.time || '')}</span><span class="c-log__lv">${esc((ln.level||'INFO'))}</span><span class="c-log__text">${esc(ln.text || '')}</span>`;
    pane.appendChild(line);
    pane.scrollTop = pane.scrollHeight;
}

/* ---------- footer ------------------------------------------------------- */
function renderFooter() {
    const el = document.getElementById('footerInfo');
    el.innerHTML =
        `${esc((ENV.os || '').toUpperCase())} · ${esc(ENV.arch || '')}<br>` +
        `Elevación: ${esc(ENV.mechanism || '')}<br>` +
        `<span style="opacity:.7">v${esc(ENV.appVersion || '')} · ${esc(ENV.author || 'Dr. Abel Coronado')}</span>`;
}
