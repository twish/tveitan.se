package render

import (
	"html/template"

	"github.com/twish/tveitan.se/pkg/content"
)

// Minimal, look-agnostic defaults so the engine renders something usable on its
// own. A real site replaces these via WithHeading / WithNav / WithFooter.

type plainHeading struct{}

func (plainHeading) Heading(doc content.Doc) template.HTML {
	if doc.Banner != "" {
		return template.HTML(`<div class="hero"><pre>` + template.HTMLEscapeString(doc.Banner) + `</pre></div>`)
	}
	return template.HTML(`<header class="hero"><h1>` + template.HTMLEscapeString(doc.Title) + `</h1></header>`)
}

type plainNav struct{}

func (plainNav) Nav(current string, docs []content.Doc, sections []content.Section) template.HTML {
	return ""
}

type plainFooter struct{}

func (plainFooter) Footer(docs []content.Doc, sections []content.Section) template.HTML {
	return ""
}
