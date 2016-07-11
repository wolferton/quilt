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

func (fi *FacilitiesInitialisor) BootstrapFrameworkLogging(protoComponents []*ioc.ProtoComponent, bootStrapLogLevel int) ([]*ioc.ProtoComponent, *logger.ComponentLoggerManager) {

	frameworkLoggingManager := logger.CreateComponentLoggerManager(bootStrapLogLevel, nil)
	frameworkLoggingManagerProto := ioc.CreateProtoComponent(frameworkLoggingManager, frameworkLoggingManagerName)

	fi.FrameworkLoggingManager = frameworkLoggingManager

	return append(protoComponents, frameworkLoggingManagerProto), frameworkLoggingManager

}

func (fi *FacilitiesInitialisor) UpdateFrameworkLogLevel() {
	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("facilities.frameworkLogger.defaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	fi.FrameworkLoggingManager.UpdateGlobalThreshold(defaultLogLevel)
	fi.FrameworkLoggingManager.UpdateLocalThreshold(defaultLogLevel)
}

func (fi *FacilitiesInitialisor) InitialiseApplicationLogger(protoComponents []*ioc.ProtoComponent) []*ioc.ProtoComponent {

	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("facilities.applicationLogger.defaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	initialLogLevelsByComponent := fi.ConfigAccessor.ObjectVal("facilities.applicationLogger.componentLogLevels")

	applicationLoggingManager := logger.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)

	applicationLoggingMangagerProto := ioc.CreateProtoComponent(applicationLoggingManager, applicationLoggingManagerName)
	protoComponents = append(protoComponents, applicationLoggingMangagerProto)

	applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
	applicationLoggingDecorator.LoggerManager = applicationLoggingManager

	applicationLoggingDecorator.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger("ApplicationLogDecorator")

	applicationLoggingDecoratorProto := ioc.CreateProtoComponent(applicationLoggingDecorator, applicationLoggingDecoratorName)

	return append(protoComponents, applicationLoggingDecoratorProto)
}

func (fi *FacilitiesInitialisor) InitialiseHttpServer(protoComponents []*ioc.ProtoComponent, configAccessor *config.ConfigAccessor, frameworkLoggingManager *logger.ComponentLoggerManager) []*ioc.ProtoComponent {

	if !configAccessor.BoolValue("facilities.httpServer.enabled") {
		return protoComponents
	}

	httpServerConfig := httpserver.ParseDefaultHttpServerConfig(fi.ConfigInjector)

	httpServer := new(httpserver.QuiltHttpServer)
	httpServer.Config = httpServerConfig
	httpServer.Logger = frameworkLoggingManager.CreateLogger(httpServerName)

	proto := ioc.CreateProtoComponent(httpServer, httpServerName)
	protoComponents = append(protoComponents, proto)

	if !configAccessor.BoolValue("facilities.httpServer.accessLog.enabled") {
		return protoComponents
	}

	accessLogWriter := new(httpserver.AccessLogWriter)
	fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.httpServer.accessLog", accessLogWriter)

	httpServer.AccessLogWriter = accessLogWriter

	proto = ioc.CreateProtoComponent(accessLogWriter, accessLogWriterName)
	protoComponents = append(protoComponents, proto)

	return protoComponents
}

func (fi *FacilitiesInitialisor) InitialiseQueryManager(protoComponents []*ioc.ProtoComponent) []*ioc.ProtoComponent {
	if !fi.ConfigAccessor.BoolValue("facilities.queryManager.enabled") {
		return protoComponents
	} else {

		queryManager := new(querymanager.QueryManager)
		queryManager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(queryManagerName)
		fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.queryManager", queryManager)

		proto := ioc.CreateProtoComponent(queryManager, queryManagerName)

		return append(protoComponents, proto)
	}
}

func (fi *FacilitiesInitialisor) InitialiseDatabaseAccessor(protoComponents []*ioc.ProtoComponent) []*ioc.ProtoComponent {
	if !fi.ConfigAccessor.BoolValue("facilities.rdbmsAccess.enabled") {
		return protoComponents
	} else {

		manager := new(rdbms.DefaultRdbmsClientManager)
		manager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(rdbmsClientManagerName)
		fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.rdbmsAccess", manager)

		proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

		proto.AddDependency("Provider", manager.DatabaseProviderComponentName)
		proto.AddDependency("QueryManager", queryManagerName)

		return append(protoComponents, proto)
	}
}
