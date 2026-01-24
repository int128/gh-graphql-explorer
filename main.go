package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 12345, "Port number to listen on")
	flag.Parse()
	if err := startServer(port); err != nil {
		log.Fatalf("error: %v", err)
	}
}

//go:embed index.html
var staticFS embed.FS

func startServer(port int) error {
	http.Handle("GET /", http.FileServerFS(staticFS))

	graphqlProxy, err := newGraphQLProxy()
	if err != nil {
		return fmt.Errorf("create GraphQL proxy: %w", err)
	}
	http.Handle("POST /graphql", graphqlProxy)

	log.Printf("gh-graphql-explorer is available at http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func newGraphQLProxy() (*proxyHandler, error) {
	host, _ := auth.DefaultHost()
	endpoint := fmt.Sprintf("https://api.%s/graphql", host)
	if auth.IsEnterprise(host) {
		endpoint = fmt.Sprintf("https://%s/api/graphql", host)
	}
	client, err := api.DefaultHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("create GitHub client: %w", err)
	}
	return &proxyHandler{client: client, endpoint: endpoint}, nil
}

type proxyHandler struct {
	client   *http.Client
	endpoint string
}

func (h *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
