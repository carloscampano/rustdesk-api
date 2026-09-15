# Login: captcha, ban y respuesta al cliente

**Fecha:** 2026-09-15  
**Producción:** `https://rustdesk.campano.cl`

## Dónde aplica

| Superficie | Ruta | Captcha | Ban por IP |
|---|---|---|---|
| Panel admin | `POST /api/admin/login` | Sí, a las **3** fallas | Sí, a las **5** (30 min) |
| App RustDesk / API | `POST /api/login` | **No** | Sí, a las **5** (30 min) |

El captcha **solo existe en el panel** (`/_admin/`). La app de escritorio y el cliente móvil no tienen UI de captcha.

El contador es **por IP**, no por usuario. Comparte la misma ventana (10 min) entre panel y app.

Umbrales (`cmd/apimain.go` / config):

- `CaptchaThreshold`: 3
- `BanThreshold`: 5
- `AttemptsWindow`: 10 minutos
- `BanDuration`: 30 minutos

El ban vive **en memoria** del proceso `apimain`. Un `docker restart rustdesk-api` lo limpia.

## Por qué no hay captcha en `/api/login`

Tras 3 claves malas, el servidor pedía captcha. La app no puede enviarlo. El cliente entonces:

1. Recibía HTTP 200 con `{ "code": 423, "message": "Prohibido." }` (ban global del limiter).
2. Esperaba `{ "type": "access_token", "access_token", "user" }`.
3. Mostraba **Failed, bad response from server**.

Eso le pasó a `jany` desde `38.236.1.179` el 2026-09-14. La clave era correcta; la IP estaba baneada.

## Comportamiento actual

- `/api/login`: no exige captcha. Sigue contando fallos y banea a las 5. Respuesta de error nativa: `{ "error": "..." }` (HTTP 400) o ban HTTP **403** con `error` + `message`.
- Limiter (todas las rutas, incluida heartbeat): si la IP está baneada, HTTP **403** y cuerpo:

```json
{ "code": 403, "message": "Prohibido.", "error": "Prohibido.", "data": null }
```

Así el cliente muestra “Prohibido” y no “bad response”.

- `GET /api/device-group/accessible`: usuarios no admin reciben lista vacía HTTP 200 (antes 400 Permission denied, otro toast de “bad response”).

## Cómo entrar (app)

API server en el cliente: `https://rustdesk.campano.cl` (sin `:21114`; ese puerto no está en WAN).

Panel: `https://rustdesk.campano.cl/_admin/` — ahí sí puede aparecer captcha.

## Código

- `http/controller/admin/login.go` — captcha + ban
- `http/controller/api/login.go` — solo ban
- `http/middleware/limiter.go` — ban global, formato compatible con el cliente
- `http/controller/api/group.go` — device-group vacío para no admin
- `utils/login_limiter.go` — política
