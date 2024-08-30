// Package traefik_plugin_proxy_cookie a traefik plugin providing the functionality of the nginx proxy_cookie directives tp traefik.
package traefik_plugin_project_cookie //nolint

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const setCookieHeader string = "Set-Cookie"

type Rewrite struct {
	Name        string `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Regex       string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty"`
}

// Config holds the plugin configuration.
type Config struct {
	Rewrites []Rewrite `json:"rewrites,omitempty" toml:"rewrites,omitempty" yaml:"rewrites,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewrite struct {
	name        string
	regex       *regexp.Regexp
	replacement string
}

type rewriteBody struct {
	name string
	next http.Handler
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &rewriteBody{
		name: name,
		next: next,
	}, nil
}

var project = ""

func (r *rewriteBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" && strings.Contains(req.Host, "app-") && req.URL.Path == "/" {
		fmt.Println("HOST")
		fmt.Println(req.Host)
		var split1 = strings.Split(req.Host, "-")

		var appSplit = split1[1]
		fmt.Println(appSplit)

		var secondSplit = strings.Split(appSplit, ".")

		project = secondSplit[0]
		fmt.Println(project)

		writer := &responseWriter{
			writer: rw,
		}
		r.next.ServeHTTP(writer, req)
	} else {
		r.next.ServeHTTP(rw, req)
	}
}

type responseWriter struct {
	writer http.ResponseWriter
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.writer.Write(bytes)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if project != "" {
		fmt.Println("Set new cookie")
		fmt.Println("project found")
		r.writer.Header().Del(setCookieHeader)
		expiration := time.Now().Add(24 * 7 * time.Hour)
		cookie := http.Cookie{Name: "project", Value: project, Path: "/", HttpOnly: true, Expires: expiration, Domain: ".opaas.online"}
		http.SetCookie(r, &cookie)
		project = ""
	}
	r.writer.WriteHeader(statusCode)
}
