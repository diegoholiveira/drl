package server

import (
	"net/http"
	"net/http/httputil"
	"time"
)

const (
	DefaultIdleConnTimeout     time.Duration = 30 * time.Second
	DefaultMaxIdleConns        int           = 32
	DefaultMaxIdleConnsPerHost int           = 16
)

func NewReverseProxy(limiter Limiter) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: NewDirector(),
		Transport: NewRoundTripper(&http.Transport{
			MaxIdleConns:        DefaultMaxIdleConns,
			IdleConnTimeout:     DefaultIdleConnTimeout,
			MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		}, limiter),
	}
}
