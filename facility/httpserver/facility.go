package httpserver

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
)

const httpServerName = ioc.FrameworkPrefix + "HttpServer"
const accessLogWriterName = ioc.FrameworkPrefix + "AccessLogWriter"

type HttpServerFacilityBuilder struct {
}

func (hsfb *HttpServerFacilityBuilder) BuildAndRegister(lm *logger.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) {

	httpServer := new(HttpServer)
	ca.Populate("HttpServer", httpServer)

	cn.WrapAndAddProto(httpServerName, httpServer)

	if !httpServer.AccessLogging {
		return
	}

	accessLogWriter := new(AccessLogWriter)
	ca.Populate("HttpServer.AccessLog", accessLogWriter)

	httpServer.AccessLogWriter = accessLogWriter

	cn.WrapAndAddProto(accessLogWriterName, accessLogWriter)

}

func (hsfb *HttpServerFacilityBuilder) FacilityName() string {
	return "HttpServer"
}
