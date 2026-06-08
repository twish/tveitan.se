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

Pages can carry **stickers** — cassette-futurism margin elements (note, label,
snippet, image) pinned in the gutters, styled with yellow-black hazard tape.
They must be content-relevant, never decoration. Add via frontmatter; images go
in `content/media/` (served at `/media`). See `.claude/skills/stickers`.

Or let an agent write it: the `new-post` skill drafts in the voice defined in
`.claude/skills/voice.md`, `asciify` makes ascii banners, and `stickers` adds
margin elements. Tweak `voice.md` for tone, `theme/synthwave.css` for look.

## Sections

A top-level directory becomes a **section** by adding `content/<dir>/_index.md`.
Its frontmatter controls navigation and how it surfaces:

```yaml
---
title: posts
nav: true          # show in the top nav (~ / posts / projects)
order: 1
sort: date         # list entries newest-first (or "order")
style: doom-sunset # optional banner style for the listing page
home:
  mode: latest     # latest | pinned | none — how it shows on the start page
  count: 3
  # pinned: [projects/kbmkr]   # when mode: pinned
---
intro markdown shown above the listing
```

- `/posts` is a generated listing of that folder's entries; individual entries
  stay out of the nav, so it stays short as posts grow.
- The start page auto-composes a block per section (latest N or a pinned set).

## Layout

```
main.go               server: render + cache (version-keyed)
pkg/content/          Source interface + FSSource (folder of *.md)
pkg/render/           md→html + heading/nav/footer renderers + theme
theme/                layout.html + synthwave.css  (the look)
content/              the markdown                 (the words)
.claude/skills/       voice.md + new-post + asciify (the author)
caddy/tveitan.caddy   reverse-proxy site snippet
docker-compose.yml    runs the container behind the reverse proxy
deploy                ./deploy  (content+theme) | ./deploy --build (image)
```

## Reusing the engine

`pkg/render` and `pkg/content` are the generic engine — they know nothing about
ascii, figlet, or synthwave. This site's concrete look lives in `internal/site`
(ascii heading, unix nav, footer, style system, the `/styles` gallery) and is
wired into the engine from `main.go`. To reskin, implement an interface and pass
the matching option to `render.New`; the engine falls back to plain defaults for
anything you don't provide:

```go
type MyNav struct{}
func (MyNav) Nav(cur string, docs []content.Doc, secs []content.Section) template.HTML {
    return template.HTML(`<nav>…my markup…</nav>`)
}

r, _ := render.New("theme", render.WithNav(MyNav{}))
// also: render.WithHeading(…), render.WithFooter(…), render.WithExtraPages(…)
```

| Interface | Engine default | This site (`internal/site`) |
|---|---|---|
| `HeadingRenderer` | plain `<h1>` | ascii banner + synthwave styles |
| `NavRenderer` | none | unix breadcrumb + `ls` |
| `FooterRenderer` | none | brand + year |

Background and the content-area wrapper are theme concerns (`theme/`), not Go
seams — swap the theme to change them.

## Deploy

Runs as a container behind a reverse proxy with automatic TLS.

```sh
./deploy           # sync content + theme (live, no rebuild)
./deploy --build   # rebuild image + restart (code/dep changes)
```

First-time setup: install the `caddy/tveitan.caddy` snippet into the proxy's
config, point DNS at the host, and reload the proxy (it provisions TLS). The
host/path specifics live in a local, untracked notes file.
