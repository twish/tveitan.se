# tveitan.se

Personal homepage. A tiny Go server that renders markdown to HTML on demand,
caches it, and serves it in a synthwave / ascii-art style. No build step, no
database — drop a markdown file in `content/`, it's live on the next request.

## Run

```sh
go run .
# open http://localhost:8080
```

Env overrides: `ADDR` (`:8080`), `CONTENT_DIR` (`./content`), `THEME_DIR` (`./theme`).

## Write

Content is markdown under `content/` with frontmatter:

```yaml
---
title: My post
date: 2026-06-07
order: 1
banner: |        # optional verbatim ascii; omit to figlet the title
---
```

`style: <0-59>` freezes the heading look (figlet font + synthwave palette +
placement); omit it and a stable style is derived from the slug. Browse all 60
at `/styles`.

Or let an agent write it: the `new-post` skill drafts in the voice defined in
`.claude/skills/voice.md`, and `asciify` makes ascii banners. Tweak `voice.md`
for tone, `theme/synthwave.css` for look.

## Layout

```
main.go               server: render + cache (version-keyed)
internal/content/     Source interface + FSSource (folder of *.md)
internal/render/      goldmark + figlet ascii banner + theme
theme/                layout.html + synthwave.css  (the look)
content/              the markdown                 (the words)
.claude/skills/       voice.md + new-post + asciify (the author)
caddy/tveitan.caddy   reverse-proxy snippet for the VPS
docker-compose.yml    runs behind the shared Caddy on host T
deploy                ./deploy  (content+theme) | ./deploy --build (image)
```

## Deploy

Hosted on a VPS (ssh host `T`) behind a shared Caddy proxy.

```sh
./deploy           # sync content + theme (live, no rebuild)
./deploy --build   # rebuild image + restart (code/dep changes)
```

First-time setup: install `caddy/tveitan.caddy` into `/opt/caddy/config/` on the
VPS, point DNS for `tveitan.se` at the host, reload Caddy (auto-TLS).
