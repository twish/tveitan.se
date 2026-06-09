# CLAUDE.md

## Project identity — read first

This is a **private, personal hobby project**, not a Mediaflow project.

Although the git user may be configured with a `@mediaflow.com` email globally, **this repository has no connection to Mediaflow or any employer**. Do not treat it as work, do not reference internal Mediaflow tooling/processes, and do not apply any company conventions here. It is the personal property of Johannes Tveitan (GitHub: [twish](https://github.com/twish)).

## What this is

`tveitan.se` — Johannes Tveitan's personal homepage. A tiny Go web server that
renders markdown to HTML on demand and caches it, in a synthwave / ascii-art
style. Content is markdown, the voice is agent-authored, the look and tone are
both tweakable knobs.

## Architecture

- `main.go` — the server: renders + caches, keyed by a content+theme version.
  Editing a file changes the version and busts the cache on the next request,
  so content goes live with no rebuild and no restart.
- `pkg/content/` — generic engine: `Source` interface + `FSSource` (folder of
  `*.md`), `Doc`/`Section`/`Sticker`, `Entries`. The pluggable backend seam (a
  future `LifthrasirSource` implements `Source`).
- `pkg/render/` — generic engine: goldmark (md→html), cache/version + theme
  injection, fingerprinted CSS, and the `HeadingRenderer`/`NavRenderer`/
  `FooterRenderer` interfaces + `WithExtraPages`. Knows nothing about ascii.
- `internal/site/` — this site's concrete look: `AsciiHeading` (figlet + the
  synthwave style system + fonts), `UnixNav`, `Footer`, the `/styles` gallery.
  Wired into the engine in `main.go` via `render.With…` options.
- `theme/` — `layout.html` + `synthwave.css`, edited live (no rebuild). The look knob.
- `content/` — all pages as markdown + frontmatter. `index.md` is the homepage.
- `content/site.yaml` — live-read site config toggling the machine-friendly
  artifacts: `/llms.txt`, raw `/<slug>.md`, `/sitemap.xml`, `/robots.txt`.
- `.claude/skills/` — `voice.md` (the words knob) + `new-post` / `asciify`
  skills that author content in that voice.

## Hosting

Deployed as a Docker container behind a reverse proxy with automatic TLS.
`docker-compose.yml` runs the container, `caddy/tveitan.caddy` is the proxy
snippet, and `./deploy` (content+theme) / `./deploy --build` (code) publish it.
Content + theme are bind-mounted so server-side edits are live.

Operational specifics (host, paths, proxy config, DNS, deploy targets) are kept
out of the public repo in `.claude/infra.local.md` (gitignored), imported below.

@.claude/infra.local.md

## Git conventions

- Repo-local identity is `tveitan <johannes@tveitan.se>` (personal, not the work email). Already set in this repo's `.git/config`.
- **Do not** add `Co-Authored-By: Claude`, `Generated with Claude Code`, or any similar AI attribution trailers to commit messages or PR bodies.
- Remote: `git@github.com:twish/tveitan.se.git`.

## Working notes

- This is a personal learning project — explanatory, well-commented changes are welcome over terse "production" style.
- Keep it small. The `content.Source` interface is the only abstraction needed on day one; resist adding more until a real second source (lifthrasir) lands.
- Run locally: `go run .` then open http://localhost:8080. Override with `ADDR`, `CONTENT_DIR`, `THEME_DIR` env vars.
