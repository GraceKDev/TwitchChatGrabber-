package apiGateway

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"twitchChat/api"
)

// rateLimiter tracks request counts per IP within a sliding window.
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Evict timestamps outside the window
	times := rl.requests[ip]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.requests[ip] = append(valid, now)

	return len(rl.requests[ip]) <= rl.limit
}

// RateLimit middleware limits each IP to limit requests per window duration.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ip := req.RemoteAddr
			if i := strings.LastIndex(ip, ":"); i != -1 {
				ip = ip[:i]
			}
			if !rl.allow(ip) {
				http.Error(res, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(res, req)
		})
	}
}

// requireAPIKey is a middleware that rejects requests missing a valid X-API-Key header.
func requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		key := req.Header.Get("X-API-Key")
		if key == "" || key != os.Getenv("API_KEY") {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func SetupRoutes() {
	publicMux := http.NewServeMux()
	publicMux.Handle("/", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		res.WriteHeader(http.StatusOK)
		fmt.Fprintln(res, "Hello, World!")
	}))
	privateMux := http.NewServeMux()
	privateMux.Handle("/secret", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		res.WriteHeader(http.StatusOK)
		fmt.Fprintln(res, "Secret data!")
	}))

	privateMux.Handle("/getVodChatReplay", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		videoId := req.URL.Query().Get("videoId")
		if videoId == "" {
			http.Error(res, "Missing required query param: videoId", http.StatusBadRequest)
			return
		}

		offset := 0
		if raw := req.URL.Query().Get("offset"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(res, "Invalid query param: offset must be an integer", http.StatusBadRequest)
				return
			}
			offset = parsed
		}

		api.GetVideoCommentsByOffset(os.Getenv("TWITCH_CLIENT_ID"), videoId, offset)
		res.WriteHeader(http.StatusOK)
		fmt.Fprintln(res, "OK")
	}))

	http.Handle(
		"/api/v1/",
		PublicMiddleware(
			http.StripPrefix("/api/v1", publicMux),
		),
	)
	http.Handle(
		"/api/v1/private/",
		Middleware(
			http.StripPrefix("/api/v1/private", privateMux),
		),
	)

}

func Middleware(next http.Handler) http.Handler {
	limiter := RateLimit(60, time.Minute)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		requireAPIKey(limiter(next)).ServeHTTP(res, req)
	})
}

func PublicMiddleware(next http.Handler) http.Handler {
	limiter := RateLimit(60, time.Minute)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		limiter(next).ServeHTTP(res, req)
	})
}
