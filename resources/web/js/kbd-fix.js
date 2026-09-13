(function () {
  window.rdPortalLogout = function () {
    const token = localStorage.getItem("access_token") || localStorage.getItem("wc-option:local:access_token") || "";
    const headers = { "Content-Type": "application/json" };
    if (token) headers["api-token"] = token;
    const finish = function () {
      const paths = ["/", "/_admin", "/_admin/", "/webclient", "/webclient/"];
      paths.forEach(function (path) {
        document.cookie = "rd_portal=; Path=" + path + "; SameSite=Lax; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT";
        document.cookie = "rd_portal=; Path=" + path + "; SameSite=Lax; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Secure";
      });
      ["access_token", "wc-option:local:access_token", "user_info"].forEach(function (k) {
        localStorage.removeItem(k);
      });
      const go = function () {
        window.location.href = window.location.origin + "/_admin/index.html?logout=" + Date.now() + "#/login";
      };
      if (navigator.serviceWorker) {
        navigator.serviceWorker.getRegistrations().then(function (regs) {
          return Promise.all(regs.map(function (r) { return r.unregister(); }));
        }).catch(function () {}).finally(go);
      } else {
        go();
      }
    };
    const ctrl = typeof AbortController !== "undefined" ? new AbortController() : null;
    const timer = setTimeout(function () { if (ctrl) ctrl.abort(); }, 2000);
    fetch("/api/admin/logout", {
      method: "POST",
      credentials: "include",
      headers: headers,
      body: "{}",
      signal: ctrl ? ctrl.signal : undefined,
    }).catch(function () {}).finally(function () {
      clearTimeout(timer);
      finish();
    });
  };

  const style = document.createElement("style");
  style.textContent =
    ".rd-bar{position:fixed;left:0;right:0;bottom:0;z-index:99999;display:none;gap:10px;align-items:center;padding:8px 16px;font:13px/1.3 Manrope,system-ui,sans-serif}" +
    ".rd-bar label{font-weight:700;white-space:nowrap}" +
    ".rd-bar input{flex:1;height:36px;border-radius:10px;padding:0 12px;font:15px/1 Manrope,system-ui,sans-serif}" +
    ".rd-bar button{height:36px;border:0;border-radius:10px;padding:0 14px;color:#fff;font:700 13px Manrope,system-ui,sans-serif;cursor:pointer}";
  document.head.appendChild(style);

  const peerBar = document.createElement("div");
  peerBar.className = "rd-bar";
  peerBar.id = "rd-peer-bar";
  peerBar.innerHTML =
    '<label>Contraseña RustDesk</label>' +
    '<input id="rd-peer-in" type="password" autocomplete="off" placeholder="Clave del equipo" />' +
    '<button type="button" id="rd-peer-go">Conectar</button>';

  const typeBar = document.createElement("div");
  typeBar.className = "rd-bar";
  typeBar.id = "rd-kbd-bar";
  typeBar.innerHTML =
    '<label>Clave Windows</label>' +
    '<input id="rd-kbd-in" type="text" autocomplete="off" autocapitalize="off" spellcheck="false" placeholder="Escribe aquí la clave (incluye ! @)" />' +
    '<button type="button" id="rd-kbd-bang">!</button>' +
    '<button type="button" id="rd-kbd-at">@</button>' +
    '<button type="button" id="rd-kbd-go">Enviar</button>';

  function mount() {
    if (!document.body) return;
    if (!peerBar.parentNode) document.body.appendChild(peerBar);
    if (!typeBar.parentNode) document.body.appendChild(typeBar);
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mount);
  } else {
    mount();
  }

  function conn() {
    return window.curConn;
  }
  function loggedIn() {
    const c = conn();
    return !!(c && c._peerInfo);
  }
  function waitingPeerPassword() {
    const c = conn();
    return !!(c && c._hash && !c._peerInfo);
  }

  let lastFocus = "";
  function showBars() {
    const peer = waitingPeerPassword();
    const type = loggedIn();
    document.body.classList.toggle("rd-in-session", type);
    peerBar.style.display = peer ? "flex" : "none";
    typeBar.style.display = type ? "flex" : "none";
    if (peer && lastFocus !== "peer") {
      lastFocus = "peer";
    } else if (type && lastFocus !== "type") {
      lastFocus = "type";
    } else if (!peer && !type) {
      lastFocus = "";
    }
  }
  setInterval(showBars, 300);

  function sendPeer() {
    const c = conn();
    const input = document.getElementById("rd-peer-in");
    if (!c || !input || typeof c.login !== "function") return;
    const v = input.value;
    if (!v) return;
    c.login(v);
  }

  function sendChar(ch) {
    const c = conn();
    if (!c || typeof c.inputKey !== "function") return;
    if (ch === "\n") {
      c.inputKey("VK_RETURN", true, true, false, false, false, false);
      return;
    }
    // Winlogon ignores unicode seq; send Shift+physical key for ! @ etc.
    c.inputKey(ch, false, true, false, false, false, false);
  }

  function sendOsPassword() {
    const c = conn();
    const input = document.getElementById("rd-kbd-in");
    if (!c || !input) return;
    const v = input.value;
    if (!v) return;
    if (typeof c.inputOsPassword === "function") {
      c.inputOsPassword(v, false);
    } else {
      for (let i = 0; i < v.length; i++) sendChar(v[i]);
    }
    input.value = "";
  }

  function isHelper(el) {
    return el && (el.id === "rd-kbd-in" || el.id === "rd-peer-in");
  }

  document.addEventListener("click", function (e) {
    if (e.target && e.target.id === "rd-peer-go") sendPeer();
    if (e.target && e.target.id === "rd-kbd-go") sendOsPassword();
    if (e.target && e.target.id === "rd-kbd-bang") {
      const input = document.getElementById("rd-kbd-in");
      if (input) input.value += "!";
    }
    if (e.target && e.target.id === "rd-kbd-at") {
      const input = document.getElementById("rd-kbd-in");
      if (input) input.value += "@";
    }
    if (e.target && e.target.id === "rd-wc-logout") {
      e.preventDefault();
      window.rdPortalLogout && window.rdPortalLogout();
    }
  });
  ["keyup", "keypress"].forEach(function (ev) {
    document.addEventListener(ev, function (e) {
      if (isHelper(e.target)) e.stopPropagation();
    }, true);
  });
  document.addEventListener("keydown", function (e) {
    if (!isHelper(e.target)) return;
    e.stopPropagation();
    if (e.target.id === "rd-peer-in" && e.key === "Enter") {
      e.preventDefault();
      sendPeer();
    }
    if (e.target.id === "rd-kbd-in" && e.key === "Enter") {
      e.preventDefault();
      sendOsPassword();
    }
  }, true);
})();
