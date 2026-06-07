# Voice

The tonality and expressiveness of everything published on tveitan.se. Edit
this file to reshape how the site sounds — the `new-post` and `asciify` skills
read it. This is the words knob; `theme/synthwave.css` is the look knob.

## Who's speaking

Johannes Tveitan — software builder, occasional keyboard designer. Writes for
other engineers and curious people, not for recruiters or SEO.

## Default register

- **Direct and dry.** Short declaratives. Say the thing, then stop.
- **Technical but unshowy.** Name the real tool, the real tradeoff. No buzzwords,
  no "leverage", no "in today's fast-paced world".
- **A little neon.** This is a synthwave site — one vivid image per piece is
  welcome (static, signal, horizon, grid). One. Not a fog of metaphor.
- **First person, lowercase ease.** Contractions fine. Em dashes fine. It reads
  like a sharp person talking, not a press release.

## Hard rules

- No marketing voice, no hype, no exclamation-point energy.
- No hedging throat-clearing ("I think maybe it could be argued").
- Don't explain the obvious. Respect the reader's competence.
- Code examples must be real and runnable, never pseudo-filler.
- Length follows substance — a three-line post is fine if three lines is the idea.

## Knobs to tweak

Change these lines and the voice moves with them:

- **Warmth:** `dry` ↔ `warm`. Currently *dry*.
- **Density:** `terse` ↔ `expansive`. Currently *terse*.
- **Neon level:** how much synthwave imagery bleeds into the prose. Currently *low — one image max*.
- **Formality:** `casual` ↔ `formal`. Currently *casual*.

## Frontmatter contract

Every post is a markdown file under `content/` with:

```yaml
---
title: Short, lowercase-ish, no clickbait
date: YYYY-MM-DD
order: <int>   # nav/sort weight; lower first
banner: |      # optional verbatim ascii; omit to figlet the title
---
```
