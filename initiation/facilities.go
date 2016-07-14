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
	protoComponents         map[string]*ioc.ProtoComponent
}

func NewFacilitiesInitialisor(pc map[string]*ioc.ProtoComponent, flm *logger.ComponentLoggerManager) *FacilitiesInitialisor {
	fi := new(FacilitiesInitialisor)
	fi.protoComponents = pc
	fi.FrameworkLoggingManager = flm

	return fi
}

func BootstrapFrameworkLogging(protoComponents map[string]*ioc.ProtoComponent, bootStrapLogLevel int) *logger.ComponentLoggerManager {

	flm := logger.CreateComponentLoggerManager(bootStrapLogLevel, nil)
	proto := ioc.CreateProtoComponent(flm, frameworkLoggingManagerName)

	protoComponents[frameworkLoggingManagerName] = proto

	return flm

}

func (fi *FacilitiesInitialisor) Initialise(ca *config.ConfigAccessor) {
	fi.ConfigAccessor = ca

	fc := new(FacilityConfig)

	ca.Populate("Facilities", fc)

	fi.updateFrameworkLogLevel()

	if fc.ApplicationLogging {
		fi.initialiseApplicationLogger()
	}

	if fc.HttpServer {
		fi.initialiseHttpServer()
	}

	if fc.QueryManager {
		fi.initialiseQueryManager()
	}

	if fc.RdbmsAccess {
		fi.initialiseDatabaseAccessor()
	}

	if fc.JsonWs {
		json.InitialiseJsonHttp(fi.FrameworkLoggingManager, fi.ConfigAccessor, fi.protoComponents)
	}

	if fc.ServiceErrorManager {
		serviceerror.InitialiseServiceErrorManager(fi.FrameworkLoggingManager, fi.ConfigAccessor, fi.protoComponents)
	}

}

func (fi *FacilitiesInitialisor) updateFrameworkLogLevel() {
	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("FrameworkLogger.DefaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	fi.FrameworkLoggingManager.UpdateGlobalThreshold(defaultLogLevel)
	fi.FrameworkLoggingManager.UpdateLocalThreshold(defaultLogLevel)
}

func (fi *FacilitiesInitialisor) initialiseApplicationLogger() {

	defaultLogLevelLabel := fi.ConfigAccessor.StringVal("ApplicationLogger.DefaultLogLevel")
	defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	initialLogLevelsByComponent := fi.ConfigAccessor.ObjectVal("ApplicationLogger.ComponentLogLevels")

	applicationLoggingManager := logger.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)
	applicationLoggingMangagerProto := ioc.CreateProtoComponent(applicationLoggingManager, applicationLoggingManagerName)
	fi.protoComponents[applicationLoggingManagerName] = applicationLoggingMangagerProto

	applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
	applicationLoggingDecorator.LoggerManager = applicationLoggingManager
	applicationLoggingDecorator.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(applicationLoggingDecoratorName)
	applicationLoggingDecoratorProto := ioc.CreateProtoComponent(applicationLoggingDecorator, applicationLoggingDecoratorName)

	fi.protoComponents[applicationLoggingDecoratorName] = applicationLoggingDecoratorProto

}

func (fi *FacilitiesInitialisor) initialiseHttpServer() {

	httpServer := new(httpserver.QuiltHttpServer)
	fi.ConfigAccessor.Populate("HttpServer", httpServer)

	httpServer.Logger = fi.FrameworkLoggingManager.CreateLogger(httpServerName)

	proto := ioc.CreateProtoComponent(httpServer, httpServerName)
	fi.protoComponents[httpServerName] = proto

	if !httpServer.AccessLogging {
		return
	}

	accessLogWriter := new(httpserver.AccessLogWriter)
	fi.ConfigAccessor.Populate("HttpServer.AccessLog", accessLogWriter)

	httpServer.AccessLogWriter = accessLogWriter

	proto = ioc.CreateProtoComponent(accessLogWriter, accessLogWriterName)
	fi.protoComponents[accessLogWriterName] = proto

}

func (fi *FacilitiesInitialisor) initialiseQueryManager() {
	queryManager := new(querymanager.QueryManager)
	queryManager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(queryManagerName)
	fi.ConfigAccessor.Populate("QueryManager", queryManager)

	proto := ioc.CreateProtoComponent(queryManager, queryManagerName)
	fi.protoComponents[queryManagerName] = proto
}

func (fi *FacilitiesInitialisor) initialiseDatabaseAccessor() {
	manager := new(rdbms.DefaultRdbmsClientManager)
	manager.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger(rdbmsClientManagerName)
	fi.ConfigAccessor.Populate("RdbmsAccess", manager)

	proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

	proto.AddDependency("Provider", manager.DatabaseProviderComponentName)
	proto.AddDependency("QueryManager", queryManagerName)

	fi.protoComponents[rdbmsClientManagerName] = proto
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
