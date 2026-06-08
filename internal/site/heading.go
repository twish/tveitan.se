// Package site holds tveitan.se's concrete look: the ascii-art heading, the
// unix-shell nav, the footer, the synthwave style system, and the /styles
// gallery. The generic engine lives in pkg/render; these types implement its
// renderer interfaces and are wired in from main.
package site

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/mbndr/figlet4go"
	"github.com/twish/tveitan.se/pkg/content"
)

//go:embed fonts/*.flf
var fontFS embed.FS

// AsciiHeading renders the figlet banner with the frozen synthwave style system
// (font + palette + placement + gradient angle), breaking out as a full-width
// hero. It satisfies render.HeadingRenderer.
type AsciiHeading struct {
	ascii *figlet4go.AsciiRender
}

// NewAsciiHeading creates the heading renderer and loads the embedded fonts.
func NewAsciiHeading() (*AsciiHeading, error) {
	ascii := figlet4go.NewAsciiRender()
	if err := loadFonts(ascii); err != nil {
		return nil, err
	}
	return &AsciiHeading{ascii: ascii}, nil
}

func (h *AsciiHeading) Heading(doc content.Doc) template.HTML {
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
func (h *AsciiHeading) banner(doc content.Doc, st bannerStyle) string {
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

// loadFonts registers every embedded *.flf under its lowercased file stem.
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

// maxLineLen returns the widest line in runes, so CSS can scale the banner font
// to fit.
func maxLineLen(s string) int {
	max := 1
	for line := range strings.SplitSeq(s, "\n") {
		if n := len([]rune(line)); n > max {
			max = n
		}
	}
	return max
}
