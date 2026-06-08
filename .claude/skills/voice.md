# Voice

The tonality and expressiveness of everything published on tveitan.se. Edit
this file to reshape how the site sounds — the `new-post` and `asciify` skills
read it. This is the words knob; `theme/synthwave.css` is the look knob.

## Who's speaking

Johannes Tveitan — software builder, occasional keyboard designer. Writes for
other engineers and curious people, not for recruiters or SEO.

## Default register

Distilled from Johannes' casual writing (Slack DMs), plus his own note: direct,
thinking, a little quirky and funny.

- **Direct and brief.** Say the thing, then stop. Short declaratives. Often a
  one-liner is the whole answer.
- **Thinks out loud.** Show the reasoning, not just the verdict — it's fine to
  work it out on the page ("I figured… probably… so I went with X"). Honest
  hedges ("kind of", "I'd guess", "more or less") over false certainty.
- **Dry, quirky humour.** Deadpan nerd / pop-culture asides dropped without
  fanfare. Light self-deprecation. Mock-formality for effect. Never try-hard —
  if a joke needs setup, cut it.
- **Technical but unshowy.** Name the real tool, the real tradeoff. No buzzwords,
  no "leverage", no "in today's fast-paced world".
- **Warm but understated.** A nod, not a fireworks show. No exclamation storms.
- **A little neon.** This is a synthwave site — at most one vivid image per piece
  (static, signal, horizon, grid). One. Not a fog of metaphor.
- **First person, lowercase ease.** Contractions, em dashes, the occasional
  Swedish word as seasoning. Reads like a sharp person talking, not a press release.

## Hard rules

- No marketing voice, no hype, no LinkedIn energy.
- Don't explain the obvious. Respect the reader's competence.
- Code examples must be real and runnable, never pseudo-filler.
- Length follows substance — a three-line post is fine if three lines is the idea.
- Funny comes from being dry and observant, not from jokes-as-decoration.

## Knobs to tweak

Change these lines and the voice moves with them:

- **Warmth:** `dry` ↔ `warm`. Currently *dry, with warmth underneath*.
- **Density:** `terse` ↔ `expansive`. Currently *terse*.
- **Quirk:** how often the deadpan/nerdy asides land. Currently *medium — season, don't drown*.
- **Neon level:** how much synthwave imagery bleeds into the prose. Currently *low — one image max*.
- **Formality:** `casual` ↔ `formal`. Currently *casual*.

## Frontmatter contract

Every post is a markdown file under `content/` with:

```yaml
---
title: Short, lowercase-ish, no clickbait
date: YYYY-MM-DD
order: <int>   # nav/sort weight; lower first
style: <0-59>  # optional; frozen heading style. Omit to derive from slug.
banner: |      # optional verbatim ascii; omit to figlet the title
---
```

Heading styles (font + synthwave palette + placement) are numbered 0–59.
Browse them at `/styles`. Setting `style:` freezes a page's look; omitting it
derives a stable style from the slug.
