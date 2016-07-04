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

		return []*ioc.ProtoComponent{responseWriterProto, responseErrorWriterProto, statusDeterminerProto, jsonUnmarshallerProto}
	}
}
