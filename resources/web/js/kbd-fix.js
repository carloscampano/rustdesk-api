(function () {
  window.rdPortalLogout = function () {
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
    '<label>Escribir en Windows</label>' +
    '<input id="rd-kbd-in" type="text" autocomplete="off" autocapitalize="off" spellcheck="false" placeholder="Clic en el campo remoto, luego escribe aquí: ! @ mayúsculas" />';

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
      const el = document.getElementById("rd-peer-in");
      if (el) el.focus();
    } else if (type && lastFocus !== "type") {
      lastFocus = "type";
      const el = document.getElementById("rd-kbd-in");
      if (el) el.focus();
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
    if (!c) return;
    if (ch === "\n") {
      c.inputKey("VK_RETURN", true, true, false, false, false, false);
      return;
    }
    if (typeof c.inputString === "function") {
      c.inputString(ch);
      return;
    }
    c.inputKey(ch, false, true, false, false, false, false);
  }

  document.addEventListener("click", function (e) {
    if (e.target && e.target.id === "rd-peer-go") sendPeer();
    if (e.target && e.target.id === "rd-wc-logout") {
      e.preventDefault();
      window.rdPortalLogout && window.rdPortalLogout();
    }
  });
  document.addEventListener("keydown", function (e) {
    if (e.target && e.target.id === "rd-peer-in" && e.key === "Enter") {
      e.preventDefault();
      sendPeer();
    }
    if (e.target && e.target.id === "rd-kbd-in") {
      if (e.key === "Enter") {
        e.preventDefault();
        sendChar("\n");
        e.target.value = "";
      } else if (e.key === "Backspace" && !e.target.value) {
        e.preventDefault();
        const c = conn();
        if (c) c.inputKey("VK_BACK", true, true, false, false, false, false);
      }
    }
  });
  document.addEventListener("input", function (e) {
    if (!e.target || e.target.id !== "rd-kbd-in") return;
    const v = e.target.value;
    e.target.value = "";
    for (let i = 0; i < v.length; i++) sendChar(v[i]);
  });
})();
