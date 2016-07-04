package json

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/ws"
)

const jsonResponseWriterComponentName = "quiltJsonResponseWriter"
const jsonUnmarshallerComponentName = "quiltJsonUnmarshaller"
const jsonErrorResponseWriterComponentName = "quiltJsonErrorResponseWriter"
const jsonHandlerDecoratorComponentName = "quiltJsonHandlerDecorator"
const wsHttpStatusDeterminerComponentName = "quiltHttpStatusDeterminer"

func InitialiseJsonHttp(logManager *logger.ComponentLoggerManager, config *config.ConfigAccessor) []*ioc.ProtoComponent {

	if !config.BoolValue("facilities.jsonWs.enabled") {
		return []*ioc.ProtoComponent{}
	} else {

		responseWriter := new(DefaultJsonResponseWriter)
		responseWriter.FrameworkLogger = logManager.CreateLogger(jsonResponseWriterComponentName)
		responseWriterProto := ioc.CreateProtoComponent(responseWriter, jsonResponseWriterComponentName)

		responseErrorWriter := new(DefaultJsonErrorResponseWriter)
		responseErrorWriter.FrameworkLogger = logManager.CreateLogger(jsonErrorResponseWriterComponentName)
		responseErrorWriterProto := ioc.CreateProtoComponent(responseErrorWriter, jsonErrorResponseWriterComponentName)

		statusDeterminer := new(ws.DefaultHttpStatusCodeDeterminer)
		statusDeterminerProto := ioc.CreateProtoComponent(statusDeterminer, wsHttpStatusDeterminerComponentName)

		jsonUnmarshaller := new(DefaultJsonUnmarshaller)
		jsonUnmarshaller.FrameworkLogger = logManager.CreateLogger(jsonUnmarshallerComponentName)
		jsonUnmarshallerProto := ioc.CreateProtoComponent(jsonUnmarshaller, jsonUnmarshallerComponentName)

		decoratorLogger := logManager.CreateLogger(jsonHandlerDecoratorComponentName)
		decorator := JsonWsHandlerDecorator{decoratorLogger, responseWriter, responseErrorWriter, statusDeterminer, jsonUnmarshaller}
		decoratorProto := ioc.CreateProtoComponent(&decorator, jsonHandlerDecoratorComponentName)

		return []*ioc.ProtoComponent{responseWriterProto, responseErrorWriterProto, statusDeterminerProto, jsonUnmarshallerProto, decoratorProto}
	}
}

type JsonWsHandlerDecorator struct {
	FrameworkLogger      logger.Logger
	ResponseWriter       ws.WsResponseWriter
	ErrorResponseWriter  ws.WsErrorResponseWriter
	StatusCodeDeterminer ws.HttpStatusCodeDeterminer
	Unmarshaller         ws.WsUnmarshaller
}

func (jwhd *JsonWsHandlerDecorator) OfInterest(component *ioc.Component) bool {
	switch component.Instance.(type) {
	default:
		jwhd.FrameworkLogger.LogTracef("No interest %s", component.Name)
		return false
	case *ws.WsHandler:
		return true
	}
}

func (jwhd *JsonWsHandlerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	h := component.Instance.(*ws.WsHandler)
	l := jwhd.FrameworkLogger
	l.LogTracef("Decorating component %s", component.Name)

	if h.StatusDeterminer == nil {
		h.StatusDeterminer = jwhd.StatusCodeDeterminer
	}

	if h.ResponseWriter == nil {
		h.ResponseWriter = jwhd.ResponseWriter
	}

	if h.ErrorResponseWriter == nil {
		h.ErrorResponseWriter = jwhd.ErrorResponseWriter
	}

	if h.Unmarshaller == nil {
		l.LogTracef("%s needs Unmarshaller", component.Name)
		h.Unmarshaller = jwhd.Unmarshaller
	}

}
