---
title: tveitan.se
order: 2
summary: this site — a tiny markdown-rendering Go engine wearing synthwave.
stickers:
  - {type: label, text: "// go · markdown · no db", side: left, at: 12%, rotate: -2, size: sm, gap: 4.5rem}
  - {type: snippet, text: "render.New(theme,\n  WithHeading(ascii),\n  WithNav(unix),\n)", side: right, at: 15%, rotate: -3, size: md, gap: 3rem}
  - {type: note, text: "engine lives in pkg/, the synthwave half in internal/ — someone could bolt their own look on top.", side: left, at: 50%, rotate: 2, size: md, gap: 3.5rem}
  - {type: github, href: "https://github.com/twish/tveitan.se", side: right, at: 56%, rotate: -3, size: md, gap: 3.5rem}
---

# tveitan.se

This site. A tiny Go server that reads markdown and spits out HTML — no build
step, no database, no framework doing things behind my back. Headings come out
as ascii art in a synthwave palette, picked at random and then frozen, so every
page keeps its own face.

I rebuilt it mostly as somewhere to tinker. It drifted into being a small
engine: the generic half is reusable, the synthwave nonsense stays mine. An
agent writes most of the words, in a voice I keep in a file and poke at when it
starts sounding like a press release.

[Source on GitHub](https://github.com/twish/tveitan.se).
