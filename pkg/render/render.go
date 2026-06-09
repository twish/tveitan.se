// Package render is a small site engine: it turns markdown docs + a theme into
// cached, ready-to-serve HTML pages. It is content- and look-agnostic — the
// heading, nav, and footer are pluggable (see HeadingRenderer, NavRenderer,
// FooterRenderer), and extra synthetic pages can be injected with WithExtraPages.
// A concrete site wires its own renderers via New's options.
package render

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/twish/tveitan.se/pkg/content"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// HeadingRenderer produces the markup for a page's heading. Swap point: return
// ascii art, a plain <h1>, a hero image — whatever the site wants.
type HeadingRenderer interface {
	Heading(doc content.Doc) template.HTML
}

// NavRenderer turns the content tree + current location into navigation markup.
type NavRenderer interface {
	Nav(current string, docs []content.Doc, sections []content.Section) template.HTML
}

// FooterRenderer produces the page footer from the content tree.
type FooterRenderer interface {
	Footer(docs []content.Doc, sections []content.Section) template.HTML
}

// ExtraPage is a synthetic page the site injects (e.g. a generated index or
// gallery). It's rendered through the layout like any doc, so it gets the same
// heading/nav/footer/theme.
type ExtraPage struct {
	Doc  content.Doc
	Body template.HTML
}

// Built is one fully rendered snapshot of the site for a given content+theme
// version. Pages are keyed by slug. Everything here is immutable once returned.
type Built struct {
	Pages   map[string]string
	Files   map[string]File // machine-friendly artifacts (llms.txt, *.md, sitemap…), keyed by URL path
	CSS     []byte
	CSSHref string
}

type pageData struct {
	Title   string
	Heading template.HTML
	Nav     template.HTML
	Body    template.HTML
	Footer  template.HTML
	CSSHref string
}

// Renderer orchestrates a page: markdown body + pluggable heading/nav/footer.
type Renderer struct {
	themeDir string
	md       goldmark.Markdown
	heading  HeadingRenderer
	nav      NavRenderer
	footer   FooterRenderer
	extra    []ExtraPage
}

// Option overrides a Renderer default.
type Option func(*Renderer)

// WithHeading swaps the heading renderer.
func WithHeading(h HeadingRenderer) Option { return func(r *Renderer) { r.heading = h } }

// WithNav swaps the nav renderer.
func WithNav(n NavRenderer) Option { return func(r *Renderer) { r.nav = n } }

// WithFooter swaps the footer renderer.
func WithFooter(f FooterRenderer) Option { return func(r *Renderer) { r.footer = f } }

// WithExtraPages injects synthetic pages, rendered through the layout.
func WithExtraPages(pages ...ExtraPage) Option {
	return func(r *Renderer) { r.extra = append(r.extra, pages...) }
}

// New builds a renderer reading its theme from themeDir. Defaults are minimal
// (plain heading, no nav/footer); pass With… options to plug in real renderers.
func New(themeDir string, opts ...Option) (*Renderer, error) {
	r := &Renderer{
		themeDir: themeDir,
		md:       goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Typographer)),
		heading:  plainHeading{},
		nav:      plainNav{},
		footer:   plainFooter{},
	}
	for _, o := range opts {
		o(r)
	}
	return r, nil
}

// ThemeVersion hashes the theme files so a theme edit busts the render cache
// the same way a content edit does.
func (r *Renderer) ThemeVersion() (string, error) {
	h := sha256.New()
	for _, name := range []string{"layout.html", "synthwave.css"} {
		info, err := os.Stat(filepath.Join(r.themeDir, name))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s|%d|%d\n", name, info.Size(), info.ModTime().UnixNano())
	}
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// Build renders every doc, section listing, and injected extra page into a full
// HTML page. Callers gate it behind a version check so it only runs on change.
func (r *Renderer) Build(docs []content.Doc, sections []content.Section, cfg SiteConfig) (*Built, error) {
	css, err := os.ReadFile(filepath.Join(r.themeDir, "synthwave.css"))
	if err != nil {
		return nil, fmt.Errorf("read css: %w", err)
	}
	cssHref := "/assets/synthwave." + hex.EncodeToString(sha256Sum(css))[:12] + ".css"

	tmpl, err := template.ParseFiles(filepath.Join(r.themeDir, "layout.html"))
	if err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}

	pages := make(map[string]string, len(docs)+len(sections)+len(r.extra))
	for _, doc := range docs {
		body, err := r.markdown(doc.Body)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", doc.Slug, err)
		}
		// The start page composes the sections' latest/pinned blocks beneath it.
		if doc.Slug == "index" {
			body += template.HTML(homeBlocksHTML(sections, docs))
		}
		page, err := r.page(tmpl, doc, body, cssHref, docs, sections)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", doc.Slug, err)
		}
		pages[doc.Slug] = page
	}

	// One listing page per section at /<slug>: intro + the folder's entries.
	for _, sec := range sections {
		intro, err := r.markdown(sec.Body)
		if err != nil {
			return nil, fmt.Errorf("render section %s: %w", sec.Slug, err)
		}
		body := intro + template.HTML(listingHTML(content.Entries(sec, docs)))
		doc := content.Doc{Slug: sec.Slug, Title: sec.Title, Banner: sec.Banner, Style: sec.Style}
		page, err := r.page(tmpl, doc, body, cssHref, docs, sections)
		if err != nil {
			return nil, fmt.Errorf("render section %s: %w", sec.Slug, err)
		}
		pages[sec.Slug] = page
	}

	for _, ep := range r.extra {
		page, err := r.page(tmpl, ep.Doc, ep.Body, cssHref, docs, sections)
		if err != nil {
			return nil, fmt.Errorf("render extra %s: %w", ep.Doc.Slug, err)
		}
		pages[ep.Doc.Slug] = page
	}

	return &Built{
		Pages:   pages,
		Files:   r.artifacts(docs, sections, cfg),
		CSS:     css,
		CSSHref: cssHref,
	}, nil
}

// page renders one doc through the layout, asking the heading/nav/footer
// renderers for their markup.
func (r *Renderer) page(tmpl *template.Template, doc content.Doc, body template.HTML, cssHref string, docs []content.Doc, sections []content.Section) (string, error) {
	body += template.HTML(stickersHTML(doc.Stickers))
	data := pageData{
		Title:   doc.Title,
		Heading: r.heading.Heading(doc),
		Nav:     r.nav.Nav(doc.Slug, docs, sections),
		Body:    body,
		Footer:  r.footer.Footer(docs, sections),
		CSSHref: cssHref,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *Renderer) markdown(src string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// stickersHTML renders the margin elements. They live inside <article> and are
// absolutely positioned into the gutters on wide screens, collapsing inline on
// narrow ones (see CSS).
func stickersHTML(stk []content.Sticker) string {
	if len(stk) == 0 {
		return ""
	}
	// Order by vertical position so that when they collapse inline on narrow
	// screens they stack in reading order.
	sorted := make([]content.Sticker, len(stk))
	copy(sorted, stk)
	sort.SliceStable(sorted, func(i, j int) bool { return atValue(sorted[i].At) < atValue(sorted[j].At) })

	var b strings.Builder
	b.WriteString(`<div class="stickers">`)
	for _, s := range sorted {
		side := "right"
		if s.Side == "left" {
			side = "left"
		}
		typ := s.Type
		if typ == "" {
			typ = "note"
		}
		// "github" is an image preset (tinted logo + repo link); style it as an image.
		class := typ
		if typ == "github" {
			class = "image"
		}
		size := s.Size
		if size != "sm" && size != "lg" {
			size = "md"
		}
		gap := s.Gap
		if gap == "" {
			gap = "3.5rem"
		}
		linked := ""
		if s.Href != "" {
			linked = " linked"
		}
		fmt.Fprintf(&b, `<aside class="sticker sticker-%s side-%s size-%s%s" style="--at:%s;--rot:%ddeg;--gap:%s">`,
			template.HTMLEscapeString(class), side, size, linked,
			template.HTMLEscapeString(s.At), s.Rotate, template.HTMLEscapeString(gap))
		b.WriteString(stickerInner(s, typ))
		b.WriteString(`</aside>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

// trimScheme drops a leading http(s):// and any trailing slash for display.
func trimScheme(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	return strings.TrimRight(url, "/")
}

// atValue parses the leading number out of an "NN%" position for sorting.
func atValue(at string) int {
	n, _ := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(at), "%"))
	return n
}

func stickerInner(s content.Sticker, typ string) string {
	esc := template.HTMLEscapeString
	var inner string
	switch typ {
	case "github":
		// Reusable preset: the tinted GitHub mark + a caption (the repo, derived
		// from the href if not given). Expects the logo at /media/github.svg.
		cap := s.Text
		if cap == "" {
			cap = trimScheme(s.Href)
		}
		inner = `<figure><img src="/media/github.svg" alt="` + esc(cap) + `" loading="lazy"><figcaption>` + esc(cap) + `</figcaption></figure>`
	case "image":
		cap := ""
		if s.Text != "" {
			cap = `<figcaption>` + esc(s.Text) + `</figcaption>`
		}
		inner = `<figure><img src="` + esc(s.Src) + `" alt="` + esc(s.Text) + `" loading="lazy">` + cap + `</figure>`
	case "label":
		inner = `<span class="s-label">` + esc(s.Text) + `</span>`
	case "snippet":
		inner = `<pre class="s-snip">` + esc(s.Text) + `</pre>`
	default:
		inner = `<p>` + esc(s.Text) + `</p>`
	}
	if s.Href != "" {
		return `<a class="sticker-a" href="` + esc(s.Href) + `">` + inner + `</a>`
	}
	return inner
}

func listingHTML(entries []content.Doc) string {
	esc := template.HTMLEscapeString
	if len(entries) == 0 {
		return `<p class="listing-empty">Nothing here yet.</p>`
	}
	var b strings.Builder
	b.WriteString(`<ul class="listing">`)
	for _, d := range entries {
		date := ""
		if d.Date != "" {
			date = `<time>` + esc(d.Date) + `</time>`
		}
		fmt.Fprintf(&b, `<li><a href="/%s">%s</a>%s</li>`, esc(d.Slug), esc(d.Title), date)
	}
	b.WriteString(`</ul>`)
	return b.String()
}

// homeBlocksHTML composes the start-page section blocks: each section surfaces
// its latest N entries or a pinned set, per its `home` metadata.
func homeBlocksHTML(sections []content.Section, docs []content.Doc) string {
	secs := append([]content.Section(nil), sections...)
	sort.SliceStable(secs, func(i, j int) bool { return secs[i].Order < secs[j].Order })

	var b strings.Builder
	for _, sec := range secs {
		var entries []content.Doc
		switch sec.HomeMode {
		case "latest":
			entries = content.Entries(sec, docs)
			n := sec.HomeCount
			if n <= 0 {
				n = 3
			}
			if len(entries) > n {
				entries = entries[:n]
			}
		case "pinned":
			entries = pickPinned(sec, docs)
		default:
			continue
		}
		if len(entries) == 0 {
			continue
		}
		fmt.Fprintf(&b, `<section class="home-block"><h2><a href="/%s">%s</a></h2>%s</section>`,
			template.HTMLEscapeString(sec.Slug), template.HTMLEscapeString(sec.Title), listingHTML(entries))
	}
	return b.String()
}

func pickPinned(sec content.Section, docs []content.Doc) []content.Doc {
	bySlug := make(map[string]content.Doc, len(docs))
	for _, d := range docs {
		bySlug[d.Slug] = d
	}
	var out []content.Doc
	for _, slug := range sec.HomePinned {
		if d, ok := bySlug[slug]; ok {
			out = append(out, d)
		}
	}
	return out
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
