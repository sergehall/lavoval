package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

type PublicHandler struct {
	page *template.Template
	cfg  config.Config
}

type publicPageData struct {
	AppName    string
	AppEnv     string
	Host       string
	APIVersion string
	HealthURL  string
	RoutesJSON string
}

func NewPublicHandler(cfg config.Config) *PublicHandler {
	return &PublicHandler{
		cfg:  cfg,
		page: template.Must(template.New("public-api-page").Parse(publicAPIPageHTML)),
	}
}

func (h *PublicHandler) Index(w http.ResponseWriter, r *http.Request) {
	routesPayload := map[string]any{
		"service": h.cfg.AppName,
		"env":     h.cfg.AppEnv,
		"routes": map[string]string{
			"livez":   "/livez",
			"healthz": "/healthz",
			"readyz":  "/readyz",
			"metrics": "/metrics",
		},
		"api": map[string]any{
			"base": "/api/v1",
			"namespaces": []string{
				"/api/v1/auth",
				"/api/v1/skills",
				"/api/v1/me",
				"/api/v1/runtime",
				"/api/v1/admin",
			},
		},
	}

	routesJSON, err := json.MarshalIndent(routesPayload, "", "  ")
	if err != nil {
		http.Error(w, "failed to render API info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.page.Execute(w, publicPageData{
		AppName:    h.cfg.AppName,
		AppEnv:     h.cfg.AppEnv,
		Host:       r.Host,
		APIVersion: "v1",
		HealthURL:  "/healthz",
		RoutesJSON: string(routesJSON),
	})
}

const publicAPIPageHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="robots" content="noindex, nofollow" />
    <title>{{ .AppName }} API</title>
    <style>
      :root {
        color-scheme: dark;
        --bg: #07111f;
        --bg-panel: rgba(10, 23, 43, 0.86);
        --bg-panel-strong: rgba(14, 31, 58, 0.95);
        --border: rgba(144, 180, 255, 0.18);
        --text: #ecf4ff;
        --muted: #9db2d0;
        --accent: #6dd3ff;
        --accent-strong: #9ff7d8;
        --shadow: 0 24px 80px rgba(0, 0, 0, 0.45);
      }

      * {
        box-sizing: border-box;
      }

      body {
        margin: 0;
        min-height: 100vh;
        display: grid;
        place-items: center;
        padding: 32px 20px;
        font-family:
          ui-sans-serif,
          -apple-system,
          BlinkMacSystemFont,
          "Segoe UI",
          sans-serif;
        color: var(--text);
        background:
          radial-gradient(circle at top, rgba(109, 211, 255, 0.16), transparent 34%),
          radial-gradient(circle at right, rgba(159, 247, 216, 0.08), transparent 28%),
          linear-gradient(160deg, #040a13 0%, #081425 42%, #0e1830 100%);
      }

      .shell {
        width: min(920px, 100%);
        border: 1px solid var(--border);
        border-radius: 28px;
        background: var(--bg-panel);
        box-shadow: var(--shadow);
        backdrop-filter: blur(18px);
        overflow: hidden;
      }

      .hero {
        padding: 36px 32px 28px;
        border-bottom: 1px solid var(--border);
      }

      .eyebrow {
        display: inline-flex;
        align-items: center;
        gap: 10px;
        padding: 8px 14px;
        border-radius: 999px;
        background: rgba(109, 211, 255, 0.08);
        color: var(--accent);
        letter-spacing: 0.18em;
        font-size: 12px;
        text-transform: uppercase;
      }

      .dot {
        width: 10px;
        height: 10px;
        border-radius: 999px;
        background: #ffc76d;
        box-shadow: 0 0 0 6px rgba(255, 199, 109, 0.12);
      }

      h1 {
        margin: 20px 0 10px;
        font-size: clamp(36px, 6vw, 64px);
        line-height: 0.94;
        letter-spacing: -0.05em;
      }

      .subtitle {
        max-width: 620px;
        margin: 0;
        color: var(--muted);
        font-size: 17px;
        line-height: 1.65;
      }

      .toolbar {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        margin-top: 28px;
      }

      .btn,
      .link-chip {
        appearance: none;
        border: 0;
        cursor: pointer;
        border-radius: 14px;
        padding: 13px 16px;
        font: inherit;
        text-decoration: none;
        transition:
          transform 160ms ease,
          background 160ms ease,
          border-color 160ms ease;
      }

      .btn {
        background: linear-gradient(135deg, #6dd3ff, #9ff7d8);
        color: #04111b;
        font-weight: 700;
      }

      .link-chip {
        border: 1px solid var(--border);
        background: rgba(255, 255, 255, 0.03);
        color: var(--text);
      }

      .btn:hover,
      .link-chip:hover {
        transform: translateY(-1px);
      }

      .content {
        display: grid;
        gap: 20px;
        padding: 28px 32px 32px;
      }

      .status-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
        gap: 14px;
      }

      .card {
        border: 1px solid var(--border);
        border-radius: 20px;
        padding: 18px;
        background: var(--bg-panel-strong);
      }

      .docs-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: 14px;
      }

      .route-list {
        display: grid;
        gap: 10px;
        margin-top: 16px;
      }

      .route-row {
        display: flex;
        justify-content: space-between;
        gap: 16px;
        padding: 12px 14px;
        border: 1px solid var(--border);
        border-radius: 14px;
        background: rgba(255, 255, 255, 0.02);
      }

      .route-row code,
      .namespace-path,
      .code-inline {
        color: var(--accent-strong);
      }

      .route-copy,
      .namespace-copy {
        margin: 6px 0 0;
        color: var(--muted);
        font-size: 14px;
        line-height: 1.55;
      }

      .namespace-card h3 {
        margin: 0 0 8px;
        font-size: 16px;
      }

      .namespace-path {
        display: inline-block;
        margin: 0 0 10px;
        font-size: 13px;
      }

      .example-list {
        display: grid;
        gap: 12px;
        margin-top: 16px;
      }

      .label {
        margin: 0 0 8px;
        color: var(--muted);
        font-size: 12px;
        letter-spacing: 0.14em;
        text-transform: uppercase;
      }

      .value {
        margin: 0;
        font-size: 20px;
        font-weight: 700;
      }

      .value[data-status="ok"] {
        color: var(--accent-strong);
      }

      .panel-title {
        margin: 0 0 12px;
        font-size: 18px;
      }

      .panel-copy {
        margin: 0;
        color: var(--muted);
        line-height: 1.65;
      }

      .json-panel {
        display: none;
        margin-top: 16px;
        border-radius: 20px;
        border: 1px solid var(--border);
        background: #04101e;
        overflow: hidden;
      }

      .json-panel.is-open {
        display: block;
      }

      .json-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 14px 18px;
        border-bottom: 1px solid var(--border);
        color: var(--muted);
        font-size: 13px;
        letter-spacing: 0.08em;
        text-transform: uppercase;
      }

      pre {
        margin: 0;
        padding: 18px;
        overflow: auto;
        color: #d8e8ff;
        font-size: 13px;
        line-height: 1.6;
      }

      .example-block {
        border: 1px solid var(--border);
        border-radius: 16px;
        overflow: hidden;
        background: #04101e;
      }

      .example-label {
        padding: 12px 16px;
        border-bottom: 1px solid var(--border);
        color: var(--muted);
        font-size: 12px;
        letter-spacing: 0.1em;
        text-transform: uppercase;
      }

      .example-block pre {
        padding: 16px;
      }

      .footer {
        display: flex;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 12px;
        padding: 18px 32px 28px;
        border-top: 1px solid var(--border);
        color: var(--muted);
      }

      .footer-meta {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 10px;
      }

      .footer-badge {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 8px 12px;
        border: 1px solid var(--border);
        border-radius: 999px;
        background: rgba(255, 255, 255, 0.03);
        color: var(--text);
        font-size: 12px;
        letter-spacing: 0.08em;
        text-transform: uppercase;
      }

      @media (max-width: 640px) {
        .hero,
        .content,
        .footer {
          padding-inline: 20px;
        }
      }
    </style>
  </head>
  <body>
    <main class="shell">
      <section class="hero">
        <div class="eyebrow">
          <span class="dot" aria-hidden="true"></span>
          Public API Surface
        </div>
        <h1>{{ .AppName }} API</h1>
        <p class="subtitle">
          A minimal public gateway for health, readiness, metrics, and versioned application routes.
          Built for operators, integrations, and quick verification.
        </p>
        <div class="toolbar">
          <button id="info-toggle" class="btn" type="button" aria-expanded="false" aria-controls="api-info-panel">
            Info
          </button>
          <a class="link-chip" href="/healthz">Health</a>
          <a class="link-chip" href="/readyz">Ready</a>
          <a class="link-chip" href="/metrics">Metrics</a>
        </div>
      </section>

      <section class="content">
        <div class="status-grid">
          <article class="card">
            <p class="label">Service</p>
            <p class="value">{{ .AppName }}</p>
          </article>
          <article class="card">
            <p class="label">Environment</p>
            <p class="value">{{ .AppEnv }}</p>
          </article>
          <article class="card">
            <p class="label">Health Endpoint</p>
            <p id="health-status" class="value">Checking…</p>
          </article>
        </div>

        <article class="card">
          <h2 class="panel-title">Usage</h2>
          <p class="panel-copy">
            Use <code>/healthz</code> for effective runtime configuration flags, <code>/readyz</code> for readiness,
            <code>/livez</code> for heartbeat checks, and <code>/api/v1/*</code> for the versioned application surface.
          </p>

          <div class="route-list" aria-label="Core routes">
            <div class="route-row">
              <code>GET /livez</code>
              <span class="route-copy">Simple heartbeat for liveness probes.</span>
            </div>
            <div class="route-row">
              <code>GET /healthz</code>
              <span class="route-copy">Safe runtime flags for env and provider health.</span>
            </div>
            <div class="route-row">
              <code>GET /readyz</code>
              <span class="route-copy">Readiness signal for traffic acceptance.</span>
            </div>
            <div class="route-row">
              <code>GET /metrics</code>
              <span class="route-copy">Prometheus metrics surface.</span>
            </div>
          </div>

          <section id="api-info-panel" class="json-panel" aria-live="polite">
            <div class="json-header">
              <span>Endpoint Index</span>
              <span>JSON</span>
            </div>
            <pre>{{ .RoutesJSON }}</pre>
          </section>
        </article>

        <section class="docs-grid" aria-label="API namespaces">
          <article class="card namespace-card">
            <h3>Authentication</h3>
            <span class="namespace-path">/api/v1/auth</span>
            <p class="namespace-copy">
              Registration, login, OAuth start and complete flows, email verification, password reset,
              and MFA enrollment and sign-in completion.
            </p>
          </article>
          <article class="card namespace-card">
            <h3>Public skills</h3>
            <span class="namespace-path">/api/v1/skills</span>
            <p class="namespace-copy">
              Public catalog listing and skill detail routes for published marketplace offers.
            </p>
          </article>
          <article class="card namespace-card">
            <h3>Account</h3>
            <span class="namespace-path">/api/v1/me</span>
            <p class="namespace-copy">
              Authenticated profile, security settings, and creator-owned skill management.
            </p>
          </article>
          <article class="card namespace-card">
            <h3>Runtime</h3>
            <span class="namespace-path">/api/v1/runtime</span>
            <p class="namespace-copy">
              Skill execution endpoints for starting runs and inspecting run history.
            </p>
          </article>
          <article class="card namespace-card">
            <h3>Admin</h3>
            <span class="namespace-path">/api/v1/admin</span>
            <p class="namespace-copy">
              Governance, moderation, mail operations, runtime oversight, and internal operational tooling.
            </p>
          </article>
        </section>

        <article class="card">
          <h2 class="panel-title">Quick start</h2>
          <p class="panel-copy">
            Start with the health surfaces, then move into the versioned API. For authenticated routes,
            send a bearer token after completing the auth flow under
            <span class="code-inline"> /api/v1/auth</span>.
          </p>

          <div class="example-list">
            <div class="example-block">
              <div class="example-label">Health</div>
              <pre><code>curl https://api.lavoval.com/healthz</code></pre>
            </div>
            <div class="example-block">
              <div class="example-label">Public catalog</div>
              <pre><code>curl https://api.lavoval.com/api/v1/skills</code></pre>
            </div>
            <div class="example-block">
              <div class="example-label">Authenticated request</div>
              <pre><code>curl https://api.lavoval.com/api/v1/me \
  -H "Authorization: Bearer &lt;access-token&gt;"</code></pre>
            </div>
          </div>
        </article>
      </section>

      <footer class="footer">
        <div class="footer-meta">
          <span>Serving public API entrypoint for <strong>{{ .Host }}</strong></span>
        </div>
        <div class="footer-meta">
          <span class="footer-badge">env {{ .AppEnv }}</span>
          <span class="footer-badge">api {{ .APIVersion }}</span>
        </div>
      </footer>
    </main>

    <script>
      const button = document.getElementById('info-toggle');
      const panel = document.getElementById('api-info-panel');
      const healthStatus = document.getElementById('health-status');

      button.addEventListener('click', () => {
        const isOpen = panel.classList.toggle('is-open');
        button.setAttribute('aria-expanded', String(isOpen));
        button.textContent = isOpen ? 'Hide info' : 'Info';
      });

      fetch('{{ .HealthURL }}', { headers: { Accept: 'application/json' } })
        .then((response) => {
          if (!response.ok) {
            throw new Error('health request failed');
          }
          return response.json();
        })
        .then((payload) => {
          const status = payload && payload.data && payload.data.status ? payload.data.status : 'unknown';
          healthStatus.textContent = status.toUpperCase();
          healthStatus.dataset.status = status;
        })
        .catch(() => {
          healthStatus.textContent = 'UNAVAILABLE';
        });
    </script>
  </body>
</html>`
