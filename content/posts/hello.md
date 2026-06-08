---
title: Hello, neon world
date: 2026-06-07
order: 1
stickers:
  - {type: image, src: /media/horizon.svg, text: "synthwave horizon", side: right, at: 10%, rotate: -3}
  - {type: label, text: "// first post", side: left, at: 16%, rotate: 2}
  - {type: note, text: "no build step — drop a markdown file in a folder and it's live on the next request.", side: right, at: 52%, rotate: 2}
  - {type: snippet, text: "neon on black,\nscanlines,\na grid to the horizon", side: left, at: 62%, rotate: -2}
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

