// Command tveitan serves the personal site at tveitan.se.
//
// It renders markdown to HTML on demand and caches the result, keyed by a
// content+theme version. Editing a markdown file or the theme changes the
// version, which busts the cache on the next request — so content goes live by
// dropping files in a folder, no rebuild and no restart. The same seam
// (content.Source) lets a future backend fetch documents from elsewhere.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/twish/tveitan.se/internal/site"
	"github.com/twish/tveitan.se/pkg/content"
	"github.com/twish/tveitan.se/pkg/render"
)

func main() {
	addr := env("ADDR", ":8080")
	contentDir := env("CONTENT_DIR", "./content")
	themeDir := env("THEME_DIR", "./theme")

	heading, err := site.NewAsciiHeading()
	if err != nil {
		log.Fatal(err)
	}
	renderer, err := render.New(themeDir,
		render.WithHeading(heading),
		render.WithNav(site.UnixNav{}),
		render.WithFooter(site.Footer{}),
		render.WithExtraPages(render.ExtraPage{Doc: site.GalleryDoc(), Body: heading.GalleryBody()}),
	)
	if err != nil {
		log.Fatal(err)
	}
	app := &server{
		source:   content.NewFSSource(contentDir),
		renderer: renderer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/assets/", app.serveAsset)
	mux.Handle("/media/", http.StripPrefix("/media/",
		http.FileServer(http.Dir(filepath.Join(contentDir, "media")))))
	mux.HandleFunc("/", app.servePage)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("tveitan.se listening on %s (content=%s theme=%s)", addr, contentDir, themeDir)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// server holds the content source, renderer, and the cached render of the current
// version. current() rebuilds lazily when the version changes.
type server struct {
	source   content.Source
	renderer *render.Renderer

	mu      sync.RWMutex
	version string
	built   *render.Built
}

// current returns the rendered snapshot for the live content+theme, rebuilding
// only when the combined version has changed since the last call.
func (s *server) current(ctx context.Context) (string, *render.Built, error) {
	contentVer, err := s.source.Version(ctx)
	if err != nil {
		return "", nil, err
	}
	themeVer, err := s.renderer.ThemeVersion()
	if err != nil {
		return "", nil, err
	}
	version := contentVer + "." + themeVer

	s.mu.RLock()
	if s.version == version && s.built != nil {
		built := s.built
		s.mu.RUnlock()
		return version, built, nil
	}
	s.mu.RUnlock()

	docs, err := s.source.List(ctx)
	if err != nil {
		return "", nil, err
	}
	sections, err := s.source.Sections(ctx)
	if err != nil {
		return "", nil, err
	}
	built, err := s.renderer.Build(docs, sections)
	if err != nil {
		return "", nil, err
	}

	s.mu.Lock()
	s.version = version
	s.built = built
	s.mu.Unlock()
	return version, built, nil
}

func (s *server) servePage(w http.ResponseWriter, r *http.Request) {
	version, built, err := s.current(r.Context())
	if err != nil {
		log.Printf("render: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	slug := strings.Trim(r.URL.Path, "/")
	if slug == "" {
		slug = "index"
	}
	page, ok := built.Pages[slug]
	if !ok {
		s.notFound(w, built)
		return
	}

	etag := `"` + version + "." + slug + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache") // always revalidate; cheap via ETag
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(page))
}

func (s *server) notFound(w http.ResponseWriter, built *render.Built) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if page, ok := built.Pages["404"]; ok {
		w.Write([]byte(page))
		return
	}
	w.Write([]byte("404 — not found"))
}

// serveAsset serves the fingerprinted stylesheet. The hash in the URL makes the
// content immutable, so we cache it hard.
func (s *server) serveAsset(w http.ResponseWriter, r *http.Request) {
	_, built, err := s.current(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if r.URL.Path != built.CSSHref {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(built.CSS)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
