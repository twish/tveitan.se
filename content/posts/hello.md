---
title: Hello, neon world
date: 2026-06-07
order: 1
summary: first post — why the rebuilt site is a tiny Go server that renders markdown live.
stickers:
  - {type: image, src: /media/horizon.svg, text: "synthwave horizon", side: right, at: 8%, rotate: -3, size: lg, gap: 3rem}
  - {type: label, text: "// first post", side: left, at: 14%, rotate: 2, size: sm, gap: 5rem}
  - {type: note, text: "no build step — drop a markdown file in a folder and it's live on the next request.", side: right, at: 56%, rotate: 2, size: md, gap: 4.5rem}
  - {type: snippet, text: "neon on black,\nscanlines,\na grid to the horizon", side: left, at: 64%, rotate: -2, size: md, gap: 3rem}
---

# Hello, neon world

First post on the rebuilt site. The whole thing is a tiny Go server that reads
markdown from a folder, renders it, and caches the result until the file
changes. No build step, no database, no fuss.

The look is **synthwave** — neon on black, scanlines, a grid running off toward
the horizon. The words are written in a voice I keep in `voice.md` and tweak
when the mood needs to shift.

```go
fmt.Println("more soon")
```

