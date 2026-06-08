// Package content defines where markdown comes from and how it's modeled.
//
// Source is the pluggable seam: today the site reads markdown from a folder
// (FSSource); tomorrow a LifthrasirSource can implement the same interface to
// pull a document structure from the lifthrasir API. The rest of the program
// only ever talks to the interface.
package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Doc is one renderable page: its identity, presentation hints, and raw body.
type Doc struct {
	Slug   string // url path without extension, e.g. "index" or "posts/hello"
	Title  string
	Banner string // literal ascii art to use verbatim; empty means figlet the title
	Style  string // frozen heading style: index or name; empty means derive from slug
	Date   string
	Order  int  // sort weight in navigation; lower comes first
	Draft  bool // excluded from List and Nav
	Body   string

	Stickers []Sticker // cassette-futurism margin elements; must be page-relevant
}

// Sticker is a margin element pinned beside the content: a relevant note, label,
// pulled snippet, or image. Content, never decoration.
type Sticker struct {
	Type   string // note | label | snippet | image
	Text   string // note/label/snippet body, or image caption
	Src    string // image source (type image)
	Side   string // left | right
	At     string // vertical position down the article, e.g. "25%"
	Rotate int    // tilt in degrees
	Size   string // sm | md | lg
	Gap    string // distance from the text column, e.g. "4rem"; varies the spacing
}

// Section is an indexed top-level directory (defined by its _index.md). It gets
// a listing page at /<slug> and can appear in the nav and on the start page.
type Section struct {
	Slug   string // directory name, e.g. "posts"
	Title  string
	Banner string
	Style  string
	Nav    bool   // show in the top nav
	Order  int    // nav + start-page ordering
	Sort   string // how to sort entries: "date" (newest first) or "order"
	Body   string // intro markdown shown above the listing

	HomeMode   string   // how it surfaces on the start page: latest | pinned | none
	HomeCount  int      // for latest
	HomePinned []string // entry slugs, for pinned
}

// Section returns the top-level directory a doc belongs to, or "" for a
// top-level page.
func (d Doc) Section() string {
	if i := strings.IndexByte(d.Slug, '/'); i >= 0 {
		return d.Slug[:i]
	}
	return ""
}

// Entries returns a section's pages (the docs directly under its directory),
// sorted by the section's rule: "date" newest-first, otherwise by order.
func Entries(sec Section, docs []Doc) []Doc {
	var out []Doc
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

// Source is everything the renderer needs from a content backend.
//
// Version must change whenever any doc changes, and must be cheap to compute —
// it is the cache key. A folder source hashes file stats; a remote source can
// return an etag or a content digest.
type Source interface {
	List(ctx context.Context) ([]Doc, error)
	Sections(ctx context.Context) ([]Section, error)
	Version(ctx context.Context) (string, error)
}

// frontmatter is the YAML block parsed from the top of a markdown file.
type frontmatter struct {
	Title  string `yaml:"title"`
	Banner string `yaml:"banner"`
	// Style accepts either an int index (style: 23) or a name (style: slant-sunset),
	// so it's read loosely and normalized to a string.
	Style  any `yaml:"style"`
	Date   string      `yaml:"date"`
	Order  int         `yaml:"order"`
	Draft  bool        `yaml:"draft"`

	Stickers []stickerFM `yaml:"stickers"`
}

type stickerFM struct {
	Type   string `yaml:"type"`
	Text   string `yaml:"text"`
	Src    string `yaml:"src"`
	Side   string `yaml:"side"`
	At     any    `yaml:"at"` // "25%" or 25; normalized to a percent string
	Rotate int    `yaml:"rotate"`
	Size   string `yaml:"size"`
	Gap    string `yaml:"gap"`
}

// FSSource serves markdown from a directory tree of *.md files.
type FSSource struct {
	dir string
}

// NewFSSource roots a folder source at dir.
func NewFSSource(dir string) *FSSource {
	return &FSSource{dir: dir}
}

// List reads and parses every non-draft markdown file under the root.
func (s *FSSource) List(ctx context.Context) ([]Doc, error) {
	var docs []Doc
	err := filepath.WalkDir(s.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if d.Name() == "_index.md" {
			return nil // section definition, not a page
		}
		doc, err := s.read(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if doc.Draft {
			return nil
		}
		docs = append(docs, doc)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Order != docs[j].Order {
			return docs[i].Order < docs[j].Order
		}
		return docs[i].Title < docs[j].Title
	})
	return docs, nil
}

// Version hashes the path, size and modtime of every markdown file. Cheap
// enough to call on every request; changes the moment any file is touched, so
// edits go live without a restart and the render cache busts on its own.
func (s *FSSource) Version(ctx context.Context) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(s.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s|%d|%d\n", path, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// sectionFM is the frontmatter of a directory's _index.md.
type sectionFM struct {
	Title  string `yaml:"title"`
	Banner string `yaml:"banner"`
	Style  any    `yaml:"style"`
	Nav    bool   `yaml:"nav"`
	Order  int    `yaml:"order"`
	Sort   string `yaml:"sort"`
	Home   struct {
		Mode   string   `yaml:"mode"`
		Count  int      `yaml:"count"`
		Pinned []string `yaml:"pinned"`
	} `yaml:"home"`
}

// Sections reads every top-level directory's _index.md.
func (s *FSSource) Sections(ctx context.Context) ([]Section, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var sections []Section
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(s.dir, e.Name(), "_index.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue // a directory without _index.md is not a section
		}
		var fm sectionFM
		body := splitInto(raw, &fm)
		title := fm.Title
		if title == "" {
			title = e.Name()
		}
		sort := fm.Sort
		if sort == "" {
			sort = "order"
		}
		sections = append(sections, Section{
			Slug:       e.Name(),
			Title:      title,
			Banner:     fm.Banner,
			Style:      styleString(fm.Style),
			Nav:        fm.Nav,
			Order:      fm.Order,
			Sort:       sort,
			Body:       string(body),
			HomeMode:   fm.Home.Mode,
			HomeCount:  fm.Home.Count,
			HomePinned: fm.Home.Pinned,
		})
	}
	return sections, nil
}

func (s *FSSource) read(path string) (Doc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Doc{}, err
	}
	fm, body := splitFrontmatter(raw)

	rel, err := filepath.Rel(s.dir, path)
	if err != nil {
		return Doc{}, err
	}
	slug := strings.TrimSuffix(filepath.ToSlash(rel), ".md")

	title := fm.Title
	if title == "" {
		title = slug
	}
	return Doc{
		Slug:   slug,
		Title:  title,
		Banner: fm.Banner,
		Style:  styleString(fm.Style),
		Date:   fm.Date,
		Order:  fm.Order,
		Draft:  fm.Draft,
		Body:   string(body),
		Stickers: stickers(fm.Stickers),
	}, nil
}

func stickers(in []stickerFM) []Sticker {
	var out []Sticker
	for _, s := range in {
		out = append(out, Sticker{
			Type:   s.Type,
			Text:   s.Text,
			Src:    s.Src,
			Side:   s.Side,
			At:     atPercent(s.At),
			Rotate: s.Rotate,
			Size:   s.Size,
			Gap:    s.Gap,
		})
	}
	return out
}

// atPercent normalizes a vertical position to a CSS percent string; a bare
// number becomes "<n>%".
func atPercent(v any) string {
	if v == nil {
		return "20%"
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" {
		return "20%"
	}
	if !strings.Contains(s, "%") {
		return s + "%"
	}
	return s
}

// styleString normalizes a loosely-typed frontmatter `style` (int or name) to a
// trimmed string; nil/absent becomes empty.
func styleString(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

// splitFrontmatter peels a leading `---\n…\n---\n` YAML block off the document.
// A file without one is all body — frontmatter is optional.
func splitFrontmatter(raw []byte) (frontmatter, []byte) {
	var fm frontmatter
	body := splitInto(raw, &fm)
	return fm, body
}

// splitInto peels the leading YAML frontmatter and unmarshals it into target,
// returning the remaining body. Missing or malformed frontmatter leaves target
// zero-valued and returns the whole input as body, so a bad block never blanks
// the page.
func splitInto(raw []byte, target any) []byte {
	const fence = "---"
	text := string(raw)
	if !strings.HasPrefix(text, fence+"\n") && !strings.HasPrefix(text, fence+"\r\n") {
		return raw
	}
	rest := text[len(fence):]
	rest = strings.TrimLeft(rest, "\r\n")
	end := strings.Index(rest, "\n"+fence)
	if end < 0 {
		return raw
	}
	head := rest[:end]
	body := strings.TrimLeft(rest[end+len("\n"+fence):], "\r\n")
	if err := yaml.Unmarshal([]byte(head), target); err != nil {
		return raw
	}
	return []byte(body)
}
