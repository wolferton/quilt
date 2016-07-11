package httpserver

import (
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"net/http"
	"regexp"
	"time"
)

const defaultHttpServerConfigBase = "facilities.httpServer"

const commonLogLineFormat = "%s - - %s \"%s %s %s\" %s %s\n"

type RegisteredProvider struct {
	Provider HttpEndpointProvider
	Pattern  *regexp.Regexp
}

type QuiltHttpServer struct {
	Config                      HttpServerConfig
	registeredProvidersByMethod map[string][]*RegisteredProvider
	componentContainer          *ioc.ComponentContainer
	Logger                      logger.Logger
	AccessLogWriter             *AccessLogWriter
}

func (qhs *QuiltHttpServer) Container(container *ioc.ComponentContainer) {
	qhs.componentContainer = container
}

func (qhs *QuiltHttpServer) registerProvider(endPointProvider HttpEndpointProvider) {

	for _, method := range endPointProvider.SupportedHttpMethods() {

		pattern := endPointProvider.RegexPattern()
		compiledRegex, regexError := regexp.Compile(pattern)

		if regexError != nil {
			qhs.Logger.LogErrorf("Unable to compile regular expression from pattern %s: %s", pattern, regexError.Error())
		}

		qhs.Logger.LogTracef("Registering %s %s", pattern, method)

		rp := RegisteredProvider{endPointProvider, compiledRegex}

		providersForMethod := qhs.registeredProvidersByMethod[method]

		if providersForMethod == nil {
			providersForMethod = make([]*RegisteredProvider, 1)
			providersForMethod[0] = &rp
			qhs.registeredProvidersByMethod[method] = providersForMethod
		} else {
			qhs.registeredProvidersByMethod[method] = append(providersForMethod, &rp)
		}
	}

}

func (qhs *QuiltHttpServer) StartComponent() error {

	qhs.registeredProvidersByMethod = make(map[string][]*RegisteredProvider)

	for name, component := range qhs.componentContainer.AllComponents() {
		provider, found := component.Instance.(HttpEndpointProvider)

		if found {
			qhs.Logger.LogDebugf("Found HttpEndpointProvider %s", name)

			qhs.registerProvider(provider)

		}
	}

	return nil
}

func (qhs *QuiltHttpServer) AllowAccess() error {
	http.Handle("/", http.HandlerFunc(qhs.handleAll))

	listenAddress := fmt.Sprintf(":%d", qhs.Config.Port)

	go http.ListenAndServe(listenAddress, nil)

	qhs.Logger.LogInfof("HTTP server started listening on %d", qhs.Config.Port)

	return nil
}

func (h *QuiltHttpServer) handleAll(responseWriter http.ResponseWriter, request *http.Request) {

	received := time.Now()

	contentType := fmt.Sprintf("%s; charset=%s", h.Config.ContentType, h.Config.Encoding)
	responseWriter.Header().Set("Content-Type", contentType)

	providersByMethod := h.registeredProvidersByMethod[request.Method]

	path := request.URL.Path

	h.Logger.LogTracef("Finding provider to handle %s %s from %d providers", path, request.Method, len(providersByMethod))

	for _, handlerPattern := range providersByMethod {

		pattern := handlerPattern.Pattern

		h.Logger.LogTracef("Testing %s", pattern.String())

		if pattern.MatchString(path) {
			h.Logger.LogTracef("Matches %s", pattern.String())

			wrw := new(wrappedResponseWriter)
			wrw.rw = responseWriter

			handlerPattern.Provider.ServeHTTP(wrw, request)

			if h.AccessLogWriter != nil {
				finished := time.Now()
				h.AccessLogWriter.LogRequest(request, wrw, &received, &finished, nil)
			}
		}

	}

}

/*
func (h *QuiltHttpServer) writeAccessLog(responseWriter *wrappedResponseWriter, request *http.Request) {
	f := h.AccessLog
	s := strconv.Itoa(responseWriter.Status)
	b := strconv.Itoa(responseWriter.BytesServed)
	t := time.Now().Format(commonLogDateFormat)
	ll := fmt.Sprintf(commonLogLineFormat, request.RemoteAddr, t, request.Method, request.RequestURI, request.Proto, s, b)

	f.WriteString(ll)
}*/

type HttpServerConfig struct {
	Port        int
	ContentType string
	Encoding    string
}

func ParseDefaultHttpServerConfig(injector *config.ConfigInjector) HttpServerConfig {
	return ParseHttpServerConfig(injector, defaultHttpServerConfigBase)
}

func ParseHttpServerConfig(injector *config.ConfigInjector, baseConfigPath string) HttpServerConfig {

	var httpServerConfig HttpServerConfig
	injector.PopulateObjectFromJsonPath(baseConfigPath, &httpServerConfig)

	return httpServerConfig
}

type wrappedResponseWriter struct {
	rw          http.ResponseWriter
	Status      int
	BytesServed int
}

func (wrw *wrappedResponseWriter) Header() http.Header {
	return wrw.rw.Header()
}

func (wrw *wrappedResponseWriter) Write(b []byte) (int, error) {

	wrw.BytesServed += len(b)

	return wrw.rw.Write(b)
}

func (wrw *wrappedResponseWriter) WriteHeader(i int) {
	wrw.Status = i
	wrw.rw.WriteHeader(i)
}
