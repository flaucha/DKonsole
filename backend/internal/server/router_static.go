package server

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/middleware"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

func registerStaticRoutes(c RouterConfig) {
	staticDir := c.Deps.StaticDir
	if staticDir == "" {
		staticDir = "static"
	}

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		utils.LogWarn("Static directory not found, frontend will not be served", map[string]interface{}{
			"static_dir": staticDir,
		})
	} else {
		secureStaticFiles := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fixer := &contentTypeFixer{
					ResponseWriter: w,
					path:           r.URL.Path,
				}
				middleware.SecurityHeadersHandler(next).ServeHTTP(fixer, r)
			})
		}

		assetsDir := filepath.Join(staticDir, "assets")
		fileServer := http.FileServer(http.Dir(assetsDir))
		c.Mux.Handle("/assets/", secureStaticFiles(http.StripPrefix("/assets/", fileServer)))

		c.Mux.HandleFunc("/favicon.ico", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "favicon.ico"))
		}))
		c.Mux.HandleFunc("/favicon.png", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			faviconPath := filepath.Join(staticDir, "favicon.png")
			if _, err := os.Stat(faviconPath); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeFile(w, r, faviconPath)
		}))
		c.Mux.HandleFunc("/logo-full-dark.png", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			logoPath := filepath.Join(staticDir, "logo-full-dark.png")
			if _, err := os.Stat(logoPath); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeFile(w, r, logoPath)
		}))
		c.Mux.HandleFunc("/logo-full-light.png", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			logoPath := filepath.Join(staticDir, "logo-full-light.png")
			if _, err := os.Stat(logoPath); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeFile(w, r, logoPath)
		}))
		c.Mux.HandleFunc("/robots.txt", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "robots.txt"))
		}))

		c.Mux.HandleFunc("/", middleware.SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				http.NotFound(w, r)
				return
			}
			if r.URL.Path == "/favicon.ico" || r.URL.Path == "/favicon.png" || r.URL.Path == "/logo-full-dark.png" || r.URL.Path == "/logo-full-light.png" || r.URL.Path == "/robots.txt" {
				http.NotFound(w, r)
				return
			}
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
			} else {
				http.NotFound(w, r)
			}
		}))
	}
}

// ResponseWriter wrapper to fix Content-Type after FileServer sets it
type contentTypeFixer struct {
	http.ResponseWriter
	path string
}

func (c *contentTypeFixer) WriteHeader(code int) {
	ext := filepath.Ext(c.path)
	switch ext {
	case ".js":
		c.ResponseWriter.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".css":
		c.ResponseWriter.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".json":
		c.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		c.ResponseWriter.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		c.ResponseWriter.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		c.ResponseWriter.Header().Set("Content-Type", "image/svg+xml")
	case ".woff", ".woff2":
		c.ResponseWriter.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		c.ResponseWriter.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		c.ResponseWriter.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	default:
		if ct := mime.TypeByExtension(ext); ct != "" {
			c.ResponseWriter.Header().Set("Content-Type", ct)
		}
	}
	c.ResponseWriter.WriteHeader(code)
}
