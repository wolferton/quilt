package httpserver

import (
	"net/http"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}

type HttpEndpointProvider interface {
	SupportedHttpMethods() []string
	RegexPattern() string
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type ClientIdentity struct {
}

type ServiceErrorCategory int

const (
	Unexpected = iota
	Client
	Logic
	Infrastructure
	Security
)

type CategorisedError struct {
	Category ServiceErrorCategory
	Label    string
}

type ServiceErrors struct {
	Errors []CategorisedError
}

func (se *ServiceErrors) HasErrors() bool {
	return len(se.Errors) != 0
}
