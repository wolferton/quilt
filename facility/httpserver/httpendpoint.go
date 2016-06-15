package httpserver

import (
	"net/http"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}
