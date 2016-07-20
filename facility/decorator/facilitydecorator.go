package decorator

import (
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"reflect"
)

//TODO Rename application log var
const expectedApplicationLoggerFieldName string = "QuiltApplicationLogger"
const expectedFrameworkLoggerFieldName string = "FrameworkLogger"

func HasFieldOfName(component *ioc.Component, fieldName string) bool {
	reflectComponent := reflect.ValueOf(component.Instance).Elem()
	reflectFieldOfInterest := reflectComponent.FieldByName(fieldName)

	return reflectFieldOfInterest.IsValid()
}

func TypeOfField(component *ioc.Component, name string) reflect.Type {
	reflectComponent := reflect.ValueOf(component.Instance).Elem()
	return reflectComponent.FieldByName(name).Type()
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
			frameworkLog.LogTracef("%s NEEDS an ApplicationLogger", component.Name)

		} else {
			frameworkLog.LogTracef("%s does not need an ApplicationLogger (no field named %s)", component.Name, expectedApplicationLoggerFieldName)
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
		ald.FrameworkLogger.LogErrorf("Unable to inject an ApplicationLogger into component %s because field %s is not of the expected type logger.Logger", component.Name, expectedApplicationLoggerFieldName)
	}

}

type FrameworkLogDecorator struct {
	LoggerManager   *logger.ComponentLoggerManager
	FrameworkLogger logger.Logger
}

func (fld *FrameworkLogDecorator) OfInterest(component *ioc.Component) bool {

	result := HasFieldOfName(component, expectedFrameworkLoggerFieldName)

	frameworkLog := fld.FrameworkLogger

	if frameworkLog.IsLevelEnabled(logger.Trace) {
		if result {
			frameworkLog.LogTracef("%s NEEDS a %s", component.Name, expectedFrameworkLoggerFieldName)

		} else {
			frameworkLog.LogTracef("%s does not need a %s", component.Name, expectedFrameworkLoggerFieldName)
		}
	}

	return result
}

func (fld *FrameworkLogDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	logger := fld.LoggerManager.CreateLogger(component.Name)

	targetFieldType := TypeOfField(component, expectedFrameworkLoggerFieldName)
	typeOfLogger := reflect.TypeOf(logger)

	if typeOfLogger.AssignableTo(targetFieldType) {
		reflectComponent := reflect.ValueOf(component.Instance).Elem()
		reflectComponent.FieldByName(expectedFrameworkLoggerFieldName).Set(reflect.ValueOf(logger))
	} else {
		fld.FrameworkLogger.LogErrorf("Unable to inject a FrameworkLogger into component %s because field %s is not of the expected type logger.Logger", component.Name, expectedFrameworkLoggerFieldName)
	}

}
