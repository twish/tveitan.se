package site

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/twish/tveitan.se/pkg/content"
)

// GalleryDoc is the synthetic page the /styles gallery renders into.
func GalleryDoc() content.Doc {
	return content.Doc{Slug: "styles", Title: "styles", Style: "slant-cyan-pink"}
}

// GalleryBody builds the /styles page: every style rendered with sample text,
// labelled by index and name, so a specific look can be picked by number.
func (h *AsciiHeading) GalleryBody() template.HTML {
	const sample = "tveitan"
	var b strings.Builder
	b.WriteString(`<p>Every heading style, frozen. Set <code>style: N</code> (or the name) in a page's frontmatter to pin one; leave it out and it's derived from the slug.</p>`)
	b.WriteString(`<div class="gallery">`)
	for i, st := range styles {
		art := figlet(h.ascii, sample, st.Font)
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
