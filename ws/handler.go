package ws

import (
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
}

//HttpEndpointProvider
func (wh *WsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logic := wh.Logic
	jsonReq, err := wh.Unmarshaller.Unmarshall(req, logic)
	l := wh.QuiltApplicationLogger

	if err != nil {
		wh.unmarshallError(err, w)
		return
	}

	var errors ServiceErrors

	validator, found := logic.(WsRequestValidator)

	if found {
		validator.Validate(&errors, jsonReq)
	}

	if errors.HasErrors() {
		err = wh.errorResponse(&errors, w)

		if err != nil {
			l.LogErrorf("Problem writing an HTTP response for request after failed validation %s: %s", req.URL, err)
		}
	} else {

		wsRes := NewWsResponse()
		logic.Process(jsonReq, wsRes)

		err = wh.writeResponse(wsRes, w)

		if err != nil {
			l.LogErrorf("Problem writing an HTTP response for request the request was processed %s: %s", req.URL, err)
		}
	}

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

func (wh *WsHandler) unmarshallError(err error, w http.ResponseWriter) {
	wh.QuiltApplicationLogger.LogErrorf("Error unmarshalling request body %s", err)

}

func (wh *WsHandler) errorResponse(errors *ServiceErrors, w http.ResponseWriter) error {

	status := wh.StatusDeterminer.DetermineCodeFromErrors(errors)

	w.WriteHeader(status)
	err := wh.ErrorResponseWriter.Write(errors, w)

	return err

}

func (wh *WsHandler) writeResponse(res *WsResponse, w http.ResponseWriter) error {

	errors := res.Errors

	if errors.HasErrors() {
		err := wh.errorResponse(errors, w)
		return err
	}

	status := wh.StatusDeterminer.DetermineCode(res)
	w.WriteHeader(status)

	err := wh.ResponseWriter.Write(res, w)
	return err
}
