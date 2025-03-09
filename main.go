package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	// ctxIPv4 is the context key for the client address.
	ctxIPv4 = "addr.ip"
)

// server is a simple HTTP server that forwards requests to a backend.
type server struct {
	forward string
}

// Serve forwards the request to the backend and copies the response back to the client.
func (s *server) Serve(w http.ResponseWriter, r *http.Request) {
	attrs, realip := determineIP(r)

	ctx := context.WithValue(r.Context(), ctxIPv4, realip)
	pan := NewSpan(ctx, "proxy", attrs)
	defer pan.End()

	// Forward the request to the backend
	u := &url.URL{
		Scheme:   "http",
		Host:     s.forward,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}
	req, err := http.NewRequest(r.Method, u.String(), r.Body)
	if err != nil {
		log.Printf("error creating request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		pan.Err(err)
		return
	}
	req.Header = r.Header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error forwarding request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		pan.Err(err)
		return
	}
	defer resp.Body.Close()

	pan.SetAttributes(map[string]interface{}{
		"http.status_code": resp.StatusCode,
	})

	// Copy the response back to the client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("error copying response: %v", err)
		pan.Err(err)
	}
	return
}

func main() {
	// Useful otel env vars
	//   OTEL_EXPORTER_OTLP_INSECURE bool
	//   OTEL_SERVICE_NAME string
	//   OTEL_EXPORTER_OTLP_ENDPOINT string
	//   OTEL_RESOURCE_ATTRIBUTES string

	listen := flag.String("listen", ":8080", "listen address")
	forward := flag.String("forward", "localhost:8000", "forward address")
	flag.Parse()

	log.Printf("listening on %s, forwarding to %s", *listen, *forward)

	_, err := setupOTelSDK(context.Background())
	if err != nil {
		log.Fatalf("failed to setup OpenTelemetry SDK: %v", err)
	}

	s := &server{forward: *forward}
	log.Fatal(http.ListenAndServe(*listen, http.HandlerFunc(s.Serve)))
}
