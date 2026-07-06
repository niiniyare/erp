/* ============================================================
   AWO ERP — Shell Logic
   ------------------------------------------------------------
   Responsibilities:
   1. Boot the app shell (sidebar, topbar, theme, mobile nav)
   2. Fetch nav from the backend:        GET  /sdui/nav
   3. Fetch full AMIS page schemas from: GET  /sdui/{entityId}
   4. Wire AMIS's fetcher to our REST conventions (offset/limit,
      {status,msg,data} envelope, tenant header, bearer token)

   AUTH IS CURRENTLY DISABLED — see the "AUTH (DISABLED)" section
   near the bottom. The backend is expected to fully own page
   rendering: every route hands amis-ui a complete, permission-
   gated, feature-flag-aware schema. Once the auth endpoints are
   finalized, uncomment that section and remove the DEV_* stubs.
   ============================================================ */

(function () {
  'use strict';

  /* ── Config ────────────────────────────────────────────────── */
  var NAV_API   = '/sdui/nav';     // returns [{ group, items:[{id,label,icon,module}] }]
  var SDUI_BASE = '/sdui';         // GET /sdui/{entityId} -> full amis schema (JSON)
  var DASHBOARD_SCHEMA = '/sdui/__dashboard';

  /* ── Dev stubs (remove once AUTH section below is re-enabled) ─
     Since login is disabled for now, we assume a single dev
     tenant and no bearer token. The backend should treat requests
     with no Authorization header as an authenticated dev session,
     or short-circuit its own auth middleware in this environment.
  ------------------------------------------------------------- */
  var currentTenant = localStorage.getItem('awo-tenant') || 'dev';
  var currentToken  = localStorage.getItem('awo-token') || '';

  /* ── State ─────────────────────────────────────────────────── */
  var currentInstance = null;
  var navIndex = {}; // entityId -> { label, module }

  /* ── Theme ─────────────────────────────────────────────────── */
  var themes = ['system', 'light', 'dark'];
  var themeIdx = themes.indexOf(localStorage.getItem('awo-theme') || 'system');

  function applyTheme() {
    var t = themes[themeIdx];
    document.documentElement.setAttribute('data-theme', t);
    var label = document.getElementById('themeLabel');
    if (label) label.textContent = t.charAt(0).toUpperCase() + t.slice(1);
    localStorage.setItem('awo-theme', t);
  }
  window.cycleTheme = function () {
    themeIdx = (themeIdx + 1) % themes.length;
    applyTheme();
  };
  applyTheme();

  /* ── Sidebar collapse / mobile drawer ─────────────────────────── */
  var sidebar = document.getElementById('sidebar');
  var sidebarCollapsed = localStorage.getItem('awo-sidebar') === 'collapsed';
  if (sidebarCollapsed) sidebar.classList.add('collapsed');

  window.toggleSidebar = function () {
    sidebarCollapsed = !sidebarCollapsed;
    sidebar.classList.toggle('collapsed', sidebarCollapsed);
    localStorage.setItem('awo-sidebar', sidebarCollapsed ? 'collapsed' : '');
  };

  var backdrop = document.getElementById('sidebar-backdrop');
  window.openMobile  = function () { sidebar.classList.add('mobile-open'); backdrop.classList.add('visible'); };
  window.closeMobile = function () { sidebar.classList.remove('mobile-open'); backdrop.classList.remove('visible'); };

  /* ── AMIS fetcher ──────────────────────────────────────────────
     Translates amis' page/perPage query params into offset/limit,
     unwraps our {status,msg,data} envelope, and attaches tenant +
     auth headers to every request.
  ------------------------------------------------------------- */
  function amisFetcher(api) {
    var url    = api.url || '';
    var method = (api.method || 'get').toLowerCase();

    var urlObj  = new URL(url, window.location.origin);
    var page    = urlObj.searchParams.get('page') || urlObj.searchParams.get('$page');
    var perPage = urlObj.searchParams.get('perPage') || urlObj.searchParams.get('$perPage');

    if (page !== null) {
      var limit  = perPage ? parseInt(perPage, 10) : 20;
      var offset = (parseInt(page, 10) - 1) * limit;
      urlObj.searchParams.delete('page');
      urlObj.searchParams.delete('$page');
      urlObj.searchParams.set('offset', String(offset));
    }
    if (perPage !== null) {
      urlObj.searchParams.delete('perPage');
      urlObj.searchParams.delete('$perPage');
      urlObj.searchParams.set('limit', perPage);
    }
    url = urlObj.toString();

    var headers = Object.assign(
      {
        'Content-Type': 'application/json',
        'X-Awo-Tenant': currentTenant
      },
      currentToken ? { 'Authorization': 'Bearer ' + currentToken } : {},
      api.headers || {}
    );

    var opts = { method: method, headers: headers, credentials: 'include' };
    if (method !== 'get' && method !== 'head' && api.data) {
      opts.body = JSON.stringify(api.data);
    }

    return fetch(url, opts)
      .then(function (res) {
        // AUTH (DISABLED): 401 handling kept here so it's a no-op
        // drop-in once login is restored — just uncomment showLogin().
        if (res.status === 401) {
          console.warn('[amis] 401 Unauthorized — auth is currently disabled in shell.js');
          // showLogin('Session expired — please sign in again.');
          return { status: 401, msg: 'Unauthorized' };
        }
        if (res.status === 204) return { status: 0, msg: '' };
        return res.json().then(function (body) {
          if (body.status === undefined) {
            if (!res.ok) return { status: res.status, msg: body.error || body.msg || res.statusText };
            return { status: 0, msg: '', data: body };
          }
          return body;
        });
      })
      .catch(function (err) {
        return { status: 1, msg: 'Network error: ' + err.message };
      });
  }

  var amisEmbed = amisRequire('amis/embed');
  var amisEnv = {
    locale: 'en-US',
    fetcher: amisFetcher,
    isCancel: function () { return false; },
    notify: function (type, msg) { console.log('[amis:' + type + ']', msg); },
    alert: window.alert.bind(window),
    copy: function (text) { if (navigator.clipboard) navigator.clipboard.writeText(text); }
  };

  /* ── Nav: fetched from backend, backend owns structure/labels/icons ── */
  function loadNav() {
    fetch(NAV_API, {
      credentials: 'include',
      headers: Object.assign(
        { 'X-Awo-Tenant': currentTenant },
        currentToken ? { 'Authorization': 'Bearer ' + currentToken } : {}
      )
    })
      .then(function (r) { return r.ok ? r.json() : []; })
      .then(function (groups) {
        var menu = document.getElementById('sidebarMenu');
        menu.innerHTML = '';

        // Dashboard entry is always first and implicit.
        var dash = document.createElement('a');
        dash.className = 'nav-item';
        dash.id = 'nav-__dashboard';
        dash.href = '#__dashboard';
        dash.innerHTML = '<i class="fa fa-chart-line"></i><span>Dashboard</span>';
        dash.addEventListener('click', closeMobile);
        menu.appendChild(dash);
        navIndex['__dashboard'] = { label: 'Dashboard', module: '' };

        groups.forEach(function (g) {
          var grpEl = document.createElement('div');
          grpEl.className = 'nav-group-label';
          grpEl.textContent = g.group;
          menu.appendChild(grpEl);

          g.items.forEach(function (item) {
            navIndex[item.id] = item;
            var a = document.createElement('a');
            a.className = 'nav-item';
            a.id = 'nav-' + item.id;
            a.href = '#' + item.id;
            var icon = item.icon || moduleIcon(g.group);
            a.innerHTML = '<i class="fa ' + icon + '"></i><span>' + item.label + '</span>';
            a.addEventListener('click', closeMobile);
            menu.appendChild(a);
          });
        });

        navigate();
      })
      .catch(function () {
        document.getElementById('sidebarMenu').innerHTML =
          '<div style="padding:16px;color:#94a3b8;font-size:13px;">Failed to load nav</div>';
      });
  }

  function moduleIcon(mod) {
    var icons = {
      'Finance': 'fa-book', 'Inventory': 'fa-box', 'HR': 'fa-users',
      'CRM': 'fa-building', 'Platform': 'fa-shield', 'Forecourt': 'fa-gas-pump'
    };
    return icons[mod] || 'fa-table';
  }

  /* ── Router ────────────────────────────────────────────────────
     Every entity route (including dashboard) is a full, backend-
     rendered amis schema fetched from /sdui/{entityId}. The shell
     no longer builds any page markup itself — the backend owns
     permission gating, feature flags, columns, forms, everything.
  ------------------------------------------------------------- */
  var content = document.getElementById('content');

  function setActive(entity) {
    document.querySelectorAll('.nav-item').forEach(function (el) { el.classList.remove('active'); });
    var el = document.getElementById('nav-' + entity);
    if (el) el.classList.add('active');

    var item = navIndex[entity];
    if (item) {
      document.getElementById('bc-page').textContent = item.label || entity;
      if (item.module) {
        document.getElementById('bc-module').textContent = item.module;
        document.getElementById('bc-sep').style.display = '';
      } else {
        document.getElementById('bc-module').textContent = '';
        document.getElementById('bc-sep').style.display = 'none';
      }
    }
  }

  function navigate() {
    var entity = window.location.hash.slice(1) || '__dashboard';
    setActive(entity);

    if (currentInstance) {
      try { currentInstance.unmount(); } catch (e) {}
      currentInstance = null;
    }
    content.innerHTML = '<div class="page-loading"><i class="fa fa-spinner fa-spin"></i> Loading…</div>';

    var schemaURL = entity === '__dashboard' ? DASHBOARD_SCHEMA : SDUI_BASE + '/' + entity;

    fetch(schemaURL, {
      credentials: 'include',
      headers: Object.assign(
        { 'X-Awo-Tenant': currentTenant },
        currentToken ? { 'Authorization': 'Bearer ' + currentToken } : {}
      )
    })
      .then(function (r) {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.json();
      })
      .then(function (schema) {
        // Backend is expected to return a ready-to-embed amis schema,
        // e.g. { type: 'page', title: '...', body: [...] }
        content.innerHTML = '';
        currentInstance = amisEmbed('#content', schema, {}, amisEnv);
      })
      .catch(function (err) {
        content.innerHTML = '<div class="page-error"><h3>Page error</h3><p>' + entity + ' — ' + err.message + '</p></div>';
      });
  }

  window.addEventListener('hashchange', navigate);

  /* ============================================================
     AUTH (DISABLED)
     ------------------------------------------------------------
     Everything below is intentionally inert. The login overlay
     markup/CSS still exists in shell.html / shell.css so this can
     be switched back on without rebuilding the UI. To re-enable:

       1. Delete the DEV_* stubs above (currentTenant/currentToken
          hardcoded values) and restore reads from localStorage
          only after a successful login.
       2. Uncomment showLogin()/hideLogin() and call showLogin()
          instead of loadNav() at the bottom of this file.
       3. Wire submitLogin/submitPlatformLogin/submitCreateTenant
          back to window.* and call checkSession() on boot instead
          of loadNav() directly.
       4. Re-enable the 401 handler above (showLogin) inside
          amisFetcher.
  ============================================================ */

  /*
  function showLogin(msg) {
    document.getElementById('login-overlay').classList.remove('hidden');
    if (currentTenant) document.getElementById('login-tenant').value = currentTenant;
    var errEl = document.getElementById('login-error');
    if (msg) { errEl.textContent = msg; errEl.classList.add('visible'); }
    else      { errEl.classList.remove('visible'); }
  }

  function hideLogin() {
    document.getElementById('login-overlay').classList.add('hidden');
    document.getElementById('login-error').classList.remove('visible');
  }

  window.switchTab = function (tab) {
    var isLogin = tab === 'login';
    var isPlatform = tab === 'platform';
    document.getElementById('login-form').style.display    = isLogin ? '' : 'none';
    document.getElementById('platform-form').style.display = isPlatform ? '' : 'none';
    document.getElementById('tab-login').className    = 'tab-btn' + (isLogin ? ' active' : '');
    document.getElementById('tab-platform').className = 'tab-btn' + (isPlatform ? ' active' : '');
  };

  window.submitLogin = function (e) {
    e.preventDefault();
    var tenant = document.getElementById('login-tenant').value.trim();
    var email  = document.getElementById('login-email').value.trim();
    var pass   = document.getElementById('login-password').value;
    var btn    = document.getElementById('login-submit');
    var errEl  = document.getElementById('login-error');

    btn.disabled = true; btn.textContent = 'Signing in…';
    errEl.classList.remove('visible');

    fetch('/auth/login', {
      method: 'POST', credentials: 'include',
      headers: { 'Content-Type': 'application/json', 'X-Awo-Tenant': tenant },
      body: JSON.stringify({ email: email, password: pass })
    })
    .then(function (r) { return r.json(); })
    .then(function (body) {
      if (body.status === 0 && body.data) {
        currentTenant = tenant;
        currentToken  = body.data.access_token || '';
        localStorage.setItem('awo-tenant', tenant);
        localStorage.setItem('awo-token', currentToken);
        document.getElementById('userName').textContent = body.data.display_name || email;
        hideLogin();
        loadNav();
      } else {
        errEl.textContent = body.msg || 'Login failed';
        errEl.classList.add('visible');
      }
    })
    .catch(function (err) { errEl.textContent = 'Network error: ' + err.message; errEl.classList.add('visible'); })
    .finally(function () { btn.disabled = false; btn.textContent = 'Sign in'; });
  };

  window.doLogout = function () {
    fetch('/auth/logout', {
      method: 'POST', credentials: 'include',
      headers: { 'Authorization': 'Bearer ' + currentToken, 'X-Awo-Tenant': currentTenant }
    }).finally(function () {
      localStorage.removeItem('awo-token');
      currentToken = '';
      showLogin();
    });
  };

  function checkSession() {
    if (!currentToken) { showLogin(); return; }
    fetch('/auth/me', {
      credentials: 'include',
      headers: { 'Authorization': 'Bearer ' + currentToken, 'X-Awo-Tenant': currentTenant }
    })
    .then(function (r) { return r.ok ? r.json() : null; })
    .then(function (body) {
      if (body && body.status === 0 && body.data) {
        document.getElementById('userName').textContent = body.data.display_name || body.data.user_id;
        hideLogin();
        loadNav();
      } else {
        showLogin();
      }
    })
    .catch(function () { showLogin(); });
  }
  */

  /* ── Boot ──────────────────────────────────────────────────────
     Auth disabled: skip session check, go straight to nav + router.
     A visible "DEV MODE" badge is set on the user pill so nobody
     mistakes this for a real authenticated session.
  ------------------------------------------------------------- */
  var userNameEl = document.getElementById('userName');
  if (userNameEl) userNameEl.textContent = 'Dev (no auth)';
  document.getElementById('login-overlay').classList.add('hidden');

  loadNav();
  // checkSession(); // ← restore this and remove the two lines above when auth is re-enabled
})();
