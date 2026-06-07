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
- `internal/content/` — `Source` interface + `FSSource` (folder of `*.md`). The
  pluggable seam: a future `LifthrasirSource` can fetch documents from the
  lifthrasir API by implementing the same interface.
- `internal/render/` — goldmark (md→html), figlet4go (title→ascii banner),
  theme injection, fingerprinted CSS URL for cache-busting.
- `theme/` — `layout.html` + `synthwave.css`, edited live (no rebuild). The look knob.
- `content/` — all pages as markdown + frontmatter. `index.md` is the homepage.
- `.claude/skills/` — `voice.md` (the words knob) + `new-post` / `asciify`
  skills that author content in that voice.

## Hosting

VPS, ssh host `T` (`46.246.48.99`). Runs behind a shared Caddy-in-Docker proxy
on the external `web` network — same pattern as `tidder`. `caddy/tveitan.caddy`
is the proxy snippet; `docker-compose.yml` runs the container; `./deploy` syncs.
Content + theme are bind-mounted so VPS-side edits are live.

## Git conventions

- Repo-local identity is `tveitan <johannes@tveitan.se>` (personal, not the work email). Already set in this repo's `.git/config`.
- **Do not** add `Co-Authored-By: Claude`, `Generated with Claude Code`, or any similar AI attribution trailers to commit messages or PR bodies.
- Remote: `git@github.com:twish/tveitan.se.git`.

## Working notes

- This is a personal learning project — explanatory, well-commented changes are welcome over terse "production" style.
- Keep it small. The `content.Source` interface is the only abstraction needed on day one; resist adding more until a real second source (lifthrasir) lands.
- Run locally: `go run .` then open http://localhost:8080. Override with `ADDR`, `CONTENT_DIR`, `THEME_DIR` env vars.
