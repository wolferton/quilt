package decorator
import (
    "github.com/wolferton/quilt/ioc"
    "github.com/wolferton/quilt/facility/logger"
)

type ApplicationLogDecorator struct {
    LoggerManager *logger.ComponentLoggerManager
}

func (ald *ApplicationLogDecorator) OfInterest(component *ioc.Component) bool{

    result := false

    switch component.Instance.(type) {
        case logger.ApplicationLogSource:
        result = true
    }

    return result
}

func (ald *ApplicationLogDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer){
    logSource := component.Instance.(logger.ApplicationLogSource)
    logger := ald.LoggerManager.CreateLogger(component.Name)
    logSource.SetApplicationLogger(logger)
}