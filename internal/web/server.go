package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"
)

//go:embed static/*
var staticFS embed.FS

// Server serves the jvman promotional site.
type Server struct {
	addr string
	mux  *http.ServeMux
}

// New creates a web server listening on addr (e.g. ":8080").
func New(addr string) *Server {
	s := &Server{addr: addr, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(static))
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	s.mux.Handle("GET /", http.HandlerFunc(s.handleIndex))
	s.mux.HandleFunc("GET /health", s.handleHealth)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "page not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.loggingMiddleware(s.mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	fmt.Printf("jvman site → http://localhost%s\n", s.addr)
	return srv.ListenAndServe()
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
