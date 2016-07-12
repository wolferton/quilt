package ws

import (
	"fmt"
	"github.com/wolferton/quilt/facility/logger"
	"net/http"
)

//Implements HttpEndpointProvider
type WsHandler struct {
	Unmarshaller           WsUnmarshaller
	HttpMethod             string
	HttpMethods            []string
	PathMatchPattern       string
	Logic                  WsRequestProcessor
	ResponseWriter         WsResponseWriter
	ErrorResponseWriter    WsErrorResponseWriter
	QuiltApplicationLogger logger.Logger
	StatusDeterminer       HttpStatusCodeDeterminer
	ErrorFinder            ServiceErrorFinder
	RevealPanicDetails     bool
	DisableQueryParsing    bool
}

func (wh *WsHandler) ProvideErrorFinder(finder ServiceErrorFinder) {
	wh.ErrorFinder = finder
}

//HttpEndpointProvider
func (wh *WsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			wh.writePanicResponse(r, w)
		}
	}()

	logic := wh.Logic
	wsReq, err := wh.Unmarshaller.Unmarshall(req, logic)

	if err != nil {
		wh.writeUnmarshallError(err, w)
		return
	}

	if !wh.DisableQueryParsing {
		wh.processQueryParams(req, wsReq)
	}

	var errors ServiceErrors
	errors.ErrorFinder = wh.ErrorFinder

	validator, found := logic.(WsRequestValidator)

	if found {
		validator.Validate(&errors, wsReq)
	}

	if errors.HasErrors() {
		wh.writeErrorResponse(&errors, w)
	} else {

		wh.process(wsReq, w)

	}

}

func (wh *WsHandler) processQueryParams(req *http.Request, wsReq *WsRequest) {

	values := req.URL.Query()
	wsReq.QueryParams = NewWsQueryParams(values)

}

//HttpEndpointProvider
func (wh *WsHandler) SupportedHttpMethods() []string {
	if len(wh.HttpMethods) > 0 {
		return wh.HttpMethods
	} else {
		return []string{wh.HttpMethod}
	}
}

//HttpEndpointProvider
func (wh *WsHandler) RegexPattern() string {
	return wh.PathMatchPattern
}

func (wh *WsHandler) writeUnmarshallError(err error, w http.ResponseWriter) {
	wh.QuiltApplicationLogger.LogErrorf("Error unmarshalling request body %s", err)

	var se ServiceErrors
	se.HttpStatus = http.StatusBadRequest

	message := fmt.Sprintf("There is a problem with the body of the request: %s", err)
	se.AddError(Client, "UNMARSH", message)

	wh.writeErrorResponse(&se, w)

}

func (wh *WsHandler) process(jsonReq *WsRequest, w http.ResponseWriter) {

	defer func() {
		if r := recover(); r != nil {
			wh.writePanicResponse(r, w)
		}
	}()

	wsRes := NewWsResponse(wh.ErrorFinder)
	wh.Logic.Process(jsonReq, wsRes)

	errors := wsRes.Errors

	if errors.HasErrors() {
		wh.writeErrorResponse(errors, w)

	} else {
		status := wh.StatusDeterminer.DetermineCode(wsRes)
		w.WriteHeader(status)
	}

	wh.ResponseWriter.Write(wsRes, w)
}

func (wh *WsHandler) writeErrorResponse(errors *ServiceErrors, w http.ResponseWriter) {

	l := wh.QuiltApplicationLogger

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfWithTrace("Panic recovered while trying to write a response that was already in error %s", r)
		}
	}()

	status := wh.StatusDeterminer.DetermineCodeFromErrors(errors)

	w.WriteHeader(status)
	err := wh.ErrorResponseWriter.Write(errors, w)

	if err != nil {
		l.LogErrorf("Problem writing an HTTP response that was already in error", err)
	}

}

func (wh *WsHandler) writePanicResponse(r interface{}, w http.ResponseWriter) {

	var se ServiceErrors
	se.HttpStatus = http.StatusInternalServerError

	var message string

	if wh.RevealPanicDetails {
		message = fmt.Sprintf("Unhandled error %s", r)

	} else {
		message = "A unexpected error occured while processing this request."
	}

	wh.QuiltApplicationLogger.LogErrorf("Panic recovered but error response served. %s", r)

	se.AddError(Unexpected, "UNXP", message)

	wh.writeErrorResponse(&se, w)
}
