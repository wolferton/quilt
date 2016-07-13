package initiation

import (
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/decorator"
	"github.com/wolferton/quilt/facility/httpserver"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/querymanager"
	"github.com/wolferton/quilt/facility/rdbms"
	"github.com/wolferton/quilt/ioc"
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
	ConfigInjector          *config.ConfigInjector
	FrameworkLoggingManager *logger.ComponentLoggerManager
}

func (fi *FacilitiesInitialisor) BootstrapFrameworkLogging(protoComponents map[string]*ioc.ProtoComponent, bootStrapLogLevel int) *logger.ComponentLoggerManager {

	flm := logger.CreateComponentLoggerManager(bootStrapLogLevel, nil)
	proto := ioc.CreateProtoComponent(flm, frameworkLoggingManagerName)

	fi.FrameworkLoggingManager = flm

	protoComponents[frameworkLoggingManagerName] = proto

	return flm

}

func (fi *FacilitiesInitialisor) UpdateFrameworkLogLevel() {
	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("facilities.frameworkLogger.defaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	fi.FrameworkLoggingManager.UpdateGlobalThreshold(defaultLogLevel)
	fi.FrameworkLoggingManager.UpdateLocalThreshold(defaultLogLevel)
}

func (fi *FacilitiesInitialisor) InitialiseApplicationLogger(protoComponents map[string]*ioc.ProtoComponent) {

	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("facilities.applicationLogger.defaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	initialLogLevelsByComponent := fi.ConfigAccessor.ObjectVal("facilities.applicationLogger.componentLogLevels")

	applicationLoggingManager := logger.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)
	applicationLoggingMangagerProto := ioc.CreateProtoComponent(applicationLoggingManager, applicationLoggingManagerName)
	protoComponents[applicationLoggingManagerName] = applicationLoggingMangagerProto

	applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
	applicationLoggingDecorator.LoggerManager = applicationLoggingManager
	applicationLoggingDecorator.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(applicationLoggingDecoratorName)
	applicationLoggingDecoratorProto := ioc.CreateProtoComponent(applicationLoggingDecorator, applicationLoggingDecoratorName)

	protoComponents[applicationLoggingDecoratorName] = applicationLoggingDecoratorProto

}

func (fi *FacilitiesInitialisor) InitialiseHttpServer(protoComponents map[string]*ioc.ProtoComponent) {

	if !fi.ConfigAccessor.BoolValue("facilities.httpServer.enabled") {
		return
	}

	httpServerConfig := httpserver.ParseDefaultHttpServerConfig(fi.ConfigInjector)

	httpServer := new(httpserver.QuiltHttpServer)
	httpServer.Config = httpServerConfig
	httpServer.Logger = fi.FrameworkLoggingManager.CreateLogger(httpServerName)

	proto := ioc.CreateProtoComponent(httpServer, httpServerName)
	protoComponents[httpServerName] = proto

	if !fi.ConfigAccessor.BoolValue("facilities.httpServer.accessLog.enabled") {
		return
	}

	accessLogWriter := new(httpserver.AccessLogWriter)
	fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.httpServer.accessLog", accessLogWriter)

	httpServer.AccessLogWriter = accessLogWriter

	proto = ioc.CreateProtoComponent(accessLogWriter, accessLogWriterName)
	protoComponents[accessLogWriterName] = proto

}

func (fi *FacilitiesInitialisor) InitialiseQueryManager(protoComponents map[string]*ioc.ProtoComponent) {
	if !fi.ConfigAccessor.BoolValue("facilities.queryManager.enabled") {
		return
	} else {

		queryManager := new(querymanager.QueryManager)
		queryManager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(queryManagerName)
		fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.queryManager", queryManager)

		proto := ioc.CreateProtoComponent(queryManager, queryManagerName)
		protoComponents[queryManagerName] = proto
	}
}

func (fi *FacilitiesInitialisor) InitialiseDatabaseAccessor(protoComponents map[string]*ioc.ProtoComponent) {
	if !fi.ConfigAccessor.BoolValue("facilities.rdbmsAccess.enabled") {
		return
	} else {

		manager := new(rdbms.DefaultRdbmsClientManager)
		manager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(rdbmsClientManagerName)
		fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.rdbmsAccess", manager)

		proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

		proto.AddDependency("Provider", manager.DatabaseProviderComponentName)
		proto.AddDependency("QueryManager", queryManagerName)

		protoComponents[rdbmsClientManagerName] = proto

	}
}

type FacilityConfig struct {
	HttpServer          bool
	JsonWs              bool
	FrameworkLogging    bool
	ApplicationLogging  bool
	QueryManager        bool
	RdbmsAccess         bool
	ServiceErrorManager bool
}
