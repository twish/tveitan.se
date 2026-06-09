package render

import (
	"fmt"
	"strings"

	"github.com/twish/tveitan.se/pkg/content"
	"gopkg.in/yaml.v2"
)

// SiteConfig toggles the machine-friendly artifacts and carries the metadata
// they need. Read from a YAML file at runtime (see ParseConfig), so it can be
// changed live like content.
type SiteConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	BaseURL     string `yaml:"base_url"`
	LLMsTxt     bool   `yaml:"llms_txt"`
	RawMarkdown bool   `yaml:"raw_markdown"`
	Sitemap     bool   `yaml:"sitemap"`
	Robots      bool   `yaml:"robots"`
}

// File is a non-HTML artifact served at a fixed path.
type File struct {
	ContentType string
	Body        []byte
}

// ParseConfig parses site.yaml bytes; nil/empty or malformed input yields
// defaults (all artifacts on).
func ParseConfig(data []byte) SiteConfig {
	cfg := SiteConfig{LLMsTxt: true, RawMarkdown: true, Sitemap: true, Robots: true}
	if len(data) > 0 {
		_ = yaml.Unmarshal(data, &cfg) // bad config falls back to defaults
	}
	return cfg
}

// artifacts builds the enabled machine-friendly files, keyed by URL path.
func (r *Renderer) artifacts(docs []content.Doc, sections []content.Section, cfg SiteConfig) map[string]File {
	files := map[string]File{}
	if cfg.LLMsTxt {
		files["/llms.txt"] = File{"text/plain; charset=utf-8", []byte(llmsTxt(docs, sections, cfg))}
	}
	if cfg.RawMarkdown {
		for _, d := range docs {
			files["/"+d.Slug+".md"] = File{"text/markdown; charset=utf-8", []byte(d.Body)}
		}
		for _, s := range sections {
			files["/"+s.Slug+".md"] = File{"text/markdown; charset=utf-8", []byte(s.Body)}
		}
	}
	if cfg.Sitemap {
		files["/sitemap.xml"] = File{"application/xml; charset=utf-8", []byte(sitemap(docs, sections, r.extra, cfg))}
	}
	if cfg.Robots {
		files["/robots.txt"] = File{"text/plain; charset=utf-8", []byte(robots(cfg))}
	}
	return files
}

func link(cfg SiteConfig, slug string) string { return cfg.BaseURL + "/" + slug }

// summary is the first non-heading, non-empty line of a doc, trimmed — a rough
// one-line description for the llms.txt index.
func summary(body string) string {
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 120 {
			line = strings.TrimSpace(line[:120]) + "…"
		}
		return line
	}
	return ""
}

func llmsTxt(docs []content.Doc, sections []content.Section, cfg SiteConfig) string {
	name := cfg.Name
	if name == "" {
		name = "site"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", name)
	if cfg.Description != "" {
		fmt.Fprintf(&b, "\n> %s\n", cfg.Description)
	}

	var pages []content.Doc
	for _, d := range docs {
		if d.Slug == "index" || d.Slug == "404" || strings.Contains(d.Slug, "/") {
			continue
		}
		pages = append(pages, d)
	}
	if len(pages) > 0 {
		b.WriteString("\n## Pages\n\n")
		for _, d := range pages {
			fmt.Fprintf(&b, "- [%s](%s): %s\n", d.Title, link(cfg, d.Slug), summary(d.Body))
		}
	}

	for _, s := range sortByOrder(sections) {
		fmt.Fprintf(&b, "\n## %s\n\n", s.Title)
		for _, d := range content.Entries(s, docs) {
			fmt.Fprintf(&b, "- [%s](%s): %s\n", d.Title, link(cfg, d.Slug), summary(d.Body))
		}
	}
	return b.String()
}

func sitemap(docs []content.Doc, sections []content.Section, extra []ExtraPage, cfg SiteConfig) string {
	var locs []string
	locs = append(locs, cfg.BaseURL+"/")
	for _, d := range docs {
		if d.Slug == "index" || d.Slug == "404" {
			continue
		}
		locs = append(locs, link(cfg, d.Slug))
	}
	for _, s := range sections {
		locs = append(locs, link(cfg, s.Slug))
	}
	for _, ep := range extra {
		locs = append(locs, link(cfg, ep.Doc.Slug))
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, loc := range locs {
		fmt.Fprintf(&b, "  <url><loc>%s</loc></url>\n", xmlEscape(loc))
	}
	b.WriteString("</urlset>\n")
	return b.String()
}

func robots(cfg SiteConfig) string {
	var b strings.Builder
	b.WriteString("User-agent: *\nAllow: /\n")
	if cfg.Sitemap && cfg.BaseURL != "" {
		fmt.Fprintf(&b, "\nSitemap: %s/sitemap.xml\n", cfg.BaseURL)
	}
	return b.String()
}

func sortByOrder(sections []content.Section) []content.Section {
	out := append([]content.Section(nil), sections...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Order < out[j-1].Order; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
