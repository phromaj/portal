// Command portal serves the homelab start page: a link dashboard with
// live service status. Zero third-party dependencies, stdlib only.
//
// Visual system: Linear dark-minimal (near-black canvas, Inter, indigo
// accent reserved for interactive states) + Raycast command-bar idiom
// (Cmd+K / "/" palette, kbd hints, arrow-key navigation).
// App logos in icons/*.svg are untouched brand assets.
package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

//go:embed icons/*.svg
var iconFS embed.FS

type Link struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Href  string `json:"href"`
	Logo  string `json:"logo"`  // icons/<logo>.svg
	Check string `json:"check"` // internal health URL; "" = no check (grey dot)
}

type Group struct {
	Title string `json:"title"`
	Links []Link `json:"links"`
}

// The links, read at start from the JSON file named by PORTAL_CONFIG:
// [{"title": "Media", "links": [{"id": "plex", "name": "Plex", "desc": "…",
//   "href": "https://…", "logo": "plex", "check": "http://plex:32400/identity"}]}]
var groups []Group

func loadGroups(path string) ([]Group, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gs []Group
	if err := json.Unmarshal(data, &gs); err != nil {
		return nil, err
	}
	return gs, nil
}

const page = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex, nofollow">
<meta name="theme-color" content="#08090a">
<title>Homelab</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<style>
:root {
  color-scheme: dark;
  --bg: #08090a;
  --panel: #0f1011;
  --surface: rgba(255,255,255,.02);
  --surface-hover: rgba(255,255,255,.045);
  --border: rgba(255,255,255,.08);
  --border-subtle: rgba(255,255,255,.05);
  --text: #f7f8f8;
  --text-2: #d0d6e0;
  --text-3: #8a8f98;
  --text-4: #62666d;
  --accent: #7170ff;
  --accent-dim: rgba(113,112,255,.14);
  --up: #27a644;
  --down: #ff6363;
  --mono: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
}
* { box-sizing: border-box; }
body {
  margin: 0; background: var(--bg); color: var(--text);
  font-family: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif;
  font-feature-settings: "cv01", "ss03";
  font-size: 15px; line-height: 1.6; letter-spacing: -0.011em;
  min-height: 100vh; -webkit-font-smoothing: antialiased;
}
.glow { position: fixed; inset: 0; z-index: -1; pointer-events: none; background:
  radial-gradient(640px 320px at 50% -80px, rgba(94,106,210,.13), transparent 70%), var(--bg); }
.skip { position: absolute; inset-inline-start: 12px; top: -56px; z-index: 10; padding: 8px 14px; border-radius: 6px; background: var(--accent); color: #fff; font-weight: 500; font-size: 13px; text-decoration: none; }
.skip:focus-visible { top: 12px; }
/* ---- top nav (Linear) ---- */
nav { border-bottom: 1px solid var(--border-subtle); background: rgba(8,9,10,.8); }
nav .inner { max-width: 960px; margin: 0 auto; padding: 0 24px; height: 52px; display: flex; align-items: center; gap: 12px; }
.brand { font-size: 13px; font-weight: 500; color: var(--text-2); letter-spacing: .01em; display: flex; align-items: center; gap: 8px; text-decoration: none; }
.brand-mark { width: 16px; height: 16px; border-radius: 4px; background: linear-gradient(135deg, #5e6ad2, #7170ff); flex: none; }
.pill { font-size: 11px; font-weight: 500; color: var(--text-3); border: 1px solid var(--border-subtle); border-radius: 9999px; padding: 2px 10px; display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; }
nav time { margin-inline-start: auto; font-family: var(--mono); font-size: 12px; color: var(--text-4); font-variant-numeric: tabular-nums; }
/* ---- command palette (Raycast) ---- */
main { max-width: 960px; margin: 0 auto; padding: 56px 24px 64px; }
.hero h1 { margin: 0; font-size: clamp(2rem, 5vw, 3rem); font-weight: 500; line-height: 1.05; letter-spacing: -0.022em; color: var(--text); text-wrap: balance; }
.hero p { margin: 12px 0 0; color: var(--text-3); font-size: 15px; }
.palette { margin-top: 28px; display: flex; align-items: center; gap: 10px; background: var(--panel); border: 1px solid var(--border); border-radius: 12px; padding: 6px 8px 6px 14px; box-shadow: rgba(0,0,0,.4) 0px 8px 24px -8px; }
.palette:focus-within { border-color: rgba(113,112,255,.5); box-shadow: rgba(0,0,0,.4) 0px 8px 24px -8px, 0 0 0 3px var(--accent-dim); }
.palette svg { flex: none; color: var(--text-4); }
.palette input { flex: 1; min-width: 0; background: transparent; border: 0; outline: none; color: var(--text); font: inherit; font-size: 15px; padding: 10px 0; }
.palette input::placeholder { color: var(--text-4); }
.palette button { border: 1px solid var(--border); background: rgba(255,255,255,.03); color: var(--text-2); font: inherit; font-size: 13px; font-weight: 500; border-radius: 6px; padding: 8px 14px; min-height: 36px; cursor: pointer; white-space: nowrap; }
kbd { font-family: var(--mono); font-size: 11px; color: var(--text-3); background: linear-gradient(#121212, #0d0d0d); border: 1px solid var(--border-subtle); border-bottom-width: 2px; border-radius: 5px; padding: 2px 6px; white-space: nowrap; }
.hint { display: flex; gap: 12px; margin-top: 10px; font-size: 12px; color: var(--text-4); }
/* ---- groups & cards ---- */
section { margin-top: 40px; }
section h2 { display: flex; align-items: baseline; gap: 8px; font-size: 12px; font-weight: 500; color: var(--text-3); letter-spacing: .08em; text-transform: uppercase; margin: 0 0 12px; }
section h2 .n { font-family: var(--mono); font-size: 11px; color: var(--text-4); }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 8px; }
a.card { display: flex; gap: 12px; align-items: center; padding: 12px 14px; border-radius: 8px; background: var(--surface); border: 1px solid var(--border-subtle); color: var(--text); text-decoration: none; min-height: 64px; transition-property: background-color, border-color, box-shadow; transition-duration: .15s; transition-timing-function: ease-out; }
@media (hover: hover) {
  a.card:hover { background: var(--surface-hover); border-color: var(--border); }
  .palette button:hover { background: rgba(255,255,255,.06); color: var(--text); }
}
a.card.sel { background: var(--surface-hover); border-color: rgba(113,112,255,.55); box-shadow: 0 0 0 3px var(--accent-dim); }
a.card:focus-visible, .palette button:focus-visible, .brand:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.badge { flex: none; width: 40px; height: 40px; display: grid; place-items: center; border-radius: 8px; background: rgba(255,255,255,.04); border: 1px solid var(--border-subtle); }
.badge img { width: 24px; height: 24px; }
a.card .t { font-weight: 500; font-size: 14px; letter-spacing: -0.006em; display: flex; align-items: center; gap: 8px; color: var(--text); }
a.card .d { color: var(--text-3); font-size: 13px; line-height: 1.45; margin-top: 1px; text-wrap: pretty; }
.dot { width: 8px; height: 8px; border-radius: 50%; background: #3a3d42; flex: none; }
.dot.up { background: var(--up); }
.dot.down { background: var(--down); }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; border: 0; }
footer { display: flex; flex-wrap: wrap; gap: 16px; margin-top: 56px; padding-top: 20px; border-top: 1px solid var(--border-subtle); font-size: 12px; color: var(--text-4); }
footer .k { display: inline-flex; align-items: center; gap: 6px; }
@media (prefers-reduced-motion: reduce) {
  a.card, .palette button { transition: none; }
}
@media (max-width: 600px) {
  main { padding: 36px 16px 48px; }
  nav .inner { padding: 0 16px; }
  nav time { display: none; }
  .hint { display: none; }
}
</style>
</head>
<body>
<a class="skip" href="#contenu">Aller au contenu</a>
<div class="glow" aria-hidden="true"></div>
<nav><div class="inner">
<a class="brand" href="/"><span class="brand-mark" aria-hidden="true"></span>Homelab</a>
<span class="pill"><span class="dot up" id="dot-all" aria-hidden="true"></span><span id="count">14 services</span></span>
<time id="clock"></time>
</div></nav>
<main id="contenu">
<div class="hero">
<h1>ホームラボ</h1>
<form class="palette" id="palette" action="https://duckduckgo.com/" method="get" target="_blank" role="search">
<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><circle cx="7" cy="7" r="5"/><path d="M11 11l3.5 3.5"/></svg>
<input type="search" id="q" name="q" placeholder="Rechercher sur le web ou naviguer aux flèches…" aria-label="Rechercher" autocomplete="off">
<kbd>/</kbd><kbd>⌘K</kbd>
<button type="submit">Rechercher</button>
</form>
<div class="hint"><span><kbd>↑</kbd><kbd>↓</kbd> naviguer</span><span><kbd>↵</kbd> ouvrir</span><span><kbd>esc</kbd> quitter</span></div>
</div>
{{range .Groups}}
<section>
<h2>{{.Title}} <span class="n">{{len .Links}}</span></h2>
<div class="grid">
{{range .Links}}<a class="card" href="{{.Href}}" target="_blank" rel="noopener">
<span class="badge"><img src="/icons/{{.Logo}}.svg" alt="" width="24" height="24" loading="lazy"></span>
<span><span class="t">{{.Name}}<span class="dot" id="dot-{{.ID}}" aria-hidden="true"></span><span class="sr-only" id="st-{{.ID}}"></span></span><span class="d">{{.Desc}}</span></span>
</a>{{end}}
</div>
</section>
{{end}}
<footer>
<span class="k"><span class="dot up" aria-hidden="true"></span>en ligne</span>
<span class="k"><span class="dot down" aria-hidden="true"></span>hors ligne</span>
<span class="k"><span class="dot" aria-hidden="true"></span>lien externe</span>
</footer>
</main>
<script>
var clock = document.getElementById("clock");
function tick() { try { clock.textContent = new Date().toLocaleString("fr-FR", {dateStyle: "full", timeStyle: "short"}); } catch (e) { clock.textContent = ""; } }
tick(); setInterval(tick, 30000);
function refresh() {
  fetch("api/status", {cache: "no-store"}).then(function(r) { return r.json(); }).then(function(s) {
    var up = 0, n = 0;
    Object.keys(s).forEach(function(id) {
      var d = document.getElementById("dot-" + id), t = document.getElementById("st-" + id);
      if (!d) return;
      n++; if (s[id].up) up++;
      d.className = "dot " + (s[id].up ? "up" : "down");
      if (t) t.textContent = s[id].up ? " — en ligne" : " — hors ligne";
    });
    var c = document.getElementById("count");
    if (c && n) c.textContent = up + "/" + n + " en ligne";
  }).catch(function() {});
}
refresh(); setInterval(refresh, 30000);
/* Raycast-style keyboard nav */
var q = document.getElementById("q");
var cards = Array.prototype.slice.call(document.querySelectorAll("a.card"));
var sel = -1;
function select(i) {
  if (sel >= 0 && cards[sel]) cards[sel].classList.remove("sel");
  sel = ((i % cards.length) + cards.length) % cards.length;
  cards[sel].classList.add("sel");
  cards[sel].scrollIntoView({block: "nearest"});
}
document.addEventListener("keydown", function(e) {
  var a = document.activeElement, tag = a && a.tagName ? a.tagName : "";
  var typing = tag === "INPUT" || tag === "TEXTAREA";
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") { e.preventDefault(); q.focus(); q.select(); }
  else if (e.key === "/" && !typing) { e.preventDefault(); q.focus(); }
  else if (e.key === "Escape") { if (typing) a.blur(); else if (sel >= 0 && cards[sel]) { cards[sel].classList.remove("sel"); sel = -1; } }
  else if (!typing && (e.key === "ArrowDown" || e.key === "ArrowRight")) { e.preventDefault(); select(sel + 1); }
  else if (!typing && (e.key === "ArrowUp" || e.key === "ArrowLeft")) { e.preventDefault(); select(sel - 1); }
  else if (!typing && e.key === "Enter" && sel >= 0 && cards[sel]) { cards[sel].click(); }
});
</script>
</body>
</html>`

var tmpl = template.Must(template.New("page").Parse(page))

type pageData struct {
	Groups []Group
	Year   int
}

func secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

type status struct {
	Up bool `json:"up"`
	Ms int  `json:"ms"`
}

var httpClient = &http.Client{Timeout: 2 * time.Second}

func probe(url string) status {
	start := time.Now()
	resp, err := httpClient.Get(url)
	if err != nil {
		return status{Up: false, Ms: -1}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return status{Up: false, Ms: -1}
	}
	return status{Up: true, Ms: int(time.Since(start).Milliseconds())}
}

var (
	cacheMu   sync.Mutex
	cacheAt   time.Time
	cacheData map[string]status
)

func allStatus() map[string]status {
	cacheMu.Lock()
	if time.Since(cacheAt) < 20*time.Second && cacheData != nil {
		defer cacheMu.Unlock()
		return cacheData
	}
	cacheMu.Unlock()

	type res struct {
		id string
		st status
	}
	var wg sync.WaitGroup
	out := make(chan res)
	n := 0
	for _, g := range groups {
		for _, l := range g.Links {
			if l.Check == "" {
				continue
			}
			n++
			wg.Add(1)
			go func(id, url string) {
				defer wg.Done()
				out <- res{id, probe(url)}
			}(l.ID, l.Check)
		}
	}
	go func() { wg.Wait(); close(out) }()
	m := make(map[string]status, n)
	for r := range out {
		m[r.id] = r.st
	}

	cacheMu.Lock()
	cacheData, cacheAt = m, time.Now()
	cacheMu.Unlock()
	return m
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	var err error
	if groups, err = loadGroups(os.Getenv("PORTAL_CONFIG")); err != nil {
		log.Fatalf("PORTAL_CONFIG: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = tmpl.Execute(w, pageData{Groups: groups, Year: time.Now().Year()})
	})
	mux.HandleFunc("/icons/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/icons/"):]
		if len(name) == 0 || len(name) > 32 {
			http.NotFound(w, r)
			return
		}
		for _, c := range name {
			if !(c >= 'a' && c <= 'z' || c == '-' || c == '.' || c >= '0' && c <= '9') {
				http.NotFound(w, r)
				return
			}
		}
		b, err := iconFS.ReadFile("icons/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(b)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(allStatus())
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           secure(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("portal listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}
