package ioc

import (
	"errors"
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"os"
	"reflect"
	"time"
)

const containerDecoratorComponentName = "quiltContainerDecorator"
const containerComponentName = "quiltContainer"

type ComponentContainer struct {
	allComponents    map[string]*Component
	componentsByType map[string][]interface{}
	logger           logger.Logger
	configInjector   *config.ConfigInjector
	startable        []*Component
	stoppable        []*Component
}

func (cc *ComponentContainer) AllComponents() map[string]*Component {
	return cc.allComponents
}

func (cc *ComponentContainer) FindByType(typeName string) []interface{} {
	return cc.componentsByType[typeName]
}

func (cc *ComponentContainer) StartComponents() error {
	for _, component := range cc.startable {

		startable := component.Instance.(Startable)

		err := startable.StartComponent()

		if err != nil {
			message := fmt.Sprintf("Unable to start %s: %s", component.Name, err)
			return errors.New(message)
		}

	}

	return nil
}

func (cc *ComponentContainer) ShutdownComponents() error {

	for _, c := range cc.stoppable {

		s := c.Instance.(Stoppable)
		s.PrepareToStop()
	}

	cc.waitForReadyToStop(5*time.Second, 10, 3)

	for _, c := range cc.stoppable {

		s := c.Instance.(Stoppable)
		err := s.Stop()

		if err != nil {
			cc.logger.LogErrorf("%s did not stop cleanly %s", c.Name, err)
		}

	}

	return nil

}

func (cc *ComponentContainer) waitForReadyToStop(retestInterval time.Duration, maxTries int, warnAfterTries int) {

	for i := 0; i < maxTries; i++ {

		notReady := cc.countNotReady(i > warnAfterTries)

		if notReady == 0 {
			return
		} else {
			time.Sleep(retestInterval)
		}
	}

	cc.logger.LogFatal("Some components not ready to stop, stopping anyway")

}

func (cc *ComponentContainer) countNotReady(warn bool) int {

	notReady := 0

	for _, c := range cc.stoppable {
		s := c.Instance.(Stoppable)

		ready, err := s.ReadyToStop()

		if !ready {
			notReady += 1

			if warn {
				if err != nil {
					cc.logger.LogWarnf("%s is not ready to stop: %s", c.Name, err)
				} else {
					cc.logger.LogWarnf("%s is not ready to stop (no reason given)", c.Name)
				}

			}
		}

	}

	return notReady
}

func (cc *ComponentContainer) Populate(protoComponents []*ProtoComponent, configAccessor *config.ConfigAccessor) {

	decorators := make([]ComponentDecorator, 1)

	containerDecorator := new(ContainerDecorator)
	containerDecorator.container = cc

	decorators[0] = containerDecorator

	cc.allComponents = make(map[string]*Component)
	cc.componentsByType = make(map[string][]interface{})

	for index, protoComponent := range protoComponents {

		component := protoComponent.Component

		cc.addComponent(component, index)
		decorators = cc.captureDecorator(component, decorators)

	}

	err := cc.resolveDependenciesAndConfig(protoComponents, configAccessor)

	if err != nil {
		cc.logger.LogFatal(err.Error())
		cc.logger.LogInfo("Aborting startup")
		os.Exit(-1)
	}

	cc.decorateComponents(decorators)
}

func (cc *ComponentContainer) resolveDependenciesAndConfig(protoComponents []*ProtoComponent, configAccessor *config.ConfigAccessor) error {

	fl := cc.logger

	for _, proto := range protoComponents {

		for fieldName, depName := range proto.Dependencies {

			fl.LogTracef("%s needs %s", proto.Component.Name, depName)

			requiredComponent := cc.allComponents[depName]

			if requiredComponent == nil {
				message := fmt.Sprintf("No component named %s available (required by %s.%s)", depName, proto.Component.Name, fieldName)
				return errors.New(message)
			}

			requiredInstance := requiredComponent.Instance

			targetReflect := reflect.ValueOf(proto.Component.Instance).Elem()

			defer func() {
				if r := recover(); r != nil {
					fl.LogFatalf("Problem setting %s.%s: %s ", proto.Component.Name, fieldName, r)
				}
			}()

			targetReflect.FieldByName(fieldName).Set(reflect.ValueOf(requiredInstance))
		}

		for fieldName, configPath := range proto.ConfigPromises {
			fl.LogTracef("%s needs %s", proto.Component.Name, fieldName, configPath)

			cc.configInjector.PopulateFieldFromJsonPath(fieldName, configPath, proto.Component.Instance)

		}

	}

	return nil
}

func (cc *ComponentContainer) decorateComponents(decorators []ComponentDecorator) {

	for _, component := range cc.allComponents {
		for _, decorator := range decorators {

			if decorator.OfInterest(component) {
				decorator.DecorateComponent(component, cc)
			}
		}
	}

}

func (cc *ComponentContainer) captureDecorator(component *Component, decorators []ComponentDecorator) []ComponentDecorator {

	decorator, isDecorator := component.Instance.(ComponentDecorator)

	if isDecorator {
		cc.logger.LogTracef("Found decorator %s", component.Name)
		return append(decorators, decorator)
	} else {
		return decorators
	}
}

func (cc *ComponentContainer) addComponent(component *Component, index int) {
	cc.allComponents[component.Name] = component
	cc.mapComponentToType(component)

	l := cc.logger

	_, startable := component.Instance.(Startable)

	if startable {
		l.LogTracef("%s is Startable", component.Name)
		cc.startable = append(cc.startable, component)
	}

	_, stoppable := component.Instance.(Stoppable)

	if stoppable {
		l.LogTracef("%s is Stoppable", component.Name)
		cc.stoppable = append(cc.stoppable, component)
	}

}

func (cc *ComponentContainer) mapComponentToType(component *Component) {
	componentType := reflect.TypeOf(component.Instance)
	typeName := componentType.String()

	cc.logger.LogTracef("Storing component %s of type %s", component.Name, componentType.String())

	componentsOfSameType := cc.componentsByType[typeName]

	if componentsOfSameType == nil {
		componentsOfSameType = make([]interface{}, 1)
		componentsOfSameType[0] = component.Instance
		cc.componentsByType[typeName] = componentsOfSameType
	} else {
		cc.componentsByType[typeName] = append(componentsOfSameType, component.Instance)
	}

}

func CreateContainer(protoComponents []*ProtoComponent, loggingManager *logger.ComponentLoggerManager, configAccessor *config.ConfigAccessor, configInjector *config.ConfigInjector) *ComponentContainer {

	container := new(ComponentContainer)
	container.logger = loggingManager.CreateLogger(containerComponentName)
	container.configInjector = configInjector
	container.Populate(protoComponents, configAccessor)

	return container

}

type ContainerAccessor interface {
	Container(container *ComponentContainer)
}

type ContainerDecorator struct {
	container *ComponentContainer
}

func (cd *ContainerDecorator) OfInterest(component *Component) bool {
	result := false

	switch component.Instance.(type) {
	case ContainerAccessor:
		result = true
	}

	return result
}

func (cd *ContainerDecorator) DecorateComponent(component *Component, container *ComponentContainer) {

	accessor := component.Instance.(ContainerAccessor)
	accessor.Container(container)

}
