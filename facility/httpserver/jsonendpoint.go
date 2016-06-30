package httpserver

import (
	"github.com/wolferton/quilt/facility/logger"
	"net/http"
)

//Implements HttpEndpointProvider
type JsonHandler struct {
	Unmarshaller           JsonUnmarshaller
	HttpMethod             string
	HttpMethods            []string
	PathMatchPattern       string
	Logic                  JsonRequestLogic
	ResponseWriter         JsonResponseWriter
	ErrorResponseWriter    JsonErrorResponseWriter
	QuiltApplicationLogger logger.Logger
}

//HttpEndpointProvider
func (jh *JsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	jsonReq, err := jh.Unmarshaller.Unmarshall(req)

	if err != nil {
		jh.unmarshallError(err, w)
		return
	}

	var errors ServiceErrors

	logic := jh.Logic

	logic.Validate(&errors, jsonReq)

	if errors.HasErrors() {
		jh.errorResponse(&errors, w)
		return
	}

	jsonRes := logic.Process(jsonReq)
	jh.writeResponse(jsonRes, w)

}

//HttpEndpointProvider
func (jh *JsonHandler) SupportedHttpMethods() []string {
	if len(jh.HttpMethods) > 0 {
		return jh.HttpMethods
	} else {
		return []string{jh.HttpMethod}
	}
}

//HttpEndpointProvider
func (jh *JsonHandler) RegexPattern() string {
	return jh.PathMatchPattern
}

func (jh *JsonHandler) unmarshallError(err error, w http.ResponseWriter) {

}

func (jh *JsonHandler) errorResponse(errors *ServiceErrors, w http.ResponseWriter) {

}

func (jh *JsonHandler) writeResponse(res *JsonResponse, w http.ResponseWriter) {

	jh.QuiltApplicationLogger.LogTrace("Write")

	errors := res.Errors

	if errors.HasErrors() {
		jh.errorResponse(errors, w)
		return
	}
}

func NewJsonResponse() *JsonResponse {
	r := new(JsonResponse)
	r.Errors = new(ServiceErrors)
	return r
}

type JsonRequest struct {
	PathParameters map[string]string
	HttpMethod     string
}

type JsonResponse struct {
	HttpStatus int
	Body       interface{}
	Errors     *ServiceErrors
}

type JsonRequestLogic interface {
	Validate(errors *ServiceErrors, request *JsonRequest)
	Process(request *JsonRequest) *JsonResponse
}

type JsonUnmarshaller interface {
	Unmarshall(req *http.Request) (*JsonRequest, error)
}

type DefaultJsonUnmarshaller struct {
	FrameworkLogger logger.Logger
}

func (jdu *DefaultJsonUnmarshaller) Unmarshall(httpReq *http.Request) (*JsonRequest, error) {

	var jsonReq JsonRequest

	jsonReq.HttpMethod = httpReq.Method

	return &jsonReq, nil

}

type JsonResponseWriter interface {
	Write(res *JsonResponse, request *JsonRequest, w http.ResponseWriter)
}

type DefaultJsonResponseWriter struct {
	FrameworkLogger logger.Logger
}

func (djrw *DefaultJsonResponseWriter) Write(res *JsonResponse, request *JsonRequest, w http.ResponseWriter) {
	djrw.FrameworkLogger.LogInfo("Writing response")
}

type JsonErrorResponseWriter interface {
	Write(errors *ServiceErrors, request *JsonRequest, w http.ResponseWriter)
}

type DefaultJsonErrorResponseWriter struct {
	FrameworkLogger logger.Logger
}

func (djerw *DefaultJsonErrorResponseWriter) Write(errors *ServiceErrors, request *JsonRequest, w http.ResponseWriter) {
	djerw.FrameworkLogger.LogInfo("Writing error response")
}
