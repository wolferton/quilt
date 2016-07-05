package httpserver

import (
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"net/http"
	"regexp"
)

const defaultHttpServerConfigBase = "facilities.httpServer"
const httpListenPortPath = "listenPort"
const httpContentTypePath = "contentType"
const httpEncodingPath = "encoding"

type RegisteredProvider struct {
	Provider HttpEndpointProvider
	Pattern  *regexp.Regexp
}

type QuiltHttpServer struct {
	Config                      HttpServerConfig
	registeredProvidersByMethod map[string][]*RegisteredProvider
	componentContainer          *ioc.ComponentContainer
	Logger                      logger.Logger
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

	http.Handle("/", http.HandlerFunc(qhs.handleAll))

	listenAddress := fmt.Sprintf(":%d", qhs.Config.Port)

	go http.ListenAndServe(listenAddress, nil)

	qhs.Logger.LogInfof("HTTP server started listening on %d", qhs.Config.Port)

	return nil

}

func (h *QuiltHttpServer) handleAll(responseWriter http.ResponseWriter, request *http.Request) {

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
			handlerPattern.Provider.ServeHTTP(responseWriter, request)
		}

	}

}

type HttpServerConfig struct {
	Port        int
	ContentType string
	Encoding    string
}

func ParseDefaultHttpServerConfig(configAccessor *config.ConfigAccessor) HttpServerConfig {
	return ParseHttpServerConfig(configAccessor, defaultHttpServerConfigBase)
}

func ParseHttpServerConfig(configAccessor *config.ConfigAccessor, baseConfigPath string) HttpServerConfig {

	pathSep := config.JsonPathSeparator
	var httpServerConfig HttpServerConfig

	httpServerConfig.Port = configAccessor.IntValue(baseConfigPath + pathSep + httpListenPortPath)
	httpServerConfig.ContentType = configAccessor.StringVal(baseConfigPath + pathSep + httpContentTypePath)
	httpServerConfig.Encoding = configAccessor.StringVal(baseConfigPath + pathSep + httpEncodingPath)

	return httpServerConfig
}
