package render

import (
	"fmt"
	"html/template"
	"sort"
	"strings"

	"github.com/twish/tveitan.se/internal/content"
)

// NavRenderer turns the content tree plus the current location into navigation
// markup. It's a swap point: this site renders a unix-shell breadcrumb + ls
// (unixNav); another site could implement it as a flat top menu.
type NavRenderer interface {
	Nav(current string, docs []content.Doc, sections []content.Section) template.HTML
}

// unixNav renders navigation the way a shell shows location: a pwd breadcrumb
// (~ / posts / hello) and the ls of the current directory as [ a | b | c ],
// with directories suffixed "/" and the current entry highlighted.
type unixNav struct{}

type crumb struct {
	Label string
	Href  string
	Here  bool
}

type navOpt struct {
	Label string
	Href  string
	Dir   bool
	Here  bool
}

func (unixNav) Nav(current string, docs []content.Doc, sections []content.Section) template.HTML {
	crumbs, opts, isDir := unixLayout(current, docs, sections)

	var b strings.Builder
	b.WriteString(`<nav class="nav-unix">`)

	b.WriteString(`<span class="crumbs">`)
	for i, c := range crumbs {
		if i > 0 {
			b.WriteString(`<span class="sep">/</span>`)
		}
		if c.Here || c.Href == "" {
			fmt.Fprintf(&b, `<span class="here">%s</span>`, template.HTMLEscapeString(c.Label))
		} else {
			fmt.Fprintf(&b, `<a href="%s">%s</a>`, c.Href, template.HTMLEscapeString(c.Label))
		}
	}
	if isDir {
		b.WriteString(`<span class="sep">/</span>`)
	}
	b.WriteString(`</span>`)

	if len(opts) > 0 {
		b.WriteString(`<span class="ls"><span class="bracket">[</span> `)
		for i, o := range opts {
			if i > 0 {
				b.WriteString(` <span class="pipe">|</span> `)
			}
			label := o.Label
			if o.Dir {
				label += "/"
			}
			if o.Here {
				fmt.Fprintf(&b, `<span class="here">%s</span>`, template.HTMLEscapeString(label))
			} else {
				fmt.Fprintf(&b, `<a href="%s">%s</a>`, o.Href, template.HTMLEscapeString(label))
			}
		}
		b.WriteString(` <span class="bracket">]</span></span>`)
	}

	b.WriteString(`</nav>`)
	return template.HTML(b.String())
}

// unixLayout computes the breadcrumb and the ls of the current directory.
// isDir reports whether the current location is itself a directory (home or a
// section), which gets a trailing slash in the breadcrumb.
func unixLayout(current string, docs []content.Doc, sections []content.Section) ([]crumb, []navOpt, bool) {
	secBySlug := make(map[string]content.Section, len(sections))
	for _, s := range sections {
		secBySlug[s.Slug] = s
	}

	crumbs := []crumb{{Label: "~", Href: "/", Here: current == "index"}}
	dir := "" // directory whose ls we list ("" = root)
	isDir := current == "index"

	if current != "index" {
		parts := strings.Split(current, "/")
		acc := ""
		for i, p := range parts {
			if i > 0 {
				acc += "/"
			}
			acc += p
			crumbs = append(crumbs, crumb{Label: p, Href: "/" + acc, Here: acc == current})
		}
		if _, ok := secBySlug[current]; ok {
			dir, isDir = current, true
		} else if len(parts) > 1 {
			dir = parts[0]
		}
	}

	var opts []navOpt
	if dir == "" {
		for _, d := range docs {
			if d.Slug == "index" || d.Slug == "404" || strings.Contains(d.Slug, "/") {
				continue
			}
			opts = append(opts, navOpt{Label: d.Slug, Href: "/" + d.Slug, Here: d.Slug == current})
		}
		for _, s := range sections {
			opts = append(opts, navOpt{Label: s.Slug, Href: "/" + s.Slug, Dir: true, Here: s.Slug == current})
		}
		sort.SliceStable(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	} else if sec, ok := secBySlug[dir]; ok {
		for _, d := range entriesOf(sec, docs) {
			label := d.Slug
			if i := strings.LastIndexByte(d.Slug, '/'); i >= 0 {
				label = d.Slug[i+1:]
			}
			opts = append(opts, navOpt{Label: label, Href: "/" + d.Slug, Here: d.Slug == current})
		}
	}
	return crumbs, opts, isDir
}
