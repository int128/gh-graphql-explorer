package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
)

func main() {
	port := flag.Int("port", 12345, "Port number to listen on")
	flag.Parse()
	if err := startServer(*port); err != nil {
		log.Fatalf("error: %v", err)
	}
}

//go:embed static
var staticFS embed.FS

func startServer(port int) error {
	staticRootFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("fs.Sub: %w", err)
	}
	http.Handle("GET /", http.FileServerFS(staticRootFS))

	client, err := api.DefaultHTTPClient()
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}
	http.Handle("POST /graphql", &graphqlProxy{client: client, endpoint: graphqlEndpoint()})

	log.Printf("gh-graphql-explorer is available at http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("http.Serve: %w", err)
	}
	return nil
}

func graphqlEndpoint() string {
	host, _ := auth.DefaultHost()
	if auth.IsEnterprise(host) {
		return fmt.Sprintf("https://%s/api/graphql", host)
	}
	return fmt.Sprintf("https://api.%s/graphql", host)
}

type graphqlProxy struct {
	client   *http.Client
	endpoint string
}

func (h *graphqlProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, h.endpoint, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusInternalServerError)
		return
	}
	resp, err := h.client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Printf("%s %s %s %d", req.Method, req.URL, resp.Proto, resp.StatusCode)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Proxy error: %v", err)
	}
}
