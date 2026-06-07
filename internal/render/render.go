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
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	"github.com/mbndr/figlet4go"
	"github.com/twish/tveitan.se/internal/content"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Built is one fully rendered snapshot of the site for a given content+theme
// version. Pages are keyed by slug. Everything here is immutable once returned.
type Built struct {
	Pages   map[string]string
	CSS     []byte
	CSSHref string
}

type navItem struct {
	Slug  string
	Title string
}

type pageData struct {
	Title   string
	Banner  string
	Body    template.HTML
	CSSHref string
	Nav     []navItem
}

// Renderer holds the theme location and the markdown + ascii engines.
type Renderer struct {
	themeDir string
	md       goldmark.Markdown
	ascii    *figlet4go.AsciiRender
}

// New builds a renderer that reads its theme from themeDir.
func New(themeDir string) *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
	)
	return &Renderer{
		themeDir: themeDir,
		md:       md,
		ascii:    figlet4go.NewAsciiRender(),
	}
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
func (r *Renderer) Build(docs []content.Doc) (*Built, error) {
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

	nav := buildNav(docs)
	pages := make(map[string]string, len(docs))
	for _, doc := range docs {
		body, err := r.markdown(doc.Body)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", doc.Slug, err)
		}
		var buf bytes.Buffer
		data := pageData{
			Title:   doc.Title,
			Banner:  r.banner(doc),
			Body:    body,
			CSSHref: cssHref,
			Nav:     nav,
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("execute %s: %w", doc.Slug, err)
		}
		pages[doc.Slug] = buf.String()
	}

	return &Built{Pages: pages, CSS: css, CSSHref: cssHref}, nil
}

func (r *Renderer) markdown(src string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// banner returns the ascii heading: a verbatim frontmatter banner if the author
// hand-drew one, otherwise the title run through figlet.
func (r *Renderer) banner(doc content.Doc) string {
	if doc.Banner != "" {
		return doc.Banner
	}
	art, err := r.ascii.Render(doc.Title)
	if err != nil {
		return doc.Title
	}
	return art
}

func buildNav(docs []content.Doc) []navItem {
	var nav []navItem
	for _, d := range docs {
		if d.Slug == "index" || d.Slug == "404" {
			continue
		}
		nav = append(nav, navItem{Slug: d.Slug, Title: d.Title})
	}
	sort.Slice(nav, func(i, j int) bool { return nav[i].Title < nav[j].Title })
	return nav
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
