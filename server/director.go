package server

import (
	"net/http"
)

func NewDirector() func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = "news.ycombinator.com"
	}
}
