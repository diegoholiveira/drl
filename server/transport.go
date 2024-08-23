package server

import (
	"log"
	"net/http"
	"time"
)

type Limiter interface {
	IsAllowed(string) (time.Time, bool)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func NewRoundTripper(original http.RoundTripper, limiter Limiter) http.RoundTripper {
	if original == nil {
		original = http.DefaultTransport
	}

	return roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		token := request.Header.Get("X-Limiter-Token")
		nextWindow, allowed := limiter.IsAllowed(token)
		if !allowed {
			log.Printf("Rate limit exceeded for token \"%s\"\n", token)

			headers := http.Header{
				"Retry-After": []string{nextWindow.Format(time.RFC1123)},
			}

			return &http.Response{
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: http.StatusTooManyRequests,
				Body:       http.NoBody,
				Header:     headers,
			}, nil
		}

		request.Header.Del("X-Limiter-Token")

		log.Printf("Rate limit OK for token \"%s\"\n", token)

		return original.RoundTrip(request)
	})
}
