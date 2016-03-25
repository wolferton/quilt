package ioc
import (
    "reflect"
    "github.com/wolferton/quilt/facility/logger"
)

const containerDecoratorComponentName = "quiltContainerDecorator"
const containerComponentName = "quiltContainer"

type ComponentContainer struct {
    allComponents map[string]*Component
    componentsByType map[string][]interface{}
    logger logger.Logger
}

func (cc *ComponentContainer) FindByType(typeName string) []interface{} {
    return cc.componentsByType[typeName]
}

func (cc *ComponentContainer) StartComponents() {
    for _, component := range cc.allComponents {

        startable, isStartable := component.Instance.(Startable)

        if(isStartable) {
            startable.StartComponent()
        }

    }
}

func (cc *ComponentContainer) Populate(protoComponents []*ProtoComponent) {

    decorators := make([]ComponentDecorator,1)

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

    cc.resolveDependencies(protoComponents)
    cc.decorateComponents(decorators)
}

func (cc *ComponentContainer) resolveDependencies(protoComponents []*ProtoComponent) {

    for _, proto := range protoComponents {

        for fieldName, depName := range proto.Dependencies {

            cc.logger.LogDebug(fieldName + " needs " + depName)

            requiredComponent := cc.allComponents[depName]
            requiredInstance := requiredComponent.Instance

            targetReflect := reflect.ValueOf(proto.Component.Instance).Elem()
            targetReflect.FieldByName(fieldName).Set(reflect.ValueOf(requiredInstance))
        }

    }

}

func (cc *ComponentContainer) decorateComponents(decorators []ComponentDecorator){

    for _, component := range cc.allComponents {
        for _, decorator := range decorators {

            if(decorator.OfInterest(component)){
                decorator.DecorateComponent(component, cc)
            }
        }
    }

}



func (cc *ComponentContainer) captureDecorator(component *Component, decorators []ComponentDecorator) []ComponentDecorator{

    decorator, isDecorator := component.Instance.(ComponentDecorator)

    if(isDecorator) {
        cc.logger.LogTrace("Found decorator " + component.Name)
        return append(decorators, decorator)
    } else {
        return decorators
    }
}

func (cc *ComponentContainer) addComponent(component *Component, index int) {
    cc.allComponents[component.Name] = component
    cc.mapComponentToType(component)
}

func (cc *ComponentContainer) mapComponentToType(component *Component) {
    componentType := reflect.TypeOf(component.Instance)
    typeName := componentType.String()

    cc.logger.LogDebug("Storing component" + component.Name + " of type " + componentType.String())

    componentsOfSameType := cc.componentsByType[typeName]

    if(componentsOfSameType == nil) {
        componentsOfSameType = make([]interface{},1)
        componentsOfSameType[0] = component.Instance
        cc.componentsByType[typeName] = componentsOfSameType
    } else {
        cc.componentsByType[typeName] = append(componentsOfSameType,component.Instance)
    }

}


func CreateContainer(protoComponents []*ProtoComponent, loggingManager *logger.ComponentLoggerManager) *ComponentContainer {

    container := new(ComponentContainer)
    container.logger = loggingManager.CreateLogger(containerComponentName)

    container.Populate(protoComponents)

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