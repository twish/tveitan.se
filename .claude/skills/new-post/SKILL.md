---
name: new-post
description: Draft a new tveitan.se post as a markdown file in the site's voice. Use when the user wants to write, draft, or add a blog post / page to their personal site. Reads voice.md for tonality.
---

# new-post

Create a new post for tveitan.se as a markdown file under `content/`.

## Steps

1. **Read the voice.** Always read `.claude/skills/voice.md` first and write to
   its current register and knobs. If the user's request contradicts the voice,
   follow the user but say so.
2. **Get the idea.** If the user gave a topic, use it. If vague, ask one
   sharp question, not five.
3. **Pick a slug.** Lowercase, hyphenated, no date prefix:
   `content/posts/<slug>.md`. Check it doesn't already exist.
4. **Write the frontmatter** per the contract in voice.md. Set `date` to today,
   `order` to one more than the highest existing post `order`. Pick a random
   `style` in 0–39 and write it in, so the heading look is frozen at creation
   (the user can change the number later — `/styles` shows them all). Omit
   `banner` unless the title deserves hand-drawn ascii — if so, invoke the
   `asciify` skill to generate it.
5. **Write the body.** Markdown. Open with the point, not a warm-up. Length
   follows substance. Real code only.
6. **Don't deploy.** Writing the file is the whole job — the running server
   renders it live on the next request. Mention the local URL
   (`/posts/<slug>`) so the user can preview.

## Don'ts

- Don't invent facts, projects, or quotes about Johannes.
- Don't pad to hit a length. Don't add an SEO meta-description.
- Don't touch the theme or server code — this skill only writes content.
