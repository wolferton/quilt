package json

import (
	"encoding/json"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ws"
	"net/http"
)

type DefaultJsonUnmarshaller struct {
	FrameworkLogger logger.Logger
}

func (jdu *DefaultJsonUnmarshaller) Unmarshall(httpReq *http.Request, logic interface{}) (*ws.WsRequest, error) {

	var wsReq ws.WsRequest

	targetSource, found := logic.(ws.WsUnmarshallTarget)

	if found {
		target := targetSource.UnmarshallTarget()
		err := json.NewDecoder(httpReq.Body).Decode(&target)

		if err != nil {
			return nil, err
		}

		wsReq.RequestBody = target

	}

	wsReq.HttpMethod = httpReq.Method

	return &wsReq, nil

}

type DefaultJsonResponseWriter struct {
	FrameworkLogger logger.Logger
}

func (djrw *DefaultJsonResponseWriter) Write(res *ws.WsResponse, w http.ResponseWriter) error {

	if res.Body == nil {
		return nil
	}

	wrapper := wrapJsonResponse(nil, res.Body)

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

type DefaultAbnormalResponseWriter struct {
	FrameworkLogger logger.Logger
}

func (djerw *DefaultAbnormalResponseWriter) WriteWithErrors(status int, errors *ws.ServiceErrors, w http.ResponseWriter) error {

	wrapper := wrapJsonResponse(djerw.formatErrors(errors), nil)

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	w.WriteHeader(status)

	return err
}

func (djerw *DefaultAbnormalResponseWriter) formatErrors(errors *ws.ServiceErrors) interface{} {

	f := make(map[string]string)

	for _, error := range errors.Errors {

		var c string

		switch error.Category {
		default:
			c = "?"
		case ws.Unexpected:
			c = "U"
		case ws.Security:
			c = "S"
		case ws.Logic:
			c = "L"
		case ws.Client:
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
