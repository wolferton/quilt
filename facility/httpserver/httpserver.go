package httpserver
import (
    "github.com/wolferton/quilt/config"
    "fmt"
    "net/http"
    "github.com/wolferton/quilt/ioc"
    "regexp"
    "github.com/wolferton/quilt/facility/logger"
)

const defaultHttpServerConfigBase = "facilities.httpServer"
const httpListenPortPath = "listenPort"
const httpContentTypePath = "contentType"
const httpEncodingPath = "encoding"

type HandlerPattern struct {
    Handler http.Handler
    Pattern *regexp.Regexp
}

type QuiltHttpServer struct {
    Config HttpServerConfig
    methodsToHandlerPatterns map[string][]*HandlerPattern
    componentContainer *ioc.ComponentContainer
    Logger logger.Logger
}

func (qhs *QuiltHttpServer) Container(container *ioc.ComponentContainer) {
    qhs.componentContainer = container
}

func (qhs *QuiltHttpServer) mapHandler(endPoint *HttpEndPoint) {

    handler := endPoint.Handler

    for method, pattern := range endPoint.MethodPatterns {

        compiledRegex, regexError := regexp.Compile(pattern)

        if(regexError != nil) {
            errorMessage := fmt.Sprintf("Unable to compile regular expression from pattern %s: %s", pattern, regexError.Error())
            qhs.Logger.LogError(errorMessage)
        }

        handlerPattern := HandlerPattern{handler, compiledRegex}

        sameMethod := qhs.methodsToHandlerPatterns[method]

        if(sameMethod == nil) {
            sameMethod = make([]*HandlerPattern,1)
            sameMethod[0] = &handlerPattern
            qhs.methodsToHandlerPatterns[method] = sameMethod
        } else {
            qhs.methodsToHandlerPatterns[method] = append(sameMethod,&handlerPattern)
        }
    }

}

func (qhs *QuiltHttpServer) StartComponent() {

    qhs.methodsToHandlerPatterns = make(map[string][]*HandlerPattern)

    endpoints := qhs.componentContainer.FindByType("*httpserver.HttpEndPoint")

    //log.Printf("Found %d HTTP handlers in container", len(endpoints))

    for _, endpointInterface := range endpoints {

        endpoint := endpointInterface.(*HttpEndPoint)
        qhs.mapHandler(endpoint)

    }

    startMessage := fmt.Sprintf("Starting HTTP server listening on %d\n", qhs.Config.Port)
    qhs.Logger.LogInfo(startMessage)

    http.Handle("/", http.HandlerFunc(qhs.handleAll))

    listenAddress := fmt.Sprintf(":%d", qhs.Config.Port)

    err := http.ListenAndServe(listenAddress, nil)

    if err != nil {
        //log.Fatal("ListenAndServe:", err)
    }
}

func (h *QuiltHttpServer) handleAll(responseWriter http.ResponseWriter, request *http.Request) {

    contentType := fmt.Sprintf("%s; charset=%s", h.Config.ContentType, h.Config.Encoding)
    responseWriter.Header().Set("Content-Type", contentType)

    methodHandlers := h.methodsToHandlerPatterns[request.Method]

    for _, handlerPattern := range methodHandlers {

        pattern := handlerPattern.Pattern

        if(pattern.MatchString(request.URL.Path)){
            handlerPattern.Handler.ServeHTTP(responseWriter,request)
        }


    }


    //responseWriter.WriteHeader(404)
}



type HttpServerConfig struct {
    Port int
    ContentType string
    Encoding string
}


func ParseDefaultHttpServerConfig(configAccessor *config.ConfigAccessor) HttpServerConfig {
    return ParseHttpServerConfig(configAccessor, defaultHttpServerConfigBase)
}

func ParseHttpServerConfig(configAccessor *config.ConfigAccessor, baseConfigPath string) HttpServerConfig{

    pathSep := config.JsonPathSeparator
    var httpServerConfig HttpServerConfig

    httpServerConfig.Port = configAccessor.IntValue(baseConfigPath + pathSep + httpListenPortPath)
    httpServerConfig.ContentType = configAccessor.StringVal(baseConfigPath + pathSep + httpContentTypePath)
    httpServerConfig.Encoding = configAccessor.StringVal(baseConfigPath + pathSep + httpEncodingPath)

    return httpServerConfig
}