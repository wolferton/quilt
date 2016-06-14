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


func (fi *FacilitiesInitialisor) InitialiseLogging(protoComponents []*ioc.ProtoComponent) ([]*ioc.ProtoComponent, *logger.ComponentLoggerManager) {

    applicationLoggingManager := logger.CreateComponentLoggerManager(20)
    applicationLoggingMangagerProto := ioc.CreateProtoComponent(applicationLoggingManager, applicationLoggingManagerComponentName)
    protoComponents = append(protoComponents, applicationLoggingMangagerProto)

    frameworkLoggingManager := logger.CreateComponentLoggerManager(logger.Trace)
    frameworkLoggingManagerProto := ioc.CreateProtoComponent(frameworkLoggingManager, frameworkLoggingManagerComponentName)
    protoComponents = append(protoComponents, frameworkLoggingManagerProto)

    applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
    applicationLoggingDecorator.LoggerManager = applicationLoggingManager

    applicationLoggingDecorator.FrameworkLogger = frameworkLoggingManager.CreateLoggerAtLevel("ApplicationLogDecorator",logger.Trace)

    applicationLoggingDecoratorProto := ioc.CreateProtoComponent(applicationLoggingDecorator, applicationLoggingDecoratorName)

    return append(protoComponents, applicationLoggingDecoratorProto), frameworkLoggingManager


}

func (fi *FacilitiesInitialisor) InitialiseHttpServer(protoComponents []*ioc.ProtoComponent, configAccessor *config.ConfigAccessor, frameworkLoggingManager *logger.ComponentLoggerManager) []*ioc.ProtoComponent {

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
