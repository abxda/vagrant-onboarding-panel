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
    copy:     '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>',
    terminal: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" x2="20" y1="19" y2="19"/></svg>',
    book:     '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20"/></svg>',
    folder:   '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>',
    server:   '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="2" rx="2"/><rect width="20" height="8" x="2" y="14" rx="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>',
    stop:     '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="5" y="5" rx="2"/></svg>',
    power:    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>',
};
function icon(name) { return ICONS[name] || ICONS.info; }
function esc(s) { return String(s == null ? '' : s).replace(/[&<>"]/g, c => ({ '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;' }[c])); }

/* ---------- state -------------------------------------------------------- */
let STEPS = [];
let CURRENT = 0;          // index of selected step (in PREPARACIÓN)
let MODE = 'prep';        // 'prep' = Preparación (despliegue) | 'lab' = Mi Laboratorio (trabajo)
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
    // Si la preparación ya está completa, entrar directo a Mi Laboratorio.
    const prep = STEPS.filter(isPrepStep);
    const prepDone = prep.length && prep.every(s => s.status === 'ok' || s.status === 'warn');
    MODE = prepDone ? 'lab' : 'prep';
    renderNav();
    if (MODE === 'lab') renderLab(); else renderStep();
    wireEvents();
    // initial log paint
    const snap = await window.go.main.App.GetLogSnapshot();
    snap.forEach(appendLog);
    // always-present status dashboard: refresh now and on an interval
    renderStatusStrip();
    refreshDashboard();
    setInterval(refreshDashboard, 15000);
});

function wireEvents() {
    window.runtime.EventsOn('log', appendLog);
    window.runtime.EventsOn('step:status', ({ id, status }) => {
        const s = STEPS.find(x => x.id === id);
        if (s) s.status = status;
        renderNav();
        if (STEPS[CURRENT] && STEPS[CURRENT].id === id) renderStep();
    });
    window.runtime.EventsOn('exercise:progress', (p) => {
        EX_PROGRESS = (p && p.current > 0) ? p : null;
        renderStatusStrip();
        if (!EX_PROGRESS) refreshDashboard();
    });
    window.runtime.EventsOn('services:update', (svc) => {
        LAST_DASH.services = svc || LAST_DASH.services;
        renderStatusStrip();
        if (MODE === 'lab' && LAB_TAB === 'servicios') renderLabServicios();
    });
    // Cerrado elegante (homologado con el portable): al cerrar el panel se
    // apaga la VM limpiamente y se muestra un overlay de progreso.
    window.runtime.EventsOn('shutdown:start', () => showShutdownOverlay());
    window.runtime.EventsOn('shutdown:progress', (d) => updateShutdownOverlay(d));
    window.runtime.EventsOn('shutdown:done', () => {
        const m = document.getElementById('shutdownMsg');
        if (m) m.textContent = 'Máquina virtual apagada. Cerrando…';
    });
}

function showShutdownOverlay() {
    if (document.getElementById('shutdownOverlay')) return;
    const o = document.createElement('div');
    o.id = 'shutdownOverlay';
    o.className = 'c-overlay';
    o.innerHTML =
        '<div class="c-overlay__box">' +
            '<div class="c-spinner"></div>' +
            '<h2 class="c-overlay__title">Cerrando de forma segura</h2>' +
            '<p class="c-overlay__msg" id="shutdownMsg">Preparando el apagado…</p>' +
            '<p class="c-overlay__note">' +
                'Detenemos los servicios y apagamos la máquina virtual para no dejar nada ' +
                'corriendo ni los puertos ocupados. Tu trabajo en HDFS se conserva. ' +
                'Puede tardar ~30 segundos.' +
            '</p>' +
        '</div>';
    document.body.appendChild(o);
}

function updateShutdownOverlay(d) {
    const msg = document.getElementById('shutdownMsg');
    if (msg && d && d.msg) msg.textContent = d.msg;
}

/* ---------- global progress --------------------------------------------- */
function progressPct() {
    if (!STEPS.length) return 0;
    const done = STEPS.filter(s => s.status === 'ok').length;
    return Math.round((done / STEPS.length) * 100);
}

/* ---------- sidebar: dos capas (Preparación / Mi Laboratorio) ------------ */
// El paso 'servidores' deja de ser el "paso 6" y se convierte en la puerta a
// MI LABORATORIO (capa de trabajo, idéntica entre ediciones). Los pasos
// previos son PREPARACIÓN (capa de despliegue, propia de Vagrant).
const LAB_STEP_ID = 'servidores';
function isPrepStep(s) { return s.id !== LAB_STEP_ID; }

function renderNav() {
    const nav = document.getElementById('stepNav');
    const prep = STEPS.filter(isPrepStep);
    const prepDone = prep.length && prep.every(s => s.status === 'ok' || s.status === 'warn');

    const stepBtn = (s) => {
        const i = STEPS.indexOf(s);
        const st = s.status || 'unknown';
        const numCls = st === 'ok' ? 'c-step-nav__num--ok'
            : st === 'warn' ? 'c-step-nav__num--warn'
            : st === 'error' ? 'c-step-nav__num--error'
            : st === 'running' ? 'c-step-nav__num--running' : '';
        const numInner = st === 'ok' ? icon('check') : (prep.indexOf(s) + 1);
        const active = (i === CURRENT && MODE === 'prep') ? ' is-active' : '';
        const sub = STATUS_BADGE[st] ? STATUS_BADGE[st].label : '';
        return `<button class="c-step-nav${active}" data-idx="${i}">
            <span class="c-step-nav__num ${numCls}">${numInner}</span>
            <span class="c-step-nav__body">
                <span class="c-step-nav__title">${esc(s.title)}</span>
                <span class="c-step-nav__sub">${esc(sub)}</span>
            </span>
        </button>`;
    };

    const labActive = MODE === 'lab' ? ' is-active' : '';
    // El laboratorio se habilita si la preparación está completa O si la VM ya
    // está corriendo: si la VM está arriba es que VirtualBox + Vagrant + caja +
    // 'vagrant up' funcionaron, aunque algún paso se haya quedado sin marcar.
    // Esto evita que el botón "Mi laboratorio" quede bloqueado por un estado de
    // paso incompleto (la causa que reportó el alumno).
    const vmRunning = LAST_DASH.vmState === 'running';
    const labReady = prepDone || vmRunning;
    const labLocked = !labReady;
    nav.innerHTML =
        `<div class="c-navgroup">Preparación</div>` +
        prep.map(stepBtn).join('') +
        `<div class="c-navgroup">Mi laboratorio</div>` +
        `<button class="c-step-nav c-step-nav--lab${labActive}" id="navLab" ${labLocked ? 'disabled title="Completa la preparación primero"' : ''}>
            <span class="c-step-nav__num ${labReady ? 'c-step-nav__num--ok' : ''}">${icon('server')}</span>
            <span class="c-step-nav__body">
                <span class="c-step-nav__title">Mi laboratorio</span>
                <span class="c-step-nav__sub">${labLocked ? 'Completa la preparación' : 'Servicios · Ejercicios · HDFS'}</span>
            </span>
        </button>`;

    nav.querySelectorAll('button[data-idx]').forEach(b => {
        b.addEventListener('click', () => { MODE = 'prep'; CURRENT = parseInt(b.dataset.idx, 10); renderNav(); renderStep(); });
    });
    const navLab = document.getElementById('navLab');
    if (navLab && !labLocked) {
        navLab.addEventListener('click', () => { MODE = 'lab'; renderNav(); renderLab(); });
    }
}

/* ---------- header actions ---------------------------------------------- */
function renderHeaderActions() {
    const el = document.getElementById('viewActions');
    const vmRunning = LAST_DASH.vmState === 'running';
    el.innerHTML =
        `<button class="c-btn" id="btnJupyter" ${vmRunning ? '' : 'disabled'} title="Abre Jupyter Lab en el navegador">${icon('book')} Abrir Jupyter</button>` +
        `<button class="c-btn" id="btnFolder" title="Abre la carpeta de trabajo (lo que pongas ahí lo verás en Jupyter)">${icon('folder')} Carpeta de trabajo</button>` +
        `<button class="c-btn" id="btnSmoke" title="Descarga el cuaderno de prueba (TestGlobalBigData) a tu carpeta de trabajo; aparecerá en Jupyter">${icon('book')} Cuaderno de prueba</button>` +
        `<button class="c-btn" id="btnSSH">${icon('terminal')} Consola SSH</button>` +
        `<button class="c-btn" id="btnReset">${icon('trash')} Reiniciar</button>`;
    document.getElementById('btnJupyter').addEventListener('click', openJupyter);
    document.getElementById('btnFolder').addEventListener('click', openFolder);
    document.getElementById('btnSmoke').addEventListener('click', downloadSmoke);
    document.getElementById('btnSSH').addEventListener('click', openSSH);
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
            <div style="display:flex;gap:12px;flex-wrap:wrap${s.id === 'servidores' ? ';display:none' : ''}" id="stepButtons">
                <button class="c-btn c-btn--primary" id="btnRun" ${running ? 'disabled' : ''}>
                    ${running ? '<span class="c-spinner-sm"></span>' : icon('play')} ${esc(stepActionLabel(s))}
                </button>
                <button class="c-btn" id="btnCheck" ${running ? 'disabled' : ''}>${icon('refresh')} Verificar estado</button>
            </div>
            <div id="stepResult"></div>
            <div id="toolStatus"></div>
            <div id="diagProbes"></div>
            <div id="exerciseBox"></div>
        </div>

        <div class="c-console">
            <div class="c-console__toolbar">
                <span class="c-console__title">Registro en vivo</span>
                <button class="c-btn c-btn--primary" id="btnCopyLog">${icon('copy')} Copiar consola</button>
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
    // Install steps auto-detect whether the tool is already present.
    if (s.id === 'virtualbox' || s.id === 'vagrant') {
        refreshToolStatus(s.id);
    }
    document.getElementById('btnClearLog').addEventListener('click', async () => {
        await window.go.main.App.ClearLog();
        document.getElementById('logPane').innerHTML = '';
    });
    document.getElementById('btnCopyLog').addEventListener('click', copyConsole);
}

/* ---------- MI LABORATORIO (capa de trabajo, idéntica entre ediciones) ---- */
// Sub-secciones del laboratorio. La UI aquí es homóloga al portable: el alumno
// no ve "vagrant", solo su espacio de trabajo (servicios, ejercicio, HDFS).
let LAB_TAB = 'servicios';
function renderLab() {
    renderHeaderActions();
    document.getElementById('viewTitle').textContent = 'Mi laboratorio';

    const tabs = [
        { id: 'servicios',  label: 'Servicios',  icon: 'server' },
        { id: 'ejercicios', label: 'Ejercicios', icon: 'book' },
        { id: 'hdfs',       label: 'HDFS',       icon: 'folder' },
    ];
    const subtabs = tabs.map(t =>
        `<button class="c-subtab${LAB_TAB === t.id ? ' is-active' : ''}" data-labtab="${t.id}">${icon(t.icon)} ${t.label}</button>`
    ).join('');

    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="c-card" style="padding:0;overflow:hidden">
            <div class="c-subtabs">${subtabs}</div>
            <div id="labBody" style="padding:var(--s-5)"></div>
        </div>
        <div class="c-console">
            <div class="c-console__toolbar">
                <span class="c-console__title">Registro en vivo</span>
                <button class="c-btn c-btn--primary" id="btnCopyLog">${icon('copy')} Copiar consola</button>
                <button class="c-btn" id="btnClearLog">${icon('trash')} Limpiar</button>
            </div>
            <pre class="c-console__pane" id="logPane"></pre>
        </div>`;

    content.querySelectorAll('button[data-labtab]').forEach(b => {
        b.addEventListener('click', () => { LAB_TAB = b.dataset.labtab; renderLab(); });
    });
    document.getElementById('btnClearLog').addEventListener('click', async () => {
        await window.go.main.App.ClearLog();
        document.getElementById('logPane').innerHTML = '';
    });
    document.getElementById('btnCopyLog').addEventListener('click', copyConsole);
    window.go.main.App.GetLogSnapshot().then(snap => snap.forEach(appendLog));

    // Render the active sub-tab into #labBody.
    if (LAB_TAB === 'servicios')  renderLabServicios();
    else if (LAB_TAB === 'ejercicios') renderLabEjercicios();
    else if (LAB_TAB === 'hdfs')  renderLabHdfs();
    refreshDashboard();
}

// copyConsole copies the FULL log buffer (not just what's visible) to the
// clipboard, prefixed with a context header so that when a student pastes it
// to the teacher, the OS/arch/version and the active step come with it.
async function copyConsole() {
    const snap = await window.go.main.App.GetLogSnapshot();
    const s = STEPS[CURRENT] || {};
    const header = [
        '===== Panel de Onboarding · reporte del alumno =====',
        `Sistema: ${ENV.os || '?'}/${ENV.arch || '?'}  ·  Panel v${ENV.appVersion || '?'}`,
        `Elevación: ${ENV.mechanism || '?'}`,
        `Paso actual: ${s.index || '?'} · ${s.title || '?'} (estado: ${s.status || '?'})`,
        '===================================================',
        '',
    ].join('\n');
    const body = snap.map(l => `${l.time}  ${(l.level || 'INFO').padEnd(5)}  ${l.text}`).join('\n');
    const ok = await copyText(header + body);
    const btn = document.getElementById('btnCopyLog');
    if (btn) {
        const orig = btn.innerHTML;
        btn.innerHTML = `${icon('check')} ${ok ? '¡Copiado! Pégalo en tu mensaje' : 'No se pudo copiar'}`;
        setTimeout(() => { btn.innerHTML = orig; }, 2200);
    }
}

// copyText tries the modern clipboard API, then the Wails runtime, then a
// hidden-textarea fallback — so it works across WebView2 / WKWebView / WebKitGTK.
async function copyText(text) {
    try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            await navigator.clipboard.writeText(text);
            return true;
        }
    } catch (e) { /* fall through */ }
    try {
        if (window.runtime && window.runtime.ClipboardSetText) {
            await window.runtime.ClipboardSetText(text);
            return true;
        }
    } catch (e) { /* fall through */ }
    try {
        const ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.opacity = '0';
        document.body.appendChild(ta);
        ta.select();
        const ok = document.execCommand('copy');
        document.body.removeChild(ta);
        return ok;
    } catch (e) {
        return false;
    }
}

async function runStep(s) {
    setResult('', '');
    if (s.id === 'diagnostico') { await runDiagnostics(); return; }
    const btn = document.getElementById('btnRun');
    if (btn) { btn.disabled = true; btn.innerHTML = `<span class="c-spinner-sm"></span> ${esc(s.actionLabel)}…`; }
    try {
        const res = await window.go.main.App.RunStep(s.id);
        showActionResult(res);
        STEPS = await window.go.main.App.GetSteps();
        renderNav();
        if (s.id === 'virtualbox' || s.id === 'vagrant') await refreshToolStatus(s.id);
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('play')} ${esc(s.actionLabel)}`; }
    }
}

async function refreshToolStatus(stepId) {
    const ts = await window.go.main.App.GetToolStatus(stepId);
    renderToolStatus(ts);
    const s = STEPS.find(x => x.id === stepId);
    if (s) { s.status = ts.installed ? 'ok' : (s.status === 'error' ? 'error' : s.status); renderNav(); }
}

function renderToolStatus(ts) {
    const host = document.getElementById('toolStatus');
    if (!host || !ts) return;
    if (ts.installed) {
        host.innerHTML = `<div class="c-result c-result--ok">${icon('check')} Ya instalado — versión ${esc(ts.version)} <span style="opacity:.7;margin-left:6px">${esc(ts.path)}</span></div>`;
    } else {
        const pm = ts.pkgAvailable
            ? `Se instalará con <code>${esc(ts.pkgManager)}</code> (requiere aprobar el diálogo de administrador).`
            : `No encontré un gestor de paquetes; tendrás que instalarlo manualmente desde el sitio oficial.`;
        host.innerHTML = `<div class="c-result c-result--warn">${icon('alert')} No instalado. ${pm}</div>`;
    }
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

/* ---------- Status strip (always-present dashboard) ---------------------- */
const VM_LABELS = {
    running:     { cls: 'ok',    text: 'VM encendida' },
    poweroff:    { cls: 'muted', text: 'VM apagada' },
    not_created: { cls: 'muted', text: 'VM no creada' },
    saved:       { cls: 'warn',  text: 'VM suspendida' },
    aborted:     { cls: 'danger',text: 'VM abortada' },
    '':          { cls: 'muted', text: 'Vagrant no listo' },
};
let LAST_DASH = { vmState: '', services: {}, hasVM: false };
let EX_PROGRESS = null; // {current,total,title} while the exercise runs

function dot(cls) { return `<span class="c-dot c-dot--${cls}"></span>`; }

function renderStatusStrip() {
    const el = document.getElementById('statusStrip');
    if (!el) return;
    const vm = VM_LABELS[LAST_DASH.vmState] || { cls: 'muted', text: 'VM: ' + (LAST_DASH.vmState || '—') };
    const running = LAST_DASH.vmState === 'running';
    const svc = LAST_DASH.services || {};
    const svcDot = (on) => running ? (on ? 'ok' : 'muted') : 'muted';

    const services = `
        <span class="c-chip" title="HDFS (NameNode + DataNode)">${dot(svcDot(svc.hdfs))} HDFS</span>
        <span class="c-chip" title="Apache Kafka">${dot(svcDot(svc.kafka))} Kafka</span>
        <span class="c-chip" title="Elasticsearch">${dot(svcDot(svc.elastic))} Elastic</span>
        <span class="c-chip" title="Jupyter Lab">${dot(svcDot(svc.jupyter))} Jupyter</span>`;

    let progress = '';
    if (EX_PROGRESS && EX_PROGRESS.current > 0) {
        const pct = Math.round((EX_PROGRESS.current / EX_PROGRESS.total) * 100);
        progress = `<span class="c-strip__progress">
            <span class="c-strip__progresslabel">Ejercicio ${EX_PROGRESS.current}/${EX_PROGRESS.total}: ${esc(EX_PROGRESS.title || '')}</span>
            <span class="c-strip__bar"><span class="c-strip__fill" style="width:${pct}%"></span></span>
        </span>`;
    }

    el.innerHTML =
        `<span class="c-chip c-chip--vm">${dot(vm.cls)} ${esc(vm.text)}</span>` +
        `<span class="c-strip__sep"></span>` +
        services +
        (progress ? `<span class="c-strip__sep"></span>` + progress : '');
}

async function refreshDashboard() {
    try {
        const d = await window.go.main.App.GetDashboard();
        if (!d) return;
        const prevVm = LAST_DASH.vmState;
        // No degradar el estado: si una consulta tarda y vuelve con vmState
        // vacío o 'unknown' (vagrant ocupado), conservamos el último estado
        // conocido en vez de mostrar "Vagrant no listo" — eso parecía que la VM
        // había muerto al detener servicios, cuando seguía encendida.
        if ((!d.vmState || d.vmState === 'unknown') && LAST_DASH.vmState) {
            LAST_DASH = { ...LAST_DASH, services: d.services || LAST_DASH.services };
        } else {
            LAST_DASH = d;
        }
        renderStatusStrip();
        // Si cambió el estado de la VM, refrescamos el nav: así "Mi laboratorio"
        // se desbloquea solo en cuanto la VM está corriendo.
        if (LAST_DASH.vmState !== prevVm) renderNav();
    } catch (e) { /* vagrant not ready — keep last */ }
}

async function openSSH() {
    const btn = document.getElementById('btnSSH');
    if (btn) { btn.disabled = true; btn.innerHTML = '<span class="c-spinner-sm"></span> Abriendo…'; }
    try {
        const res = await window.go.main.App.OpenVagrantSSH();
        if (!res.ok) alert(res.message || 'No se pudo abrir la consola SSH.');
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('terminal')} Consola SSH`; }
    }
}

async function openJupyter() {
    const btn = document.getElementById('btnJupyter');
    if (btn) { btn.disabled = true; btn.innerHTML = '<span class="c-spinner-sm"></span> Abriendo…'; }
    try {
        const res = await window.go.main.App.OpenJupyter();
        if (!res.ok) alert(res.message || 'No se pudo abrir Jupyter.');
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('book')} Abrir Jupyter`; }
    }
}

async function openFolder() {
    const res = await window.go.main.App.OpenWorkFolder();
    if (!res.ok) alert(res.message || 'No se pudo abrir la carpeta.');
}

async function downloadSmoke() {
    const btn = document.getElementById('btnSmoke');
    if (btn) { btn.disabled = true; btn.innerHTML = `${icon('book')} Descargando…`; }
    try {
        const res = await window.go.main.App.DownloadSmokeTest();
        if (res.ok) {
            alert('Cuaderno de prueba descargado a tu carpeta de trabajo.\nÁbrelo en Jupyter (aparece en la carpeta montada) y ejecútalo de arriba a abajo.');
        } else {
            alert(res.message || 'No se pudo descargar el cuaderno de prueba.');
        }
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('book')} Cuaderno de prueba`; }
    }
}

// stepActionLabel ajusta el texto del botón principal. En el Paso 6, si los
// servicios ya están corriendo, "Iniciar servicios" confunde -> lo cambiamos
// a "Verificar e iniciar ejercicio" para que el botón tenga sentido.
function stepActionLabel(s) {
    if (s.id === 'servidores') {
        const svc = LAST_DASH.services || {};
        if (LAST_DASH.vmState === 'running' && svc.hdfs) {
            return 'Ejecutar todo el ejercicio';
        }
        return 'Iniciar servicios y ejercicio';
    }
    return s.actionLabel;
}

/* ---------- Mi Laboratorio: estado del ejercicio ------------------------- */
let EX_STEPS = [];
// Repinta la sub-pestaña activa del laboratorio si estamos en modo lab.
function refreshLabActive() {
    if (MODE !== 'lab') return;
    if (LAB_TAB === 'servicios')  renderLabServicios();
    else if (LAB_TAB === 'ejercicios') renderLabEjercicios();
    else if (LAB_TAB === 'hdfs')  renderLabHdfs();
}

/* === Lab · Servicios ===================================================== */
function renderLabServicios() {
    const host = document.getElementById('labBody');
    if (!host) return;
    const vmRunning = LAST_DASH.vmState === 'running';
    const svc = LAST_DASH.services || {};
    const svcChip = (on, label) => {
        const cls = vmRunning && on ? 'ok' : 'muted';
        const txt = vmRunning && on ? 'activo' : (vmRunning ? 'apagado' : '—');
        return `<div class="c-svc"><span class="c-dot c-dot--${cls}"></span><span class="c-svc__name">${label}</span><span class="c-svc__state">${txt}</span></div>`;
    };
    const allUp = vmRunning && svc.hdfs && svc.kafka && svc.elastic && svc.jupyter;
    const someUp = vmRunning && (svc.hdfs || svc.kafka || svc.elastic || svc.jupyter);
    host.innerHTML = `
        <div class="c-svcpanel" style="margin:0">
            <div class="c-svcpanel__actionbar">
                <button class="c-btn c-btn--primary c-btn--lg" id="svcStartAll" ${vmRunning ? '' : 'disabled'}>${icon('play')} Levantar todos los servicios</button>
                <button class="c-btn c-btn--lg" id="svcStopAll" ${someUp ? '' : 'disabled'}>${icon('stop')} Detener todos</button>
                <div style="flex:1"></div>
                <button class="c-btn c-btn--sm" id="exRefreshSvc">${icon('refresh')} Actualizar</button>
                <button class="c-btn c-btn--sm c-btn--danger" id="vmShutdown" ${vmRunning ? '' : 'disabled'} title="Apaga la máquina virtual y libera los puertos (para poder usar el Portable)">${icon('power')} Apagar VM</button>
            </div>
            <div class="c-svcpanel__hint">${
                !vmRunning ? `${icon('alert')} La máquina virtual no está encendida. Vuelve a Preparación y completa "Levantar VM".`
                : allUp ? `${icon('check')} Todos los servicios están <b>activos</b>. Déjalo abierto durante la clase: al cerrar el panel se apagará la VM de forma segura.`
                : someUp ? `${icon('info')} Algunos servicios están apagados (en gris). Pulsa <b>"Levantar todos los servicios"</b> — enciende los que faltan sin reiniciar los que ya corren.`
                : `${icon('info')} Los servicios están apagados. Pulsa <b>"Levantar todos los servicios"</b> para encenderlos.`
            }</div>
            <div class="c-svcpanel__grid">
                ${svcChip(svc.hdfs, 'HDFS')}
                ${svcChip(svc.kafka, 'Kafka')}
                ${svcChip(svc.elastic, 'Elasticsearch')}
                ${svcChip(svc.jupyter, 'Jupyter')}
            </div>
        </div>`;
    document.getElementById('exRefreshSvc').addEventListener('click', async () => { await refreshDashboard(); renderLabServicios(); });
    document.getElementById('svcStartAll').addEventListener('click', async (e) => {
        const b = e.currentTarget;
        b.disabled = true; b.innerHTML = '<span class="c-spinner-sm"></span> Levantando servicios… (~30s)';
        const res = await window.go.main.App.StartAllServices();
        if (!res.ok) alert(res.message || 'No se pudieron levantar los servicios.');
        await refreshDashboard(); renderLabServicios();
    });
    document.getElementById('svcStopAll').addEventListener('click', async (e) => {
        if (!confirm('¿Detener todos los servicios dentro de la VM? El ejercicio dejará de funcionar hasta que los vuelvas a levantar. La VM seguirá encendida.')) return;
        const b = e.currentTarget;
        b.disabled = true; b.innerHTML = '<span class="c-spinner-sm"></span> Deteniendo…';
        const res = await window.go.main.App.StopAllServices();
        if (!res.ok) alert(res.message || 'No se pudieron detener los servicios.');
        await refreshDashboard(); renderLabServicios();
    });
    const vmBtn = document.getElementById('vmShutdown');
    if (vmBtn) vmBtn.addEventListener('click', async (e) => {
        if (!confirm('¿Apagar la máquina virtual?\n\nEsto detiene los servicios limpiamente y apaga la VM, liberando los puertos (9870/9200/8888) para que puedas usar el Portable. Tu trabajo en HDFS se conserva; la próxima vez que levantes la VM seguirá ahí.')) return;
        const b = e.currentTarget;
        b.disabled = true; b.innerHTML = '<span class="c-spinner-sm"></span> Apagando VM… (~30s)';
        const res = await window.go.main.App.ShutdownVM();
        if (!res.ok) alert(res.message || 'No se pudo apagar la VM.');
        await refreshDashboard(); renderLabServicios();
    });
}

/* === Lab · Ejercicios ==================================================== */
async function renderLabEjercicios() {
    const host = document.getElementById('labBody');
    if (!host) return;
    if (!EX_STEPS.length) {
        try { EX_STEPS = await window.go.main.App.GetExerciseSteps(); }
        catch (e) { host.innerHTML = `<div class="c-alert c-alert--warn">${icon('alert')} No pude cargar el ejercicio.</div>`; return; }
    }
    const vmRunning = LAST_DASH.vmState === 'running';
    const rows = EX_STEPS.map((s, i) => `
        <div class="c-ex-step" id="exstep-${i}">
            <div class="c-ex-step__head">
                <span class="c-ex-step__num">${s.index}</span>
                <span class="c-ex-step__title">${esc(s.title.replace(/^\d+\s·\s/, ''))}</span>
                <button class="c-btn c-btn--sm" data-exrun="${i}" ${vmRunning ? '' : 'disabled'}>${icon('play')} Ejecutar</button>
            </div>
            <div class="c-ex-step__notes">${esc(s.notes)}</div>
            <pre class="c-ex-step__cmd">${esc(s.cmd)}</pre>
        </div>`).join('');
    host.innerHTML = `
        <div class="c-ex" style="border:0;margin:0;padding:0">
            <div class="c-ex__head">
                <span class="c-ex__title">${icon('book')} Ejercicio_01 · WordCount paso a paso</span>
                <div style="flex:1"></div>
                <button class="c-btn" id="exOpenFolder" title="Abre la carpeta con mapper.py, reducer.py y los datos">${icon('folder')} Ver archivos</button>
                <button class="c-btn c-btn--primary" id="exRunAll" ${vmRunning ? '' : 'disabled'}>${icon('play')} Ejecutar todo</button>
            </div>
            <p class="c-ex__intro">Cuenta cuántas veces aparece cada palabra en un dataset usando MapReduce (Hadoop Streaming). Puedes correr cada paso por separado y ver su salida en el registro de abajo, o ejecutar todo de una vez. ${vmRunning ? '' : '<b>Necesitas los servicios encendidos</b> (pestaña Servicios).'}</p>
            ${rows}
        </div>`;
    document.getElementById('exOpenFolder').addEventListener('click', async () => {
        const res = await window.go.main.App.OpenExerciseFolder();
        if (!res.ok) alert(res.message || 'No se pudo abrir la carpeta del ejercicio.');
    });
    host.querySelectorAll('button[data-exrun]').forEach(b => {
        b.addEventListener('click', () => runExerciseStep(parseInt(b.dataset.exrun, 10), b));
    });
    document.getElementById('exRunAll').addEventListener('click', async () => {
        const res = await window.go.main.App.RunExerciseStep(0); // ensureServices + paso 1
        // Tras el primero, encadenar el resto vía "ejecutar todo" del backend:
        await runAllExerciseSteps();
        showActionResult(res);
    });
}

// runAllExerciseSteps ejecuta los pasos 2..N en orden (el 1 ya lo lanzó quien
// llama, que además asegura servicios). Reusa RunExerciseStep para coherencia.
async function runAllExerciseSteps() {
    for (let i = 1; i < EX_STEPS.length; i++) {
        const res = await window.go.main.App.RunExerciseStep(i);
        if (!res.ok) { showActionResult(res); return; }
    }
}

/* === Lab · HDFS ========================================================== */
function renderLabHdfs() {
    const host = document.getElementById('labBody');
    if (!host) return;
    const vmRunning = LAST_DASH.vmState === 'running';
    host.innerHTML = `
        <div class="c-ex" style="border:0;margin:0;padding:0">
            <div class="c-ex__head">
                <span class="c-ex__title">${icon('folder')} Explorador de HDFS</span>
                <div style="flex:1"></div>
                <button class="c-btn c-btn--primary" id="hdfsRefresh" ${vmRunning ? '' : 'disabled'}>${icon('refresh')} Cargar / Actualizar</button>
            </div>
            <p class="c-ex__intro">Navega el sistema de archivos distribuido (HDFS). Haz click en una carpeta para abrirla. Aquí verás el input y el output de los ejercicios.</p>
            <div class="c-hdfs-tree" id="hdfsTree"><div class="c-hdfs-empty">Pulsa "Cargar / Actualizar" para ver HDFS.</div></div>
        </div>`;
    document.getElementById('hdfsRefresh').addEventListener('click', loadHdfsRoot);
}

async function runExerciseStep(idx, btn) {
    if (btn) { btn.disabled = true; btn.innerHTML = '<span class="c-spinner-sm"></span> Ejecutando…'; }
    try {
        const res = await window.go.main.App.RunExerciseStep(idx);
        showActionResult(res);
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('play')} Ejecutar`; }
    }
}

async function runExerciseStep(idx, btn) {
    if (btn) { btn.disabled = true; btn.innerHTML = '<span class="c-spinner-sm"></span> Ejecutando…'; }
    try {
        const res = await window.go.main.App.RunExerciseStep(idx);
        showActionResult(res);
    } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = `${icon('play')} Ejecutar`; }
    }
}

/* ---------- Explorador HDFS (homologado con el portable) ------------------ */
async function loadHdfsRoot() {
    const tree = document.getElementById('hdfsTree');
    if (!tree) return;
    tree.innerHTML = `<div class="c-hdfs-empty"><span class="c-spinner-sm"></span> Cargando HDFS…</div>`;
    try {
        const entries = await window.go.main.App.ListHDFS('/');
        tree.innerHTML = '';
        if (!entries || !entries.length) {
            tree.innerHTML = `<div class="c-hdfs-empty">HDFS está vacío (raíz /).</div>`;
            return;
        }
        entries.forEach(e => tree.appendChild(hdfsNode(e, '/' + e.name)));
    } catch (err) {
        tree.innerHTML = `<div class="c-hdfs-empty">No pude leer HDFS: ${esc(String(err))}<br><span style="opacity:.7">¿Están los servicios arriba? (HDFS en :9870)</span></div>`;
    }
}

function hdfsNode(entry, path) {
    const node = document.createElement('div');
    node.className = 'c-hdfs-node';
    const isDir = entry.type === 'DIRECTORY';
    if (!isDir) {
        const kb = entry.length >= 1024 ? (entry.length / 1024).toFixed(1) + ' KB' : entry.length + ' B';
        node.innerHTML = `<div class="c-hdfs-file">${icon('book')} ${esc(entry.name)} <span style="opacity:.6">· ${kb}</span></div>`;
        return node;
    }
    node.innerHTML =
        `<div class="c-hdfs-node__row" data-path="${esc(path)}">` +
            `<span class="c-hdfs-node__chev">▶</span> ${icon('folder')} ${esc(entry.name)}` +
        `</div>` +
        `<div class="c-hdfs-node__children" style="display:none;"></div>`;
    const row = node.querySelector('.c-hdfs-node__row');
    const childrenEl = node.querySelector('.c-hdfs-node__children');
    let loaded = false, expanded = false;
    row.addEventListener('click', async () => {
        expanded = !expanded;
        row.querySelector('.c-hdfs-node__chev').textContent = expanded ? '▼' : '▶';
        childrenEl.style.display = expanded ? 'block' : 'none';
        if (expanded && !loaded) {
            loaded = true;
            childrenEl.innerHTML = `<div class="c-hdfs-empty"><span class="c-spinner-sm"></span> …</div>`;
            try {
                const kids = await window.go.main.App.ListHDFS(path);
                childrenEl.innerHTML = '';
                if (!kids || !kids.length) {
                    childrenEl.innerHTML = `<div class="c-hdfs-empty">(carpeta vacía)</div>`;
                } else {
                    kids.forEach(k => childrenEl.appendChild(hdfsNode(k, path + '/' + k.name)));
                }
            } catch (err) {
                childrenEl.innerHTML = `<div class="c-hdfs-empty">Error: ${esc(String(err))}</div>`;
            }
        }
    });
    return node;
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
