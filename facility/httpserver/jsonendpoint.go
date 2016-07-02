package httpserver

import (
	"encoding/json"
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
	logic := jh.Logic
	jsonReq, err := jh.Unmarshaller.Unmarshall(req, logic)

	if err != nil {
		jh.unmarshallError(err, w)
		return
	}

	var errors ServiceErrors

	logic.Validate(&errors, jsonReq)

	if errors.HasErrors() {
		jh.errorResponse(&errors, w)
		return
	}

	jsonRes := logic.Process(jsonReq)

	err = jh.writeResponse(jsonRes, w)

	if err != nil {
		jh.QuiltApplicationLogger.LogErrorf("Problem writing a HTTP response to %s after processing was complete: %s ", req.RequestURI, err)
	}

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

func (jh *JsonHandler) writeResponse(res *JsonResponse, w http.ResponseWriter) error {

	errors := res.Errors

	if errors.HasErrors() {
		jh.errorResponse(errors, w)
		return nil
	}

	err := jh.ResponseWriter.Write(res, w)

	return err
}

func NewJsonResponse() *JsonResponse {
	r := new(JsonResponse)
	r.Errors = new(ServiceErrors)
	return r
}

type JsonRequest struct {
	PathParameters map[string]string
	HttpMethod     string
	RequestBody    interface{}
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

type JsonUnmarshallTarget interface {
	UnmarshallTarget() interface{}
}

type JsonUnmarshaller interface {
	Unmarshall(req *http.Request, logic interface{}) (*JsonRequest, error)
}

type DefaultJsonUnmarshaller struct {
	FrameworkLogger logger.Logger
}

func (jdu *DefaultJsonUnmarshaller) Unmarshall(httpReq *http.Request, logic interface{}) (*JsonRequest, error) {

	var jsonReq JsonRequest

	targetSource, found := logic.(JsonUnmarshallTarget)

	if found {
		target := targetSource.UnmarshallTarget()
		err := json.NewDecoder(httpReq.Body).Decode(&target)

		if err != nil {
			return nil, err
		}

		jsonReq.RequestBody = target

	}

	jsonReq.HttpMethod = httpReq.Method

	return &jsonReq, nil

}

type JsonResponseWriter interface {
	Write(res *JsonResponse, w http.ResponseWriter) error
}

type DefaultJsonResponseWriter struct {
	FrameworkLogger logger.Logger
}

func (djrw *DefaultJsonResponseWriter) Write(res *JsonResponse, w http.ResponseWriter) error {
	data, err := json.Marshal(res.Body)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
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
