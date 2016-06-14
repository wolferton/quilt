package initiation
import (
    "github.com/wolferton/quilt/ioc"
    "github.com/wolferton/quilt/facility/logger"
    "github.com/wolferton/quilt/facility/httpserver"
    "github.com/wolferton/quilt/facility/decorator"
    "github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/querymanager"
)

const applicationLoggingManagerComponentName = "quiltApplicationLoggingManager"
const frameworkLoggingManagerComponentName = "quiltFrameworkLoggingManager"
const applicationLoggingDecoratorName = "quiltApplicationLoggingDecorator"
const httpServerComponentName = "quiltHttpServer"
const queryManagerComponentName = "quiltQueryManager"

type FacilitiesInitialisor struct{
	ConfigAccessor *config.ConfigAccessor
	ConfigInjector *config.ConfigInjector
	FrameworkLoggingManager *logger.ComponentLoggerManager
}

func (fi *FacilitiesInitialisor) BootstrapFrameworkLogging(protoComponents []*ioc.ProtoComponent, bootStrapLogLevel int) ([]*ioc.ProtoComponent, *logger.ComponentLoggerManager) {

	frameworkLoggingManager := logger.CreateComponentLoggerManager(bootStrapLogLevel, nil)
    frameworkLoggingManagerProto := ioc.CreateProtoComponent(frameworkLoggingManager, frameworkLoggingManagerComponentName)

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

	applicationLoggingMangagerProto := ioc.CreateProtoComponent(applicationLoggingManager, applicationLoggingManagerComponentName)
    protoComponents = append(protoComponents, applicationLoggingMangagerProto)

    applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
    applicationLoggingDecorator.LoggerManager = applicationLoggingManager

    applicationLoggingDecorator.FrameworkLogger = fi.FrameworkLoggingManager.CreateLogger("ApplicationLogDecorator")

    applicationLoggingDecoratorProto := ioc.CreateProtoComponent(applicationLoggingDecorator, applicationLoggingDecoratorName)

    return append(protoComponents, applicationLoggingDecoratorProto)
}


func (fi *FacilitiesInitialisor) InitialiseHttpServer(protoComponents []*ioc.ProtoComponent, configAccessor *config.ConfigAccessor, frameworkLoggingManager *logger.ComponentLoggerManager) []*ioc.ProtoComponent {

	if ! configAccessor.BoolValue("facilities.httpServer.enabled") {
		return protoComponents
	}

    httpServerConfig := httpserver.ParseDefaultHttpServerConfig(configAccessor)

    httpServer := new(httpserver.QuiltHttpServer)
    httpServer.Config = httpServerConfig
    httpServer.Logger = frameworkLoggingManager.CreateLogger(httpServerComponentName)

    proto := fi.wrapInProto(httpServer, httpServerComponentName)

    return append(protoComponents, proto)
}

func (fi *FacilitiesInitialisor) InitialiseQueryManager(protoComponents []*ioc.ProtoComponent) []*ioc.ProtoComponent{
	if ! fi.ConfigAccessor.BoolValue("facilities.queryManager.enabled") {
		return protoComponents
	} else {

		queryManager := new(querymanager.QueryManager)
		fi.ConfigInjector.PopulateObjectFromJsonPath("facilities.queryManager", queryManager)

		proto := fi.wrapInProto(queryManager, queryManagerComponentName)

		return append(protoComponents, proto)
	}
}

func (fi *FacilitiesInitialisor) wrapInProto(instance interface{}, name string) *ioc.ProtoComponent{

	component := new(ioc.Component)
	component.Instance = instance
	component.Name =  name

	proto := new(ioc.ProtoComponent)
	proto.Component = component

	return proto
}