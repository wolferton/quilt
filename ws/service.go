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
	PathParameters  map[string]string
	HttpMethod      string
	RequestBody     interface{}
	QueryParams     *WsQueryParams
	FrameworkErrors []*WsFrameworkError
}

func (wsr *WsRequest) AddFrameworkError(f *WsFrameworkError) {
	wsr.FrameworkErrors = append(wsr.FrameworkErrors, f)
}

type WsResponse struct {
	HttpStatus int
	Body       interface{}
	Errors     *ServiceErrors
}

type WsFrameworkPhase int

const (
	Unmarshall = iota
	QueryBind
	PathBind
)

type WsFrameworkError struct {
	Phase       WsFrameworkPhase
	ClientField string
	TargetField string
	Message     string
}

func NewUnmarshallWsFrameworkError(message string) *WsFrameworkError {
	f := new(WsFrameworkError)
	f.Phase = Unmarshall
	f.Message = message

	return f
}

func NewQueryBindFrameworkError(message string, param string, target string) *WsFrameworkError {
	f := new(WsFrameworkError)
	f.Phase = QueryBind
	f.Message = message
	f.ClientField = param
	f.TargetField = target

	return f
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
	Unmarshall(req *http.Request, wsReq *WsRequest) error
}

type WsResponseWriter interface {
	Write(res *WsResponse, w http.ResponseWriter) error
}

type WsAbnormalResponseWriter interface {
	WriteWithErrors(status int, errors *ServiceErrors, w http.ResponseWriter) error
}
