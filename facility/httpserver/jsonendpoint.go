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
	StatusDeterminer       HttpStatusCodeDeterminer
}

//HttpEndpointProvider
func (jh *JsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logic := jh.Logic
	jsonReq, err := jh.Unmarshaller.Unmarshall(req, logic)
	l := jh.QuiltApplicationLogger

	if err != nil {
		jh.unmarshallError(err, w)
		return
	}

	var errors ServiceErrors

	validator, found := logic.(JsonRequestValidator)

	if found {
		validator.Validate(&errors, jsonReq)
	}

	if errors.HasErrors() {
		err = jh.errorResponse(&errors, w)

		if err != nil {
			l.LogErrorf("Problem writing an HTTP response for request after failed validation %s: %s", req.URL, err)
		}
	} else {

		jsonRes := logic.Process(jsonReq)

		err = jh.writeResponse(jsonRes, w)

		if err != nil {
			l.LogErrorf("Problem writing an HTTP response for request the request was processed %s: %s", req.URL, err)
		}
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
	jh.QuiltApplicationLogger.LogErrorf("Error unmarshalling request body %s", err)

}

func (jh *JsonHandler) errorResponse(errors *ServiceErrors, w http.ResponseWriter) error {

	status := jh.StatusDeterminer.DetermineCodeFromErrors(errors)

	w.WriteHeader(status)
	err := jh.ErrorResponseWriter.Write(errors, w)

	return err

}

func (jh *JsonHandler) writeResponse(res *JsonResponse, w http.ResponseWriter) error {

	errors := res.Errors

	if errors.HasErrors() {
		err := jh.errorResponse(errors, w)
		return err
	}

	status := jh.StatusDeterminer.DetermineCode(res)
	w.WriteHeader(status)

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

type HttpStatusCodeDeterminer interface {
	DetermineCode(response *JsonResponse) int
	DetermineCodeFromErrors(errors *ServiceErrors) int
}

type DefaultHttpStatusCodeDeterminer struct {
}

func (dhscd *DefaultHttpStatusCodeDeterminer) DetermineCode(response *JsonResponse) int {
	if response.HttpStatus != 0 {
		return response.HttpStatus

	} else {
		return http.StatusOK
	}
}

func (dhscd *DefaultHttpStatusCodeDeterminer) DetermineCodeFromErrors(errors *ServiceErrors) int {

	if errors.HttpStatus != 0 {
		return errors.HttpStatus
	}

	sCount := 0
	cCount := 0
	lCount := 0

	for _, error := range errors.Errors {

		switch error.Category {
		case Unexpected:
			return http.StatusInternalServerError
		case Security:
			sCount += 1
		case Logic:
			lCount += 1
		case Client:
			cCount += 1
		}
	}

	if sCount > 0 {
		return http.StatusUnauthorized
	}

	if cCount > 0 {
		return http.StatusBadRequest
	}

	if lCount > 0 {
		return http.StatusConflict
	}

	return http.StatusOK
}

type JsonRequestLogic interface {
	Process(request *JsonRequest) *JsonResponse
}

type JsonRequestValidator interface {
	Validate(errors *ServiceErrors, request *JsonRequest)
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

	wrapper := wrapJsonResponse(nil, res.Body)

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

type JsonErrorResponseWriter interface {
	Write(errors *ServiceErrors, w http.ResponseWriter) error
}

type DefaultJsonErrorResponseWriter struct {
	FrameworkLogger logger.Logger
}

func (djerw *DefaultJsonErrorResponseWriter) Write(errors *ServiceErrors, w http.ResponseWriter) error {

	wrapper := wrapJsonResponse(djerw.formatErrors(errors), nil)

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func (djerw *DefaultJsonErrorResponseWriter) formatErrors(errors *ServiceErrors) interface{} {

	f := make(map[string]string)

	for _, error := range errors.Errors {

		var c string

		switch error.Category {
		default:
			c = "?"
		case Unexpected:
			c = "U"
		case Security:
			c = "S"
		case Logic:
			c = "L"
		case Client:
			c = "C"
		}

		f[c+"-"+error.Label] = error.Message
	}

	return f
}

func wrapJsonResponse(errors interface{}, body interface{}) interface{} {
	f := make(map[string]interface{})

	if errors != nil {
		f["Errors"] = errors
	}

	if body != nil {
		f["Response"] = body
	}

	return f
}
