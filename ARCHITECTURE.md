# Architecture

A tiny Go server that renders markdown to HTML on demand, caches it, and serves
it. No build step, no database. Content is markdown; the look (ascii headings,
unix-style nav) is pluggable behind interfaces, so the same engine can be
reskinned for a different site.

## How the site works (runtime)

Render once per content+theme **version**, cache the HTML, serve with an ETag.
Edit a file → the version changes → the cache busts → the change is live on the
next request. No restart.

```
                          REQUEST  GET /posts/hello
                              │
                              ▼
        ┌───────────────────────────────────────────────┐
        │  main  —  http router                          │
        │   /assets/*  → fingerprinted CSS (immutable)   │
        │   /media/*   → static media files              │
        │   /*         → servePage ───────────┐          │
        └─────────────────────────────────────┼──────────┘
                                               ▼
                          ┌─────────────────────────────────────┐
                          │ current(ctx)                        │
                          │  version = contentVer + themeVer    │
                          │  (cheap hash of file stats + theme) │
                          └───────────┬─────────────────────────┘
                              same ver │ changed ver
                          ┌───────────┘         └────────────┐
                          ▼                                  ▼
                  serve from CACHE              ┌─── Build(docs, sections) ──────────┐
                  (map[slug]→html)              │  for each doc:                     │
                          │                     │    body = markdown → html          │
                          ▼                     │    + home blocks (start page)      │
                  ETag → 200 / 304              │    + stickers                      │
                                                │    Heading = HeadingRenderer ──┐   │
                                                │    Nav     = NavRenderer ───┐  │   │
                                                │    → layout template        │  │   │
                                                └─────────────────────────────┼──┼───┘
                                                                              │  │
                            ┌──── PLUGGABLE SEAMS (swap = reskin engine) ─────┘  │
                            │                                                    │
         ContentSource ◄────┤   NavRenderer = unix             HeadingRenderer = ascii
          folder source     │    (~ / posts / hello,            (figlet banner +
          [other backends]  │     ls of the current dir)        style system)
```

Three seams keep the engine generic:

| Seam | This site | Another site could |
|---|---|---|
| `ContentSource` | folder of markdown | fetch from an API / CMS |
| `HeadingRenderer` | ascii-art banner | plain `<h1>` over a hero image |
| `NavRenderer` | unix breadcrumb + `ls` | flat top menu |

## How content authoring works

Everything is markdown under `content/`. A file's frontmatter drives its
title, heading style, stickers, etc. A directory becomes a **section** by adding
an `_index.md`, whose frontmatter controls navigation and how it surfaces.

```
   content/                       frontmatter drives everything
   ├── index.md  ────────────────►  start page (+ auto section blocks)
   ├── 404.md
   ├── media/...  ───────────────►  served under /media/
   │
   ├── posts/                       ┌─ _index.md = SECTION definition ──┐
   │   ├── _index.md ──────────────►│ nav: true   order   sort: date    │
   │   ├── hello.md                 │ home: { mode: latest, count: 3 }  │
   │   └── ...                      └───────────────────────────────────┘
   │                                  → nav entry         ~ / [ posts/ ]
   │                                  → /posts listing page
   │                                  → start-page block (latest N)
   │
   └── projects/
       ├── _index.md   (home: { mode: pinned, pinned: [...] })
       └── ...

   ONE content file:                        AUTHORING (optional helpers):
   ┌─────────────────────────────┐           draft post in a defined voice
   │ ---                         │           generate ascii banners
   │ title: ...                  │           add margin "stickers" (notes/images)
   │ style: <n or name>          │
   │ stickers: [ {…}, {…} ]      │           write file ─► server hashes file stats
   │ ---                         │                       ─► rebuilds ─► LIVE
   │ markdown body...            │           (no build, no database, no restart)
   └─────────────────────────────┘
```

- **Pages** (top-level `content/*.md`) and **sections** (`nav: true`) appear in
  the nav; individual entries inside a section do not, so the nav stays short as
  posts grow.
- **`/<section>`** is a generated listing of that folder's entries, sorted by
  date or order.
- The **start page** auto-composes a block per section — its latest N entries or
  a pinned set — from the section's `home` metadata.

## How deploys work

The repo is the source of record. Content and theme are bind-mounted into the
running container, and the server reads them fresh when the version changes — so
publishing content is just syncing files; only code/dependency changes need a
rebuild.

```
   LOCAL repo
   ┌──────────────────────────────────────────────┐
   │ edit markdown / theme / code                  │
   │ commit ; push ──────────────► git remote      │  (source of record)
   └───────────────┬───────────────────────────────┘
                   │   deploy (content+theme)   deploy --build (code/deps)
                   ▼                             ▼
            rsync md + theme              rsync context + rebuild image
                   │                             │
   ════════════════ SERVER (container host) ════════════════════════════
        deploy dir   (content/ theme/ bind-mounted, read-only)
                   │                             │
                   ▼                             ▼
        ┌───────────────────┐           image rebuilt from registry,
        │  app container     │◄───────── container restarts
        │  Go server :PORT   │
        └─────────┬──────────┘
                  │  internal network  (reverse_proxy app:PORT)
                  ▼
        ┌───────────────────┐
        │  reverse proxy     │           per-site config snippet,
        │  :80 :443  (TLS)   │           automatic TLS certificates
        └─────────┬──────────┘
                  │
                  ▼
              DNS  site → server
                  │
                  ▼
              🌐  https://<site>
```

Two deploy modes — the split that matters:

```
  changed only content/ or theme/    →  deploy          (rsync, fast, live, no rebuild)
  changed code / internal / deps      →  deploy --build  (rebuild image + restart)
```

Because `content/` and `theme/` are bind-mounted and re-read on version change,
syncing the files is enough. The Go code is compiled into the image, so it needs
a rebuild.
