package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxMCPBodyBytes = 1 << 20

func serveHTTP(ctx context.Context, addr string, server *mcp.Server) error {
	token := strings.TrimSpace(os.Getenv("YOKAI_FINDER_TOKEN"))
	if err := checkListenAddr(addr, token); err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           newHTTPHandler(server, token, isLoopbackHost(listenHost(addr))),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func newHTTPHandler(server *mcp.Server, token string, loopbackOnly bool) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless:      true,
		JSONResponse:   true,
		SessionTimeout: 5 * time.Minute,
	})

	mux := http.NewServeMux()
	guarded := withHTTPGuards(http.MaxBytesHandler(mcpHandler, maxMCPBodyBytes), token, loopbackOnly)
	mux.Handle("/mcp", guarded)
	mux.Handle("/mcp/", guarded)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	return mux
}

func listenHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func checkListenAddr(addr, token string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid -http address %q: %w", addr, err)
	}
	if isLoopbackHost(host) {
		return nil
	}
	if token == "" {
		return fmt.Errorf("refusing to bind %s without YOKAI_FINDER_TOKEN (loopback is allowed without a token). This server speaks cleartext HTTP; put TLS in front for any non-loopback deploy", addr)
	}
	if len(token) < 16 {
		return fmt.Errorf("YOKAI_FINDER_TOKEN must be at least 16 characters for non-loopback binds")
	}
	return nil
}

func isLoopbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "127.0.0.1", "localhost", "::1":
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func withHTTPGuards(next http.Handler, token string, loopbackOnly bool) http.Handler {
	limiter := newIPLimiter(120, time.Minute, 4096)
	want := tokenDigest(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if loopbackOnly && !localRequest(r) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if !limiter.allow(clientIP(r)) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		if token != "" {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			got := tokenDigest(strings.TrimSpace(header[len(prefix):]))
			if subtle.ConstantTimeCompare(got, want) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func localRequest(r *http.Request) bool {
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if !isLoopbackHost(host) {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		uHost := origin
		if strings.Contains(origin, "://") {
			uHost = origin[strings.Index(origin, "://")+3:]
		}
		if h, _, err := net.SplitHostPort(uHost); err == nil {
			uHost = h
		}
		uHost = strings.TrimSuffix(uHost, "/")
		if !isLoopbackHost(uHost) {
			return false
		}
	}
	return true
}

func tokenDigest(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type ipLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	maxKeys int
	hits    map[string][]time.Time
}

func newIPLimiter(limit int, window time.Duration, maxKeys int) *ipLimiter {
	return &ipLimiter{limit: limit, window: window, maxKeys: maxKeys, hits: map[string][]time.Time{}}
}

func (l *ipLimiter) allow(ip string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	kept := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.hits, ip)
	}
	if _, exists := l.hits[ip]; !exists && len(l.hits) >= l.maxKeys {
		for k := range l.hits {
			delete(l.hits, k)
			break
		}
	}
	if len(kept) >= l.limit {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}
