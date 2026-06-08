// Package render turns docs + a theme into ready-to-serve HTML pages.
//
// The theme (layout.html + synthwave.css) lives on disk so it can be tweaked
// without recompiling. The CSS is served under a content-hashed URL so browsers
// cache it forever yet never go stale: change the file, the hash changes, the
// URL changes, the browser refetches.
package render

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mbndr/figlet4go"
	"github.com/twish/tveitan.se/internal/content"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:embed fonts/*.flf
var fontFS embed.FS

// Built is one fully rendered snapshot of the site for a given content+theme
// version. Pages are keyed by slug. Everything here is immutable once returned.
type Built struct {
	Pages   map[string]string
	CSS     []byte
	CSSHref string
}

type pageData struct {
	Title   string
	Heading template.HTML // produced by the HeadingRenderer
	Nav     template.HTML // produced by the NavRenderer
	Body    template.HTML
	CSSHref string
}

// Renderer orchestrates a page: markdown body + a pluggable heading and nav.
// Swap heading/nav to reskin the engine for a different site (see HeadingRenderer,
// NavRenderer).
type Renderer struct {
	themeDir string
	md       goldmark.Markdown
	ascii    *figlet4go.AsciiRender
	heading  HeadingRenderer
	nav      NavRenderer
}

// New builds a renderer that reads its theme from themeDir and loads the
// embedded figlet fonts. It wires this site's defaults: ascii-art headings and
// unix-shell navigation.
func New(themeDir string) (*Renderer, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
	)
	ascii := figlet4go.NewAsciiRender()
	if err := loadFonts(ascii); err != nil {
		return nil, err
	}
	return &Renderer{
		themeDir: themeDir,
		md:       md,
		ascii:    ascii,
		heading:  asciiHeading{ascii: ascii},
		nav:      unixNav{},
	}, nil
}

// loadFonts registers every embedded *.flf under its lowercased file stem, so
// styles can ask for "slant", "doom", etc.
func loadFonts(ar *figlet4go.AsciiRender) error {
	entries, err := fontFS.ReadDir("fonts")
	if err != nil {
		return err
	}
	for _, e := range entries {
		data, err := fontFS.ReadFile("fonts/" + e.Name())
		if err != nil {
			return err
		}
		// Some .flf ship with CRLF; a stray \r corrupts figlet4go's endmark
		// parsing and the glyphs render stacked vertically. Normalize to LF.
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		name := strings.ToLower(strings.TrimSuffix(e.Name(), ".flf"))
		if err := ar.LoadBindataFont(data, name); err != nil {
			return fmt.Errorf("load font %s: %w", name, err)
		}
	}
	return nil
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

// Build renders every doc into a full HTML page. It reads the theme fresh each
// call; callers gate it behind a version check so it only runs when something
// actually changed.
func (r *Renderer) Build(docs []content.Doc, sections []content.Section) (*Built, error) {
	css, err := os.ReadFile(filepath.Join(r.themeDir, "synthwave.css"))
	if err != nil {
		return nil, fmt.Errorf("read css: %w", err)
	}
	cssHash := hex.EncodeToString(sha256Sum(css))[:12]
	cssHref := "/assets/synthwave." + cssHash + ".css"

	tmpl, err := template.ParseFiles(filepath.Join(r.themeDir, "layout.html"))
	if err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}

	pages := make(map[string]string, len(docs)+len(sections)+1)
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
		body := intro + template.HTML(listingHTML(entriesOf(sec, docs)))
		doc := content.Doc{Slug: sec.Slug, Title: sec.Title, Banner: sec.Banner, Style: sec.Style}
		page, err := r.page(tmpl, doc, body, cssHref, docs, sections)
		if err != nil {
			return nil, fmt.Errorf("render section %s: %w", sec.Slug, err)
		}
		pages[sec.Slug] = page
	}

	galleryDoc := content.Doc{Slug: "styles", Title: "styles", Style: "slant-cyan-pink"}
	gallery, err := r.page(tmpl, galleryDoc, r.galleryBody(), cssHref, docs, sections)
	if err != nil {
		return nil, fmt.Errorf("render gallery: %w", err)
	}
	pages["styles"] = gallery

	return &Built{Pages: pages, CSS: css, CSSHref: cssHref}, nil
}

// page renders one doc through the layout, asking the heading and nav renderers
// for their markup.
func (r *Renderer) page(tmpl *template.Template, doc content.Doc, body template.HTML, cssHref string, docs []content.Doc, sections []content.Section) (string, error) {
	body += template.HTML(stickersHTML(doc.Stickers))
	data := pageData{
		Title:   doc.Title,
		Heading: r.heading.Heading(doc),
		Nav:     r.nav.Nav(doc.Slug, docs, sections),
		Body:    body,
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

// galleryBody builds the /styles page: every style rendered with sample text,
// labelled by index and name, so a specific look can be picked by number.
func (r *Renderer) galleryBody() template.HTML {
	const sample = "tveitan"
	var b strings.Builder
	b.WriteString(`<p>Every heading style, frozen. Set <code>style: N</code> (or the name) in a page's frontmatter to pin one; leave it out and it's derived from the slug.</p>`)
	b.WriteString(`<div class="gallery">`)
	for i, st := range styles {
		art := figlet(r.ascii, sample, st.Font)
		placement := st.Align
		if st.Wide {
			placement = "wide"
		}
		fmt.Fprintf(&b,
			`<div class="style-card"><div class="style-meta"><span class="idx">%02d</span> %s · %s</div>`+
				`<pre class="banner" style="--cols:%d;--g1:%s;--g2:%s;--ga:%s">%s</pre></div>`,
			i, template.HTMLEscapeString(st.Name), placement,
			maxLineLen(art), st.Palette.g1, st.Palette.g2, st.Angle, template.HTMLEscapeString(art))
	}
	b.WriteString(`</div>`)
	return template.HTML(b.String())
}

// stickersHTML renders the margin elements. They live inside <article> and are
// absolutely positioned into the gutters on wide screens, collapsing to a
// stacked list on narrow ones (see CSS).
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
		size := s.Size
		if size != "sm" && size != "lg" {
			size = "md"
		}
		gap := s.Gap
		if gap == "" {
			gap = "3.5rem"
		}
		fmt.Fprintf(&b, `<aside class="sticker sticker-%s side-%s size-%s" style="--at:%s;--rot:%ddeg;--gap:%s">`,
			template.HTMLEscapeString(typ), side, size,
			template.HTMLEscapeString(s.At), s.Rotate, template.HTMLEscapeString(gap))
		b.WriteString(stickerInner(s, typ))
		b.WriteString(`</aside>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

// atValue parses the leading number out of an "NN%" position for sorting.
func atValue(at string) int {
	n, _ := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(at), "%"))
	return n
}

func stickerInner(s content.Sticker, typ string) string {
	esc := template.HTMLEscapeString
	switch typ {
	case "image":
		cap := ""
		if s.Text != "" {
			cap = `<figcaption>` + esc(s.Text) + `</figcaption>`
		}
		return `<figure><img src="` + esc(s.Src) + `" alt="` + esc(s.Text) + `" loading="lazy">` + cap + `</figure>`
	case "label":
		return `<span class="s-label">` + esc(s.Text) + `</span>`
	case "snippet":
		return `<pre class="s-snip">` + esc(s.Text) + `</pre>`
	default:
		return `<p>` + esc(s.Text) + `</p>`
	}
}

// entriesOf returns a section's pages (the docs directly under its directory),
// sorted by the section's rule: "date" newest-first, otherwise by order.
func entriesOf(sec content.Section, docs []content.Doc) []content.Doc {
	var out []content.Doc
	for _, d := range docs {
		if d.Section() == sec.Slug {
			out = append(out, d)
		}
	}
	if sec.Sort == "date" {
		sort.SliceStable(out, func(i, j int) bool { return out[i].Date > out[j].Date })
		return out
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Title < out[j].Title
	})
	return out
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
			entries = entriesOf(sec, docs)
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

// maxLineLen returns the widest line in runes, so the template can tell CSS how
// many columns the banner needs and the font can be scaled to fit.
func maxLineLen(s string) int {
	max := 1
	for line := range strings.SplitSeq(s, "\n") {
		if n := len([]rune(line)); n > max {
			max = n
		}
	}
	return max
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
