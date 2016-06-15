package decorator

import (
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"reflect"
)

const expectedApplicationLoggerFieldName string = "QuiltApplicationLogger"

func HasFieldOfName(component *ioc.Component, name string) bool {
	reflectComponent := reflect.ValueOf(component.Instance).Elem()
	reflectFieldOfInterest := reflectComponent.FieldByName(expectedApplicationLoggerFieldName)

	return reflectFieldOfInterest.IsValid()
}

func TypeOfField(component *ioc.Component, name string) reflect.Type {
	reflectComponent := reflect.ValueOf(component.Instance).Elem()
	return reflectComponent.FieldByName(expectedApplicationLoggerFieldName).Type()
}

type ApplicationLogDecorator struct {
	LoggerManager   *logger.ComponentLoggerManager
	FrameworkLogger logger.Logger
}

func (ald *ApplicationLogDecorator) OfInterest(component *ioc.Component) bool {

	result := HasFieldOfName(component, expectedApplicationLoggerFieldName)

	frameworkLog := ald.FrameworkLogger

	if frameworkLog.IsLevelEnabled(logger.Trace) {
		if result {
			frameworkLog.LogTrace(component.Name + " NEEDS an ApplicationLogger")

		} else {
			frameworkLog.LogTrace(component.Name + " does not need an ApplicationLogger")
		}
	}

	return result
}

func (ald *ApplicationLogDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	logger := ald.LoggerManager.CreateLogger(component.Name)

	targetFieldType := TypeOfField(component, expectedApplicationLoggerFieldName)
	typeOfLogger := reflect.TypeOf(logger)

	if typeOfLogger.AssignableTo(targetFieldType) {
		reflectComponent := reflect.ValueOf(component.Instance).Elem()
		reflectComponent.FieldByName(expectedApplicationLoggerFieldName).Set(reflect.ValueOf(logger))
	} else {
		ald.FrameworkLogger.LogError("Unable to inject an ApplicationLogger into component " + component.Name + " because field " + expectedApplicationLoggerFieldName + " is not of the expected type logger.Logger")
	}

}
