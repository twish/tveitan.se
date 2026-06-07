---
name: asciify
description: Turn a word or short title into ascii-art for a tveitan.se banner. Use when the user wants an ascii banner, ascii art heading, or a hand-drawn title for a page. Outputs a frontmatter `banner:` block.
---

# asciify

Generate ascii-art for a page banner on tveitan.se.

The server already figlets any title that has no `banner:` set, so only reach
for this when the user wants a *specific* style — box-drawing art like the
homepage, a particular figlet font, or hand-tuned lettering.

## How

1. **Pick the style** with the user if unstated: `figlet` (clean classic),
   `box-drawing` (the `┌┬┐` homepage style), or `block` (`█▀▄`).
2. **Generate it.** Quickest path is figlet on the command line:
   ```sh
   figlet -f standard "text"      # if figlet is installed
   ```
   For box-drawing or block styles, hand-compose or use an online unicode
   ascii generator's output. Keep it ≤ ~72 columns so it fits the banner
   without horizontal scroll on narrow screens.
3. **Emit a frontmatter block,** indented as a YAML literal so trailing spaces
   survive:
   ```yaml
   banner: |
     <line 1>
     <line 2>
   ```
4. **Place it** in the target page's frontmatter, or hand it back for the user
   to paste. Don't strip trailing spaces — they're load-bearing in the art.

## Notes

- The banner renders inside `<pre class="banner">` with a cyan→pink gradient,
  so design for monochrome shape, not color.
- Verbatim banners always win over figlet — setting `banner:` disables the
  automatic title rendering for that page.
