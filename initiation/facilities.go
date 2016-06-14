package initiation
import (
    "github.com/wolferton/quilt/ioc"
    "github.com/wolferton/quilt/facility/logger"
    "github.com/wolferton/quilt/facility/httpserver"
    "github.com/wolferton/quilt/facility/decorator"
    "github.com/wolferton/quilt/config"
)

const applicationLoggingManagerComponentName = "quiltApplicationLoggingManager"
const frameworkLoggingManagerComponentName = "quiltFrameworkLoggingManager"
const applicationLoggingDecoratorName = "quiltApplicationLoggingDecorator"
const httpServerComponentName = "quiltHttpServer"

type FacilitiesInitialisor struct{

}

func (fi *FacilitiesInitialisor) BootstrapFrameworkLogging(protoComponents []*ioc.ProtoComponent, bootStrapLogLevel int) ([]*ioc.ProtoComponent, *logger.ComponentLoggerManager) {


    frameworkLoggingManager := logger.CreateComponentLoggerManager(bootStrapLogLevel, nil)
    frameworkLoggingManagerProto := ioc.CreateProtoComponent(frameworkLoggingManager, frameworkLoggingManagerComponentName)
    return append(protoComponents, frameworkLoggingManagerProto), frameworkLoggingManager

}

func (fi *FacilitiesInitialisor) UpdateFrameworkLogLevel(frameworkLoggingManager *logger.ComponentLoggerManager, configAccessor *config.ConfigAccessor) {
    defaultLogLevelLabel := configAccessor.StringVal("facilities.frameworkLogger.defaultLogLevel")
    defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	frameworkLoggingManager.UpdateGlobalThreshold(defaultLogLevel)
	frameworkLoggingManager.UpdateLocalThreshold(defaultLogLevel)
}

func (fi *FacilitiesInitialisor) InitialiseApplicationLogger(protoComponents []*ioc.ProtoComponent, configAccessor *config.ConfigAccessor, frameworkLoggingManager *logger.ComponentLoggerManager) []*ioc.ProtoComponent {

    defaultLogLevelLabel := configAccessor.StringVal("facilities.applicationLogger.defaultLogLevel")
    defaultLogLevel := logger.LogLevelFromLabel(defaultLogLevelLabel)

	initialLogLevelsByComponent := configAccessor.ObjectVal("facilities.applicationLogger.componentLogLevels")



    applicationLoggingManager := logger.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)

	applicationLoggingMangagerProto := ioc.CreateProtoComponent(applicationLoggingManager, applicationLoggingManagerComponentName)
    protoComponents = append(protoComponents, applicationLoggingMangagerProto)

    applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
    applicationLoggingDecorator.LoggerManager = applicationLoggingManager

    applicationLoggingDecorator.FrameworkLogger = frameworkLoggingManager.CreateLogger("ApplicationLogDecorator")

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

    httpServerComponent := new(ioc.Component)
    httpServerComponent.Instance = httpServer
    httpServerComponent.Name =  httpServerComponentName

    proto := new(ioc.ProtoComponent)
    proto.Component = httpServerComponent

    return append(protoComponents, proto)
}
