package ws

import (
	"net/http"
)

func NewWsResponse(errorFinder ServiceErrorFinder) *WsResponse {
	r := new(WsResponse)
	r.Errors = new(ServiceErrors)
	r.Errors.ErrorFinder = errorFinder

	return r
}

type WsRequest struct {
	PathParameters map[string]string
	HttpMethod     string
	RequestBody    interface{}
	QueryParams    *WsQueryParams
}

type WsResponse struct {
	HttpStatus int
	Body       interface{}
	Errors     *ServiceErrors
}

type WsRequestProcessor interface {
	Process(request *WsRequest, response *WsResponse)
}

type WsRequestValidator interface {
	Validate(errors *ServiceErrors, request *WsRequest)
}

type WsUnmarshallTarget interface {
	UnmarshallTarget() interface{}
}

type WsUnmarshaller interface {
	Unmarshall(req *http.Request, logic interface{}) (*WsRequest, error)
}

type WsResponseWriter interface {
	Write(res *WsResponse, w http.ResponseWriter) error
}

type WsErrorResponseWriter interface {
	Write(errors *ServiceErrors, w http.ResponseWriter) error
}
