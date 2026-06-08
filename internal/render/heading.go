package render

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/mbndr/figlet4go"
	"github.com/twish/tveitan.se/internal/content"
)

// HeadingRenderer produces the markup for a page's heading. It's a swap point:
// this site renders ascii-art banners (asciiHeading); another site on the same
// engine could implement it to return a plain <h1> over a hero image.
type HeadingRenderer interface {
	Heading(doc content.Doc) template.HTML
}

// asciiHeading renders the figlet banner with the frozen synthwave style system
// (font + palette + placement + gradient angle), breaking out as a full-width
// hero.
type asciiHeading struct {
	ascii *figlet4go.AsciiRender
}

func (h asciiHeading) Heading(doc content.Doc) template.HTML {
	st := selectStyle(doc)
	banner := h.banner(doc, st)
	wide := ""
	if st.Wide {
		wide = " wide"
	}
	markup := fmt.Sprintf(
		`<div class="hero" style="text-align:%s">`+
			`<pre class="banner%s" style="--cols:%d;--g1:%s;--g2:%s;--ga:%s">%s</pre></div>`,
		st.Align, wide, maxLineLen(banner),
		st.Palette.g1, st.Palette.g2, st.Angle, template.HTMLEscapeString(banner))
	return template.HTML(markup)
}

// banner is a verbatim frontmatter banner, or the title in the style's font.
func (h asciiHeading) banner(doc content.Doc, st bannerStyle) string {
	if doc.Banner != "" {
		return doc.Banner
	}
	return figlet(h.ascii, doc.Title, st.Font)
}

func figlet(ar *figlet4go.AsciiRender, text, font string) string {
	opt := figlet4go.NewRenderOptions()
	opt.FontName = font
	art, err := ar.RenderOpts(text, opt)
	if err != nil {
		return text
	}
	return strings.Trim(art, "\n")
}
