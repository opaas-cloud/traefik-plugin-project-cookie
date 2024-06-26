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
	name     string
	next     http.Handler
	rewrites []rewrite
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rewrites := make([]rewrite, len(config.Rewrites))

	for i, rewriteConfig := range config.Rewrites {
		regex, err := regexp.Compile(rewriteConfig.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", rewriteConfig.Regex, err)
		}

		rewrites[i] = rewrite{
			name:        rewriteConfig.Name,
			regex:       regex,
			replacement: rewriteConfig.Replacement,
		}
	}

	return &rewriteBody{
		name:     name,
		next:     next,
		rewrites: rewrites,
	}, nil
}

var project = ""

func (r *rewriteBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		wrappedWriter := &responseWriter{
			writer:   rw,
			rewrites: r.rewrites,
		}
		r.next.ServeHTTP(wrappedWriter, req)
		return
	}
	var url = req.URL
	fmt.Println(url)
	fmt.Println(url.Host)
	//https://app-airoolite.opaas.online/
	var split1 = strings.Split(url.Host, "-")
	for i := range split1 {
		fmt.Println(split1[i])
	}
	if len(split1) != 2 {
		wrappedWriter := &responseWriter{
			writer:   rw,
			rewrites: r.rewrites,
		}
		r.next.ServeHTTP(wrappedWriter, req)
		return
	}
	var appSplit = split1[1]
	fmt.Println(appSplit)
	var split2 = strings.Split(appSplit, ".")
	for i := range split2 {
		fmt.Println(split2[i])
	}
	if len(split2) != 2 {
		wrappedWriter := &responseWriter{
			writer:   rw,
			rewrites: r.rewrites,
		}
		r.next.ServeHTTP(wrappedWriter, req)
		return
	}

	project = split2[0]
	fmt.Println(project)
	wrappedWriter := &responseWriter{
		writer:   rw,
		rewrites: r.rewrites,
	}

	r.next.ServeHTTP(wrappedWriter, req)
}

type responseWriter struct {
	writer   http.ResponseWriter
	rewrites []rewrite
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
		cookie := http.Cookie{Name: "project", Value: project, Path: "/", HttpOnly: true, Expires: expiration, Domain: "keycloak.opaas.online"}
		http.SetCookie(r, &cookie)
		project = ""
	}
	r.writer.WriteHeader(statusCode)
}
