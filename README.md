# portal

A homelab start page: links grouped by section, live status dots, a
command palette (Cmd+K or `/`). Go standard library only.

```sh
PORTAL_CONFIG=links.json PORT=3000 portal
```

`links.json`:

```json
[
  {
    "title": "Media",
    "links": [
      { "id": "plex", "name": "Plex", "desc": "Streaming", "href": "https://app.plex.tv/desktop", "logo": "plex" },
      { "id": "seerr", "name": "Seerr", "desc": "Requests", "href": "https://seerr.example.com", "logo": "seerr", "check": "http://seerr:5055/api/v1/settings/public" }
    ]
  }
]
```

`logo` names an SVG in `icons/`; `check` is an internal URL probed every 20 s
(no `check`: grey dot). Build with `nix build`.
