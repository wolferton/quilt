package json

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/ws"
)

const jsonResponseWriterComponentName = ioc.FrameworkPrefix + "JsonResponseWriter"
const jsonUnmarshallerComponentName = ioc.FrameworkPrefix + "JsonUnmarshaller"
const jsonAbnormalResponseWriterComponentName = ioc.FrameworkPrefix + "JsonAbnormalResponseWriter"
const jsonHandlerDecoratorComponentName = ioc.FrameworkPrefix + "JsonHandlerDecorator"
const wsHttpStatusDeterminerComponentName = ioc.FrameworkPrefix + "HttpStatusDeterminer"

func InitialiseJsonHttp(logManager *logger.ComponentLoggerManager, config *config.ConfigAccessor, protoComponents map[string]*ioc.ProtoComponent) {

	responseWriter := new(DefaultJsonResponseWriter)
	responseWriter.FrameworkLogger = logManager.CreateLogger(jsonResponseWriterComponentName)
	responseWriterProto := ioc.CreateProtoComponent(responseWriter, jsonResponseWriterComponentName)
	protoComponents[jsonResponseWriterComponentName] = responseWriterProto

	abnormalResponseWriter := new(DefaultAbnormalResponseWriter)
	abnormalResponseWriter.FrameworkLogger = logManager.CreateLogger(jsonAbnormalResponseWriterComponentName)
	abnormalResponseWriterProto := ioc.CreateProtoComponent(abnormalResponseWriter, jsonAbnormalResponseWriterComponentName)
	protoComponents[jsonAbnormalResponseWriterComponentName] = abnormalResponseWriterProto

	statusDeterminer := new(ws.DefaultHttpStatusCodeDeterminer)
	statusDeterminerProto := ioc.CreateProtoComponent(statusDeterminer, wsHttpStatusDeterminerComponentName)
	protoComponents[wsHttpStatusDeterminerComponentName] = statusDeterminerProto

	jsonUnmarshaller := new(DefaultJsonUnmarshaller)
	jsonUnmarshaller.FrameworkLogger = logManager.CreateLogger(jsonUnmarshallerComponentName)
	jsonUnmarshallerProto := ioc.CreateProtoComponent(jsonUnmarshaller, jsonUnmarshallerComponentName)
	protoComponents[jsonUnmarshallerComponentName] = jsonUnmarshallerProto

	decoratorLogger := logManager.CreateLogger(jsonHandlerDecoratorComponentName)
	decorator := JsonWsHandlerDecorator{decoratorLogger, responseWriter, abnormalResponseWriter, statusDeterminer, jsonUnmarshaller}
	decoratorProto := ioc.CreateProtoComponent(&decorator, jsonHandlerDecoratorComponentName)
	protoComponents[jsonHandlerDecoratorComponentName] = decoratorProto
}

type JsonWsHandlerDecorator struct {
	FrameworkLogger      logger.Logger
	ResponseWriter       ws.WsResponseWriter
	ErrorResponseWriter  ws.WsAbnormalResponseWriter
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
