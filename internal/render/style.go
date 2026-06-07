package render

import (
	"crypto/sha256"
	"strconv"

	"github.com/twish/tveitan.se/internal/content"
)

// palette is a two-stop gradient for the banner text. All stops stay within the
// synthwave range; "sunset" and "teal-purple" are the ARC-inspired fades.
type palette struct {
	name string
	g1   string
	g2   string
}

var palettes = []palette{
	{"cyan-pink", "#2de2e6", "#ff2e97"},
	{"sunset", "#ff2e97", "#ff8a3d"},
	{"teal-purple", "#2de2e6", "#b967ff"},
	{"purple-pink", "#b967ff", "#ff2e97"},
	{"amber-pink", "#ffd319", "#ff2e97"},
}

// figlet fonts embedded under fonts/. Names are lowercased file stems.
var fonts = []string{"standard", "slant", "small", "big", "doom", "ogre", "shadow", "speed"}

// positions cycle across the style list: three placements plus a full-width fill.
var positions = []string{"left", "center", "right", "wide"}

// gradient angles cycle too: top-to-bottom is the default synthwave look; a
// third land on a diagonal for variety.
var angles = []string{"180deg", "180deg", "135deg"}

// bannerStyle is one frozen look: a font, a palette, and a placement. The list
// is a stable 8×5 grid (40 styles); a heading's style index never moves.
type bannerStyle struct {
	Name    string
	Font    string
	Palette palette
	Align   string // text-align for the placement: left | center | right
	Wide    bool   // fill ~90vw instead of sitting at natural capped size
	Angle   string // gradient direction, e.g. 180deg (top-bottom) or 135deg
}

var styles = buildStyles()

func buildStyles() []bannerStyle {
	out := make([]bannerStyle, 0, len(fonts)*len(palettes))
	i := 0
	for _, f := range fonts {
		for _, p := range palettes {
			pos := positions[i%len(positions)]
			align := pos
			wide := false
			if pos == "wide" {
				align = "center"
				wide = true
			}
			out = append(out, bannerStyle{
				Name:    f + "-" + p.name,
				Font:    f,
				Palette: p,
				Align:   align,
				Wide:    wide,
				Angle:   angles[i%len(angles)],
			})
			i++
		}
	}
	return out
}

// selectStyle resolves a doc to its frozen style: an explicit frontmatter
// `style` (index or name) wins; otherwise it's derived deterministically from
// the slug so the same page always renders the same way.
func selectStyle(doc content.Doc) bannerStyle {
	if doc.Style != "" {
		if idx, err := strconv.Atoi(doc.Style); err == nil && idx >= 0 && idx < len(styles) {
			return styles[idx]
		}
		for _, s := range styles {
			if s.Name == doc.Style {
				return s
			}
		}
	}
	sum := sha256.Sum256([]byte(doc.Slug))
	idx := int(sum[0]) | int(sum[1])<<8
	return styles[idx%len(styles)]
}
