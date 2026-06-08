package site

import (
	"fmt"
	"html/template"
	"time"

	"github.com/twish/tveitan.se/pkg/content"
)

// Footer renders the brand mark and the current year. The year is the live part
// (baked at render, refreshes on rebuild) — a hook for whatever dynamic content
// the footer should carry later. It satisfies render.FooterRenderer.
type Footer struct{}

func (Footer) Footer(docs []content.Doc, sections []content.Section) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<footer><span class="foot-brand">// tveitan</span>`+
			` <span class="sep">·</span> `+
			`<span class="foot-year">%d</span></footer>`,
		time.Now().Year()))
}
