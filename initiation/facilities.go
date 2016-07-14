package initiation

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/decorator"
	"github.com/wolferton/quilt/facility/httpserver"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/querymanager"
	"github.com/wolferton/quilt/facility/rdbms"
	"github.com/wolferton/quilt/facility/serviceerror"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/ws/json"
)

const applicationLoggingManagerName = ioc.FrameworkPrefix + "ApplicationLoggingManager"
const frameworkLoggingManagerName = ioc.FrameworkPrefix + "FrameworkLoggingManager"
const applicationLoggingDecoratorName = ioc.FrameworkPrefix + "ApplicationLoggingDecorator"
const httpServerName = ioc.FrameworkPrefix + "HttpServer"
const accessLogWriterName = ioc.FrameworkPrefix + "AccessLogWriter"
const queryManagerName = ioc.FrameworkPrefix + "QueryManager"
const rdbmsClientManagerName = ioc.FrameworkPrefix + "RdbmsClientManager"

type FacilitiesInitialisor struct {
	ConfigAccessor          *config.ConfigAccessor
	FrameworkLoggingManager *logger.ComponentLoggerManager
	container               *ioc.ComponentContainer
}

func NewFacilitiesInitialisor(cc *ioc.ComponentContainer, flm *logger.ComponentLoggerManager) *FacilitiesInitialisor {
	fi := new(FacilitiesInitialisor)
	fi.container = cc
	fi.FrameworkLoggingManager = flm

	return fi
}

func BootstrapFrameworkLogging(bootStrapLogLevel int) (*logger.ComponentLoggerManager, *ioc.ProtoComponent) {

	flm := logger.CreateComponentLoggerManager(bootStrapLogLevel, nil)
	proto := ioc.CreateProtoComponent(flm, frameworkLoggingManagerName)

	return flm, proto

}

func (fi *FacilitiesInitialisor) Initialise(ca *config.ConfigAccessor) {
	fi.ConfigAccessor = ca

	fc := ca.ObjectVal("Facilities")
	fi.updateFrameworkLogLevel()

	if fc["ApplicationLogging"].(bool) {
		fi.initialiseApplicationLogger()
	}

	if fc["HttpServer"].(bool) {
		fi.initialiseHttpServer()
	}

	if fc["QueryManager"].(bool) {
		fi.initialiseQueryManager()
	}

	if fc["RdbmsAccess"].(bool) {
		fi.initialiseDatabaseAccessor()
	}

	if fc["JsonWs"].(bool) {
		fb := new(json.JsonWsFacilityBuilder)
		fb.BuildAndRegister(fi.FrameworkLoggingManager, fi.ConfigAccessor, fi.container)
	}

	if fc["ServiceErrorManager"].(bool) {
		serviceerror.InitialiseServiceErrorManager(fi.FrameworkLoggingManager, fi.ConfigAccessor, fi.container)
	}

}

func (fi *FacilitiesInitialisor) updateFrameworkLogLevel() {
	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("FrameworkLogger.DefaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	fi.FrameworkLoggingManager.UpdateGlobalThreshold(defaultLogLevel)
	fi.FrameworkLoggingManager.UpdateLocalThreshold(defaultLogLevel)
}

func (fi *FacilitiesInitialisor) initialiseApplicationLogger() {

	c := fi.container

	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("ApplicationLogger.DefaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	initialLogLevelsByComponent := fi.ConfigAccessor.ObjectVal("ApplicationLogger.ComponentLogLevels")

	applicationLoggingManager := logger.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)
	c.WrapAndAddProto(applicationLoggingManagerName, applicationLoggingManager)

	applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
	applicationLoggingDecorator.LoggerManager = applicationLoggingManager
	applicationLoggingDecorator.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(applicationLoggingDecoratorName)

	c.WrapAndAddProto(applicationLoggingDecoratorName, applicationLoggingDecorator)

}

func (fi *FacilitiesInitialisor) initialiseHttpServer() {

	c := fi.container

	httpServer := new(httpserver.QuiltHttpServer)
	fi.ConfigAccessor.Populate("HttpServer", httpServer)

	httpServer.Logger = fi.FrameworkLoggingManager.CreateLogger(httpServerName)

	c.WrapAndAddProto(httpServerName, httpServer)

	if !httpServer.AccessLogging {
		return
	}

	accessLogWriter := new(httpserver.AccessLogWriter)
	fi.ConfigAccessor.Populate("HttpServer.AccessLog", accessLogWriter)

	httpServer.AccessLogWriter = accessLogWriter

	c.WrapAndAddProto(accessLogWriterName, accessLogWriter)

}

func (fi *FacilitiesInitialisor) initialiseQueryManager() {
	queryManager := new(querymanager.QueryManager)
	queryManager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(queryManagerName)
	fi.ConfigAccessor.Populate("QueryManager", queryManager)

	fi.container.WrapAndAddProto(queryManagerName, queryManager)
}

func (fi *FacilitiesInitialisor) initialiseDatabaseAccessor() {
	manager := new(rdbms.DefaultRdbmsClientManager)
	manager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(rdbmsClientManagerName)
	fi.ConfigAccessor.Populate("RdbmsAccess", manager)

	proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

	proto.AddDependency("Provider", manager.DatabaseProviderComponentName)
	proto.AddDependency("QueryManager", queryManagerName)

	fi.container.AddProto(proto)

}
