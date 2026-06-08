---
name: stickers
description: Add cassette-futurism "stickers" (margin notes, labels, snippets, images) to a tveitan.se page. Use when the user wants to add stickers, margin notes, pinned images, or HUD-style asides to a blog/content page.
---

# stickers

Stickers are small elements pinned in the margins beside a page's content —
note, label, snippet, or image — styled like cassette-futurism HUD labels with
yellow-black hazard tape.

## The one rule: stickers are content, never decoration

Every sticker must be **relevant to the page it's on**. A note expands on
something in the post; a snippet pulls a real quote or line from the body; a
label marks something true; an image actually relates to the subject. Never add
stickers just to fill the gutters or for visual texture. If you can't make it
relevant, don't add it.

Aim for **1–4 per page**, fewer is fine. A short page may want none.

## Authoring

Add a `stickers` list to the page's frontmatter:

```yaml
stickers:
  - {type: image,   src: /media/thing.png, text: "caption", side: right, at: 12%, rotate: -3}
  - {type: note,    text: "a real aside about the post", side: left,  at: 40%, rotate: 2}
  - {type: snippet, text: "a line pulled from the body", side: right, at: 60%}
  - {type: label,   text: "// status", side: left, at: 8%}
```

Fields:
- `type` — `note` | `label` | `snippet` | `image`
- `text` — body for note/label/snippet; caption for image (supports `\n`)
- `src` — image path for `image`; put files in `content/media/`, reference as `/media/<file>`
- `side` — `left` | `right` (which gutter); default right
- `at` — vertical position down the article, e.g. `25%` or `25`; default `20%`
- `rotate` — tilt in degrees, e.g. `-3`; keep it subtle (±5)

## Layout notes

- On wide screens stickers sit in the gutters beside the column; below 1200px
  they collapse into a stacked row at the end of the article. Design for both.
- Stagger `at` values so stickers on the same side don't overlap.
- Images are auto-tinted to fit the palette — pick images that read at ~200px wide.
- Don't deploy; the running server renders on the next request.
